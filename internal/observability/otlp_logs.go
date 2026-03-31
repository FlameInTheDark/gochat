package observability

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	collectorpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

const (
	defaultOTLPLogsQueueSize      = 2048
	defaultOTLPLogsBatchSize      = 128
	defaultOTLPLogsFlushInterval  = 2 * time.Second
	defaultOTLPLogsRequestTimeout = 5 * time.Second
	defaultOTLPLogsRetryBackoff   = 250 * time.Millisecond
	defaultOTLPLogsMaxAttempts    = 3
)

type otlpLogExporterConfig struct {
	headers        map[string]string
	endpoint       string
	scopeName      string
	resourceAttrs  []attribute.KeyValue
	flushInterval  time.Duration
	requestTimeout time.Duration
	retryBackoff   time.Duration
	queueSize      int
	batchSize      int
	maxAttempts    int
}

type otlpLogExporter struct {
	config       otlpLogExporterConfig
	registration metric.Registration
	resource     *resourcepb.Resource

	enqueueCounter metric.Int64Counter
	successCounter metric.Int64Counter
	failureCounter metric.Int64Counter
	droppedCounter metric.Int64Counter
	batchLatency   metric.Float64Histogram
	client         *http.Client
	queue          chan []byte
	done           chan struct{}
	cancel         context.CancelFunc
	enqueuedTotal  atomic.Int64
	successTotal   atomic.Int64
	failureTotal   atomic.Int64
	droppedTotal   atomic.Int64
	lastSuccess    atomic.Int64
	shutdownOnce   sync.Once
}

type otlpLogHandler struct {
	exporter *otlpLogExporter
	decorate func(slog.Handler) slog.Handler
}

func loadOTLPLogExporterConfigFromEnv(serviceName string, resourceAttrs []attribute.KeyValue) (otlpLogExporterConfig, bool, error) {
	exporterMode := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_LOGS_EXPORTER")))
	if exporterMode == "none" {
		return otlpLogExporterConfig{}, false, nil
	}

	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"))
	headersRaw := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_LOGS_HEADERS"))
	protocol := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"))

	allowBaseFallback := serviceName == "gochat-sfu" || exporterMode == "otlp"
	if allowBaseFallback {
		if endpoint == "" {
			endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
		}
		if headersRaw == "" {
			headersRaw = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))
		}
		if protocol == "" {
			protocol = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
		}
	}
	if endpoint == "" {
		return otlpLogExporterConfig{}, false, nil
	}

	if protocol == "" {
		protocol = "http/protobuf"
	}
	if !strings.EqualFold(protocol, "http/protobuf") {
		return otlpLogExporterConfig{}, false, fmt.Errorf("unsupported OTLP logs protocol %q", protocol)
	}

	headers, err := parseOTLPHeaderEnv(headersRaw)
	if err != nil {
		return otlpLogExporterConfig{}, false, err
	}

	return otlpLogExporterConfig{
		endpoint:       normalizeOTLPSignalEndpoint(endpoint, "logs"),
		headers:        headers,
		resourceAttrs:  append([]attribute.KeyValue(nil), resourceAttrs...),
		scopeName:      normalizeScope(serviceName, "gochat/logs"),
		queueSize:      defaultOTLPLogsQueueSize,
		batchSize:      defaultOTLPLogsBatchSize,
		flushInterval:  defaultOTLPLogsFlushInterval,
		requestTimeout: defaultOTLPLogsRequestTimeout,
		retryBackoff:   defaultOTLPLogsRetryBackoff,
		maxAttempts:    defaultOTLPLogsMaxAttempts,
	}, true, nil
}

