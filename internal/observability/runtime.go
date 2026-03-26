package observability

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	RequestIDHeader = "X-Request-ID"
)

var (
	initMu         sync.RWMutex
	defaultLogger  *slog.Logger
	serviceNameMap = make(map[string]string)

	shortLatencyHistogramBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	longLivedHistogramBuckets    = []float64{1, 5, 15, 30, 60, 120, 300, 600, 1800}
)

type Runtime struct {
	serviceName string
	logger      *slog.Logger

	traceProvider *sdktrace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
	logExporter   *openObserveLogExporter
}

type metricViewSpec struct {
	namePattern       string
	kind              sdkmetric.InstrumentKind
	boundaries        []float64
	allowedAttributes []attribute.Key
}

func Init(serviceName string, attrs ...attribute.KeyValue) (*Runtime, error) {
	serviceName = strings.TrimSpace(serviceName)
	resourceAttrs := buildResourceAttributes(serviceName, attrs)
	logAttrs := buildLogAttrs(attrs)
	stdoutHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(&contextualHandler{
		next:               stdoutHandler,
		serviceName:        serviceName,
		deploymentEnvValue: deploymentEnv(),
		fixedAttrs:         logAttrs,
	})
	rt := &Runtime{
		serviceName: serviceName,
		logger:      logger,
	}

	setDefaultLogger(rt.logger)

	otel.SetTextMapPropagator(textMapPropagator())

	var errs []error
	if shouldEnableOTLP() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := resource.New(
			ctx,
			resource.WithTelemetrySDK(),
			resource.WithProcess(),
			resource.WithHost(),
			resource.WithAttributes(resourceAttrs...),
		)
		if err != nil {
			errs = append(errs, err)
		} else {
			traceExporter, traceErr := otlptracehttp.New(ctx)
			if traceErr != nil {
				logger.Warn("unable to initialize OTLP trace exporter", slog.String("error", traceErr.Error()))
			} else {
				tp := sdktrace.NewTracerProvider(
					sdktrace.WithResource(res),
					sdktrace.WithSpanProcessor(newFilteringSpanProcessor(
						sdktrace.NewBatchSpanProcessor(traceExporter),
						ParseLogLevel(os.Getenv("LOG_LEVEL")),
					)),
				)
				otel.SetTracerProvider(tp)
				rt.traceProvider = tp
			}

			metricExporter, metricErr := otlpmetrichttp.New(ctx)
			if metricErr != nil {
				logger.Warn("unable to initialize OTLP metric exporter", slog.String("error", metricErr.Error()))
			} else {
				mp := sdkmetric.NewMeterProvider(
					sdkmetric.WithResource(res),
					sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(15*time.Second))),
					sdkmetric.WithView(metricViews()...),
				)
				otel.SetMeterProvider(mp)
				rt.meterProvider = mp
			}

			if traceErr != nil && metricErr != nil {
				errs = append(errs, errors.Join(traceErr, metricErr))
			}
		}
	}

	if cfg, enabled, err := loadOpenObserveLogExporterConfigFromEnv(); err != nil {
		errs = append(errs, err)
		logger.Warn("unable to initialize OpenObserve log exporter", slog.String("error", err.Error()))
	} else if enabled {
		exporter, exportErr := newOpenObserveLogExporter(rt.serviceName, cfg)
		if exportErr != nil {
			errs = append(errs, exportErr)
			logger.Warn("unable to initialize OpenObserve log exporter", slog.String("error", exportErr.Error()))
		} else {
			rt.logExporter = exporter
			handlers := []slog.Handler{
				stdoutHandler,
				slog.NewJSONHandler(exporter.Writer(), nil),
			}
			rt.logger = slog.New(&contextualHandler{
				next:               &fanoutHandler{handlers: handlers},
				serviceName:        serviceName,
				deploymentEnvValue: deploymentEnv(),
				fixedAttrs:         logAttrs,
			})
			setDefaultLogger(rt.logger)
		}
	}

	return rt, errors.Join(errs...)
}

func (r *Runtime) Logger() *slog.Logger {
	if r == nil || r.logger == nil {
		return fallbackLogger()
	}
	return r.logger
}

func (r *Runtime) Tracer(scope string) trace.Tracer {
	name := normalizeScope(r.serviceName, scope)
	return otel.Tracer(name)
}

func (r *Runtime) Meter(scope string) metric.Meter {
	name := normalizeScope(r.serviceName, scope)
	return otel.Meter(name)
}

func (r *Runtime) Close() error {
	if r == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var errs []error
	if r.logExporter != nil {
		errs = append(errs, r.logExporter.Close())
	}
	if r.meterProvider != nil {
		errs = append(errs, r.meterProvider.Shutdown(ctx))
	}
	if r.traceProvider != nil {
		errs = append(errs, r.traceProvider.Shutdown(ctx))
	}

	return errors.Join(errs...)
}

func Logger() *slog.Logger {
	initMu.RLock()
	logger := defaultLogger
	initMu.RUnlock()
	if logger != nil {
		return logger
	}
	return fallbackLogger()
}

func Tracer(scope string) trace.Tracer {
	return otel.Tracer(scope)
}

func Meter(scope string) metric.Meter {
	return otel.Meter(scope)
}

func deploymentEnv() string {
	if value := strings.TrimSpace(os.Getenv("GOCHAT_DEPLOYMENT_ENV")); value != "" {
		return value
	}
	return "local"
}

