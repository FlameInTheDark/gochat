package observability

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	defaultOpenObserveLogsStream         = "gochat_logs"
	defaultOpenObserveLogsQueueSize      = 2048
	defaultOpenObserveLogsBatchSize      = 128
	defaultOpenObserveLogsFlushInterval  = 2 * time.Second
	defaultOpenObserveLogsRequestTimeout = 5 * time.Second
	defaultOpenObserveLogsRetryBackoff   = 250 * time.Millisecond
	defaultOpenObserveLogsMaxAttempts    = 3
)

type openObserveLogExporterConfig struct {
	endpoint       string
	authHeader     string
	stream         string
	queueSize      int
	batchSize      int
	flushInterval  time.Duration
	requestTimeout time.Duration
	retryBackoff   time.Duration
	maxAttempts    int
}

type openObserveLogExporter struct {
	client        *http.Client
	config        openObserveLogExporterConfig
	queue         chan []byte
	done          chan struct{}
	cancel        context.CancelFunc
	registration  metric.Registration
	shutdownOnce  sync.Once
	enqueuedTotal atomic.Int64
	successTotal  atomic.Int64
	failureTotal  atomic.Int64
	droppedTotal  atomic.Int64
	lastSuccess   atomic.Int64

	enqueueCounter metric.Int64Counter
	successCounter metric.Int64Counter
	failureCounter metric.Int64Counter
	droppedCounter metric.Int64Counter
	batchLatency   metric.Float64Histogram
}

func loadOpenObserveLogExporterConfigFromEnv() (openObserveLogExporterConfig, bool, error) {
	enabledRaw := strings.TrimSpace(os.Getenv("OPENOBSERVE_LOGS_ENABLED"))
	if enabledRaw == "" {
		return openObserveLogExporterConfig{}, false, nil
	}
	enabled, err := strconv.ParseBool(enabledRaw)
	if err != nil {
		return openObserveLogExporterConfig{}, false, fmt.Errorf("parse OPENOBSERVE_LOGS_ENABLED: %w", err)
	}
	if !enabled {
		return openObserveLogExporterConfig{}, false, nil
	}

	cfg := openObserveLogExporterConfig{
		endpoint:       strings.TrimSpace(os.Getenv("OPENOBSERVE_LOGS_ENDPOINT")),
		authHeader:     strings.TrimSpace(os.Getenv("OPENOBSERVE_LOGS_AUTH")),
		stream:         defaultIfEmpty(strings.TrimSpace(os.Getenv("OPENOBSERVE_LOGS_STREAM")), defaultOpenObserveLogsStream),
		queueSize:      defaultOpenObserveLogsQueueSize,
		batchSize:      defaultOpenObserveLogsBatchSize,
		flushInterval:  defaultOpenObserveLogsFlushInterval,
		requestTimeout: defaultOpenObserveLogsRequestTimeout,
		retryBackoff:   defaultOpenObserveLogsRetryBackoff,
		maxAttempts:    defaultOpenObserveLogsMaxAttempts,
	}
	if cfg.endpoint == "" {
		return openObserveLogExporterConfig{}, false, errors.New("OPENOBSERVE_LOGS_ENDPOINT is required when OPENOBSERVE_LOGS_ENABLED=true")
	}
	if cfg.authHeader == "" {
		return openObserveLogExporterConfig{}, false, errors.New("OPENOBSERVE_LOGS_AUTH is required when OPENOBSERVE_LOGS_ENABLED=true")
	}
	return cfg, true, nil
}

func newOpenObserveLogExporter(serviceName string, cfg openObserveLogExporterConfig) (*openObserveLogExporter, error) {
	if cfg.queueSize <= 0 {
		cfg.queueSize = defaultOpenObserveLogsQueueSize
	}
	if cfg.batchSize <= 0 {
		cfg.batchSize = defaultOpenObserveLogsBatchSize
	}
	if cfg.flushInterval <= 0 {
		cfg.flushInterval = defaultOpenObserveLogsFlushInterval
	}
	if cfg.requestTimeout <= 0 {
		cfg.requestTimeout = defaultOpenObserveLogsRequestTimeout
	}
	if cfg.retryBackoff <= 0 {
		cfg.retryBackoff = defaultOpenObserveLogsRetryBackoff
	}
	if cfg.maxAttempts <= 0 {
		cfg.maxAttempts = defaultOpenObserveLogsMaxAttempts
	}
	if cfg.stream == "" {
		cfg.stream = defaultOpenObserveLogsStream
	}

	ctx, cancel := context.WithCancel(context.Background())
	exporter := &openObserveLogExporter{
		client: &http.Client{
			Timeout: cfg.requestTimeout,
		},
		config: cfg,
		queue:  make(chan []byte, cfg.queueSize),
		done:   make(chan struct{}),
		cancel: cancel,
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
		return nil, fmt.Errorf("register OpenObserve log exporter metrics: %w", err)
	}

	go exporter.run(ctx)
	return exporter, nil
}

func (e *openObserveLogExporter) Writer() io.Writer {
	return &jsonLineWriter{exporter: e}
}

func (e *openObserveLogExporter) Close() error {
	if e == nil {
		return nil
	}

	e.shutdownOnce.Do(func() {
		e.cancel()
		<-e.done
		if e.registration != nil {
			e.registration.Unregister()
		}
	})
	return nil
}