func newOTLPLogExporter(serviceName string, cfg otlpLogExporterConfig) (*otlpLogExporter, error) {
	if cfg.queueSize <= 0 {
		cfg.queueSize = defaultOTLPLogsQueueSize
	}
	if cfg.batchSize <= 0 {
		cfg.batchSize = defaultOTLPLogsBatchSize
	}
	if cfg.flushInterval <= 0 {
		cfg.flushInterval = defaultOTLPLogsFlushInterval
	}
	if cfg.requestTimeout <= 0 {
		cfg.requestTimeout = defaultOTLPLogsRequestTimeout
	}
	if cfg.retryBackoff <= 0 {
		cfg.retryBackoff = defaultOTLPLogsRetryBackoff
	}
	if cfg.maxAttempts <= 0 {
		cfg.maxAttempts = defaultOTLPLogsMaxAttempts
	}
	if cfg.scopeName == "" {
		cfg.scopeName = normalizeScope(serviceName, "gochat/logs")
	}

	ctx, cancel := context.WithCancel(context.Background())
	exporter := &otlpLogExporter{
		client: &http.Client{
			Timeout: cfg.requestTimeout,
		},
		config: cfg,
		queue:  make(chan []byte, cfg.queueSize),
		done:   make(chan struct{}),
		cancel: cancel,
		resource: &resourcepb.Resource{
			Attributes: otlpAttributesFromResource(cfg.resourceAttrs),
		},
	}

	meter := otel.Meter(normalizeScope(serviceName, "gochat/logs-exporter"))
	exporter.enqueueCounter, _ = meter.Int64Counter("gochat.logs.exporter.enqueued")
	exporter.successCounter, _ = meter.Int64Counter("gochat.logs.exporter.send_success")
	exporter.failureCounter, _ = meter.Int64Counter("gochat.logs.exporter.send_failure")
	exporter.droppedCounter, _ = meter.Int64Counter("gochat.logs.exporter.dropped")
	exporter.batchLatency, _ = meter.Float64Histogram("gochat.logs.exporter.batch.duration")
	lastSuccessGauge, err := meter.Int64ObservableGauge("gochat.logs.exporter.last_success_unix")
	if err == nil {
		exporter.registration, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
			observer.ObserveInt64(lastSuccessGauge, exporter.lastSuccess.Load())
			return nil
		}, lastSuccessGauge)
	}
	if err != nil {
		cancel()
		return nil, fmt.Errorf("register OTLP log exporter metrics: %w", err)
	}

	go exporter.run(ctx)
	return exporter, nil
}

func newOTLPLogHandler(exporter *otlpLogExporter) slog.Handler {
	return &otlpLogHandler{
		exporter: exporter,
		decorate: func(handler slog.Handler) slog.Handler { return handler },
	}
}

func (h *otlpLogHandler) Enabled(context.Context, slog.Level) bool {
	return h != nil && h.exporter != nil
}

func (h *otlpLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if h == nil || h.exporter == nil {
		return nil
	}

	var buf bytes.Buffer
	var handler slog.Handler = slog.NewJSONHandler(&buf, nil)
	if h.decorate != nil {
		handler = h.decorate(handler)
	}
	if err := handler.Handle(ctx, record.Clone()); err != nil {
		return err
	}
	h.exporter.enqueue(ctx, buf.Bytes())
	return nil
}

func (h *otlpLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	prev := h.decorate
	if prev == nil {
		prev = func(handler slog.Handler) slog.Handler { return handler }
	}
	return &otlpLogHandler{
		exporter: h.exporter,
		decorate: func(handler slog.Handler) slog.Handler {
			return prev(handler).WithAttrs(attrs)
		},
	}
}

func (h *otlpLogHandler) WithGroup(name string) slog.Handler {
	prev := h.decorate
	if prev == nil {
		prev = func(handler slog.Handler) slog.Handler { return handler }
	}
	return &otlpLogHandler{
		exporter: h.exporter,
		decorate: func(handler slog.Handler) slog.Handler {
			return prev(handler).WithGroup(name)
		},
	}
}

func (e *otlpLogExporter) Close() error {
	if e == nil {
		return nil
	}

	var closeErr error
	e.shutdownOnce.Do(func() {
		e.cancel()
		<-e.done
		if e.registration != nil {
			closeErr = e.registration.Unregister()
		}
	})
	if closeErr != nil {
		Logger().Warn("unable to unregister OTLP log exporter metrics", "error", closeErr.Error())
	}
	return closeErr
}

func (e *otlpLogExporter) run(ctx context.Context) {
	defer close(e.done)

	ticker := time.NewTicker(e.config.flushInterval)
	defer ticker.Stop()

	batch := make([][]byte, 0, e.config.batchSize)
	flush := func(flushCtx context.Context) {
		if len(batch) == 0 {
			return
		}
		_ = e.flushBatch(flushCtx, batch)
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			for {
				select {
				case entry := <-e.queue:
					batch = append(batch, entry)
				default:
					flush(BackgroundFromContext(ctx))
					return
				}
			}
		case entry := <-e.queue:
			batch = append(batch, entry)
			if len(batch) >= e.config.batchSize {
				flush(ctx)
			}
		case <-ticker.C:
			flush(ctx)
		}
	}
}

func (e *otlpLogExporter) enqueue(ctx context.Context, line []byte) {
	if e == nil {
		return
	}
	entry := bytes.TrimSpace(line)
	if len(entry) == 0 {
		return
	}

	select {
	case e.queue <- bytes.Clone(entry):
		e.recordEnqueued(ctx, 1)
	default:
		e.recordDropped(ctx, 1)
	}
}