func shouldEnableOTLP() bool {
	for _, key := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
	} {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return true
		}
	}
	return false
}

func normalizeScope(serviceName, scope string) string {
	scope = strings.TrimSpace(scope)
	if scope != "" {
		return scope
	}
	if serviceName != "" {
		return serviceName
	}
	return "gochat"
}

func buildResourceAttributes(serviceName string, extra []attribute.KeyValue) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.DeploymentEnvironment(deploymentEnv()),
		attribute.String("service.name", serviceName),
		attribute.String("deployment.environment", deploymentEnv()),
	}
	for _, attr := range extra {
		if strings.TrimSpace(string(attr.Key)) == "" {
			continue
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

func buildLogAttrs(extra []attribute.KeyValue) []slog.Attr {
	if len(extra) == 0 {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(extra))
	for _, attr := range extra {
		key := strings.TrimSpace(string(attr.Key))
		if key == "" {
			continue
		}
		attrs = append(attrs, slog.Any(key, attr.Value.AsInterface()))
	}
	return attrs
}

func setDefaultLogger(logger *slog.Logger) {
	initMu.Lock()
	defer initMu.Unlock()
	defaultLogger = logger
}

func fallbackLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func metricViews() []sdkmetric.View {
	specs := []metricViewSpec{
		{
			namePattern:       "gochat.http.server.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("service"), attribute.Key("http.method"), attribute.Key("http.route")},
		},
		{
			namePattern:       "gochat.dependency.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("dependency.system"), attribute.Key("dependency.operation"), attribute.Key("dependency.result")},
		},
		{
			namePattern:       "gochat.postgres.probe.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("postgres.target"), attribute.Key("probe")},
		},
		{
			namePattern:       "gochat.indexer.consume.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("subject"), attribute.Key("operation")},
		},
		{
			namePattern:       "gochat.embedder.consume.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("subject")},
		},
		{
			namePattern:       "gochat.sfu.heartbeat.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("status")},
		},
		{
			namePattern:       "gochat.sfu.admin_close.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        shortLatencyHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("status")},
		},
		{
			namePattern: "gochat.logs.exporter.batch.duration",
			kind:        sdkmetric.InstrumentKindHistogram,
			boundaries:  shortLatencyHistogramBuckets,
		},
		{
			namePattern:       "gochat.ws.connection.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        longLivedHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("ws.compression")},
		},
		{
			namePattern:       "gochat.sfu.peer.duration",
			kind:              sdkmetric.InstrumentKindHistogram,
			boundaries:        longLivedHistogramBuckets,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id")},
		},
		{
			namePattern:       "gochat.http.server.requests",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("service"), attribute.Key("http.method"), attribute.Key("http.route"), attribute.Key("http.status_code")},
		},
		{
			namePattern:       "gochat.http.server.auth_failures",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("service"), attribute.Key("http.method"), attribute.Key("http.route")},
		},
		{
			namePattern:       "gochat.http.server.rate_limit_hits",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("service"), attribute.Key("http.method"), attribute.Key("http.route")},
		},
		{
			namePattern:       "gochat.http.server.idempotency_hits",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("service"), attribute.Key("http.method"), attribute.Key("http.route")},
		},
		{
			namePattern:       "gochat.http.server.errors",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("service"), attribute.Key("http.method"), attribute.Key("http.route")},
		},
		{
			namePattern:       "gochat.dependency.calls",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("dependency.system"), attribute.Key("dependency.operation"), attribute.Key("dependency.result")},
		},
		{
			namePattern:       "gochat.dependency.errors",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("dependency.system"), attribute.Key("dependency.operation"), attribute.Key("dependency.result")},
		},
		{
			namePattern:       "gochat.sfu.peers.active",
			kind:              sdkmetric.InstrumentKindUpDownCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id")},
		},
		{
			namePattern:       "gochat.sfu.tracks.active",
			kind:              sdkmetric.InstrumentKindUpDownCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("track.kind")},
		},
		{
			namePattern:       "gochat.sfu.joins",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id")},
		},
		{
			namePattern:       "gochat.sfu.leaves",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id")},
		},
		{
			namePattern:       "gochat.sfu.renegotiations",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id")},
		},
		{
			namePattern:       "gochat.sfu.offers",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("direction")},
		},
		{
			namePattern:       "gochat.sfu.answers",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("direction")},
		},
		{
			namePattern:       "gochat.sfu.candidates",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("direction")},
		},
		{
			namePattern:       "gochat.sfu.bitrate_disconnects",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id")},
		},
		{
			namePattern:       "gochat.sfu.heartbeats",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("status")},
		},
		{
			namePattern:       "gochat.sfu.connection_state_changes",
			kind:              sdkmetric.InstrumentKindCounter,
			allowedAttributes: []attribute.Key{attribute.Key("voice.region"), attribute.Key("service.instance.id"), attribute.Key("state")},
		},
	}

	views := make([]sdkmetric.View, 0, len(specs))
	for _, spec := range specs {
		stream := sdkmetric.Stream{}
		if len(spec.boundaries) > 0 {
			stream.Aggregation = sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: append([]float64(nil), spec.boundaries...),
				NoMinMax:   true,
			}
		}
		if len(spec.allowedAttributes) > 0 {
			stream.AttributeFilter = attribute.NewAllowKeysFilter(spec.allowedAttributes...)
		}
		views = append(views, sdkmetric.NewView(sdkmetric.Instrument{
			Name: spec.namePattern,
			Kind: spec.kind,
		}, stream))
	}
	return views
}