func (e *openObserveLogExporter) run(ctx context.Context) {
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
					flush(context.Background())
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

func (e *openObserveLogExporter) enqueue(line []byte) {
	if e == nil {
		return
	}
	entry := bytes.TrimSpace(line)
	if len(entry) == 0 {
		return
	}

	select {
	case e.queue <- bytes.Clone(entry):
		e.recordEnqueued(1)
	default:
		e.recordDropped(1)
	}
}

func (e *openObserveLogExporter) flushBatch(ctx context.Context, batch [][]byte) error {
	body := buildJSONBatchBody(batch)
	if len(body) == 0 {
		return nil
	}

	started := time.Now()
	var lastErr error
	for attempt := 1; attempt <= e.config.maxAttempts; attempt++ {
		err := e.sendBatch(ctx, body)
		if err == nil {
			e.recordSuccess(int64(len(batch)))
			e.batchLatency.Record(context.Background(), time.Since(started).Seconds())
			e.lastSuccess.Store(time.Now().Unix())
			return nil
		}
		lastErr = err
		if !shouldRetryOpenObserveSend(err) || attempt == e.config.maxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			lastErr = ctx.Err()
			e.recordFailure(int64(len(batch)))
			return lastErr
		case <-time.After(time.Duration(attempt) * e.config.retryBackoff):
		}
	}

	e.recordFailure(int64(len(batch)))
	return lastErr
}

func (e *openObserveLogExporter) sendBatch(ctx context.Context, body []byte) error {
	requestCtx, cancel := context.WithTimeout(ctx, e.config.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, openObserveLogsIngestURL(e.config.endpoint, e.config.stream), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", e.config.authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return openObserveSendError{
			statusCode: resp.StatusCode,
			message:    strings.TrimSpace(string(respBody)),
		}
	}
	return nil
}

func (e *openObserveLogExporter) recordEnqueued(delta int64) {
	e.enqueuedTotal.Add(delta)
	if e.enqueueCounter != nil {
		e.enqueueCounter.Add(context.Background(), delta)
	}
}

func (e *openObserveLogExporter) recordSuccess(delta int64) {
	e.successTotal.Add(delta)
	if e.successCounter != nil {
		e.successCounter.Add(context.Background(), delta)
	}
}

func (e *openObserveLogExporter) recordFailure(delta int64) {
	e.failureTotal.Add(delta)
	if e.failureCounter != nil {
		e.failureCounter.Add(context.Background(), delta)
	}
}

func (e *openObserveLogExporter) recordDropped(delta int64) {
	e.droppedTotal.Add(delta)
	if e.droppedCounter != nil {
		e.droppedCounter.Add(context.Background(), delta)
	}
}

type jsonLineWriter struct {
	exporter *openObserveLogExporter
	mu       sync.Mutex
	pending  []byte
}

func (w *jsonLineWriter) Write(p []byte) (int, error) {
	if w == nil || w.exporter == nil || len(p) == 0 {
		return len(p), nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.pending = append(w.pending, p...)
	for {
		idx := bytes.IndexByte(w.pending, '\n')
		if idx < 0 {
			break
		}
		line := bytes.TrimSpace(w.pending[:idx])
		if len(line) > 0 {
			w.exporter.enqueue(line)
		}
		w.pending = append([]byte(nil), w.pending[idx+1:]...)
	}

	return len(p), nil
}

type openObserveSendError struct {
	statusCode int
	message    string
}

func (e openObserveSendError) Error() string {
	if e.message == "" {
		return fmt.Sprintf("OpenObserve returned HTTP %d", e.statusCode)
	}
	return fmt.Sprintf("OpenObserve returned HTTP %d: %s", e.statusCode, e.message)
}

func buildJSONBatchBody(entries [][]byte) []byte {
	nonEmpty := 0
	size := 2
	for _, entry := range entries {
		trimmed := bytes.TrimSpace(entry)
		if len(trimmed) == 0 {
			continue
		}
		nonEmpty++
		size += len(trimmed)
	}
	if nonEmpty == 0 {
		return nil
	}
	if nonEmpty > 1 {
		size += nonEmpty - 1
	}

	body := make([]byte, 0, size)
	body = append(body, '[')
	first := true
	for _, entry := range entries {
		trimmed := bytes.TrimSpace(entry)
		if len(trimmed) == 0 {
			continue
		}
		if !first {
			body = append(body, ',')
		}
		body = append(body, trimmed...)
		first = false
	}
	body = append(body, ']')
	return body
}

func openObserveLogsIngestURL(endpoint, stream string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if strings.HasSuffix(endpoint, "/_json") {
		return endpoint
	}
	if strings.HasSuffix(endpoint, "/"+stream) {
		return endpoint + "/_json"
	}
	return endpoint + "/" + stream + "/_json"
}

func shouldRetryOpenObserveSend(err error) bool {
	if err == nil {
		return false
	}
	var sendErr openObserveSendError
	if errors.As(err, &sendErr) {
		return sendErr.statusCode == http.StatusTooManyRequests || sendErr.statusCode >= http.StatusInternalServerError
	}
	return true
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