func (e *otlpLogExporter) flushBatch(ctx context.Context, batch [][]byte) error {
	payload, recordCount, err := e.buildRequest(batch)
	if err != nil {
		e.recordFailure(ctx, int64(len(batch)))
		return err
	}
	if recordCount == 0 {
		return nil
	}

	body, err := proto.Marshal(payload)
	if err != nil {
		e.recordFailure(ctx, int64(recordCount))
		return err
	}

	started := time.Now()
	var lastErr error
	for attempt := 1; attempt <= e.config.maxAttempts; attempt++ {
		err = e.sendBatch(ctx, body)
		if err == nil {
			recordCtx := BackgroundFromContext(ctx)
			e.recordSuccess(recordCtx, int64(recordCount))
			e.batchLatency.Record(recordCtx, time.Since(started).Seconds())
			e.lastSuccess.Store(time.Now().Unix())
			return nil
		}
		lastErr = err
		if !shouldRetryOTLPSend(err) || attempt == e.config.maxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			lastErr = ctx.Err()
			e.recordFailure(ctx, int64(recordCount))
			return lastErr
		case <-time.After(time.Duration(attempt) * e.config.retryBackoff):
		}
	}

	e.recordFailure(ctx, int64(recordCount))
	return lastErr
}

func (e *otlpLogExporter) buildRequest(batch [][]byte) (*collectorpb.ExportLogsServiceRequest, int, error) {
	records := make([]*logspb.LogRecord, 0, len(batch))
	for _, entry := range batch {
		record, err := otlpLogRecordFromJSON(entry)
		if err != nil {
			return nil, 0, err
		}
		if record != nil {
			records = append(records, record)
		}
	}
	if len(records) == 0 {
		return nil, 0, nil
	}

	return &collectorpb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: e.resource,
				ScopeLogs: []*logspb.ScopeLogs{
					{
						Scope:      &commonpb.InstrumentationScope{Name: e.config.scopeName},
						LogRecords: records,
					},
				},
			},
		},
	}, len(records), nil
}

func (e *otlpLogExporter) sendBatch(ctx context.Context, body []byte) error {
	requestCtx, cancel := context.WithTimeout(ctx, e.config.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, e.config.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	for key, value := range e.config.headers {
		req.Header.Set(key, value)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return otlpSendError{
			statusCode: resp.StatusCode,
			message:    strings.TrimSpace(string(respBody)),
		}
	}
	return nil
}

func (e *otlpLogExporter) recordEnqueued(ctx context.Context, delta int64) {
	e.enqueuedTotal.Add(delta)
	if e.enqueueCounter != nil {
		e.enqueueCounter.Add(BackgroundFromContext(ctx), delta)
	}
}

func (e *otlpLogExporter) recordSuccess(ctx context.Context, delta int64) {
	e.successTotal.Add(delta)
	if e.successCounter != nil {
		e.successCounter.Add(BackgroundFromContext(ctx), delta)
	}
}

func (e *otlpLogExporter) recordFailure(ctx context.Context, delta int64) {
	e.failureTotal.Add(delta)
	if e.failureCounter != nil {
		e.failureCounter.Add(BackgroundFromContext(ctx), delta)
	}
}

func (e *otlpLogExporter) recordDropped(ctx context.Context, delta int64) {
	e.droppedTotal.Add(delta)
	if e.droppedCounter != nil {
		e.droppedCounter.Add(BackgroundFromContext(ctx), delta)
	}
}

type otlpSendError struct {
	message    string
	statusCode int
}

func (e otlpSendError) Error() string {
	if e.message == "" {
		return fmt.Sprintf("OTLP logs endpoint returned HTTP %d", e.statusCode)
	}
	return fmt.Sprintf("OTLP logs endpoint returned HTTP %d: %s", e.statusCode, e.message)
}

func otlpLogRecordFromJSON(raw []byte) (*logspb.LogRecord, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		now := time.Now()
		return &logspb.LogRecord{
			TimeUnixNano:         uint64(now.UnixNano()),
			ObservedTimeUnixNano: uint64(now.UnixNano()),
			SeverityNumber:       logspb.SeverityNumber_SEVERITY_NUMBER_INFO,
			SeverityText:         "INFO",
			Body:                 &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: string(trimmed)}},
		}, nil
	}

	now := time.Now()
	recordTime := now
	if value, ok := payload["time"]; ok {
		if parsed, err := time.Parse(time.RFC3339Nano, fmt.Sprint(value)); err == nil {
			recordTime = parsed
		}
	}

	severityText := strings.ToUpper(strings.TrimSpace(fmt.Sprint(payload["level"])))
	if severityText == "" {
		severityText = "INFO"
	}

	bodyText := strings.TrimSpace(fmt.Sprint(payload["msg"]))
	if bodyText == "" {
		bodyText = string(trimmed)
	}

	traceID := parseHexBytes(payload["trace_id"], 16)
	spanID := parseHexBytes(payload["span_id"], 8)

	return &logspb.LogRecord{
		TimeUnixNano:         uint64(recordTime.UnixNano()),
		ObservedTimeUnixNano: uint64(now.UnixNano()),
		SeverityNumber:       severityNumberFromText(severityText),
		SeverityText:         severityText,
		Body:                 &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: bodyText}},
		Attributes:           otlpAttributesFromMap(payload),
		TraceId:              traceID,
		SpanId:               spanID,
	}, nil
}

func parseHexBytes(value any, size int) []byte {
	raw := strings.TrimSpace(fmt.Sprint(value))
	if raw == "" {
		return nil
	}
	decoded, err := hex.DecodeString(raw)
	if err != nil || len(decoded) != size {
		return nil
	}
	return decoded
}

func severityNumberFromText(level string) logspb.SeverityNumber {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return logspb.SeverityNumber_SEVERITY_NUMBER_DEBUG
	case "WARN", "WARNING":
		return logspb.SeverityNumber_SEVERITY_NUMBER_WARN
	case "ERROR":
		return logspb.SeverityNumber_SEVERITY_NUMBER_ERROR
	default:
		return logspb.SeverityNumber_SEVERITY_NUMBER_INFO
	}
}

func parseOTLPHeaderEnv(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	headers := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid OTLP header %q", part)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid OTLP header %q", part)
		}
		if decoded, err := url.QueryUnescape(value); err == nil {
			value = decoded
		}
		headers[key] = value
	}
	return headers, nil
}

func normalizeOTLPSignalEndpoint(raw, signal string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasSuffix(strings.TrimRight(raw, "/"), "/v1/"+signal) {
		return raw
	}
	return strings.TrimRight(raw, "/") + "/v1/" + signal
}

func shouldRetryOTLPSend(err error) bool {
	if err == nil {
		return false
	}
	var sendErr otlpSendError
	if errors.As(err, &sendErr) {
		return sendErr.statusCode == http.StatusTooManyRequests || sendErr.statusCode >= http.StatusInternalServerError
	}
	return true
}

func otlpAttributesFromResource(attrs []attribute.KeyValue) []*commonpb.KeyValue {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]*commonpb.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		key := strings.TrimSpace(string(attr.Key))
		if key == "" {
			continue
		}
		out = append(out, &commonpb.KeyValue{
			Key:   key,
			Value: otlpAnyValueFromInterface(attr.Value.AsInterface()),
		})
	}
	return out
}

func otlpAttributesFromMap(payload map[string]any) []*commonpb.KeyValue {
	if len(payload) == 0 {
		return nil
	}
	out := make([]*commonpb.KeyValue, 0, len(payload))
	for key, value := range payload {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out = append(out, &commonpb.KeyValue{
			Key:   key,
			Value: otlpAnyValueFromInterface(value),
		})
	}
	return out
}

func otlpAnyValueFromInterface(value any) *commonpb.AnyValue {
	switch typed := value.(type) {
	case nil:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: ""}}
	case string:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: typed}}
	case bool:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_BoolValue{BoolValue: typed}}
	case int:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case int8:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case int16:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case int32:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case int64:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: typed}}
	case uint:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case uint8:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case uint16:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case uint32:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case uint64:
		if typed > uint64(^uint64(0)>>1) {
			return &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: fmt.Sprint(typed)}}
		}
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: int64(typed)}}
	case float32:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_DoubleValue{DoubleValue: float64(typed)}}
	case float64:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_DoubleValue{DoubleValue: typed}}
	case []any:
		values := make([]*commonpb.AnyValue, 0, len(typed))
		for _, item := range typed {
			values = append(values, otlpAnyValueFromInterface(item))
		}
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_ArrayValue{ArrayValue: &commonpb.ArrayValue{Values: values}}}
	case map[string]any:
		values := make([]*commonpb.KeyValue, 0, len(typed))
		for key, item := range typed {
			values = append(values, &commonpb.KeyValue{
				Key:   key,
				Value: otlpAnyValueFromInterface(item),
			})
		}
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_KvlistValue{KvlistValue: &commonpb.KeyValueList{Values: values}}}
	default:
		return &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: fmt.Sprint(value)}}
	}
}
