package observability

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

type fiberHeaderCarrier struct {
	values map[string][]string
}

func (c fiberHeaderCarrier) Get(key string) string {
	values := c.values[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c fiberHeaderCarrier) Set(key, value string) {
	c.values[key] = []string{value}
}

func (c fiberHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.values))
	for key := range c.values {
		keys = append(keys, key)
	}
	return keys
}

type HTTPServerTelemetry struct {
	serviceName string

	requestsTotal     metric.Int64Counter
	authFailuresTotal metric.Int64Counter
	rateLimitTotal    metric.Int64Counter
	idempotencyHits   metric.Int64Counter
	serverErrorsTotal metric.Int64Counter
	inflightRequests  metric.Int64UpDownCounter
	requestDuration   metric.Float64Histogram
}

func RequestContextMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		ctx = textMapPropagator().Extract(ctx, fiberHeaderCarrier{values: c.GetReqHeaders()})

		requestID := strings.TrimSpace(c.Get(RequestIDHeader))
		if requestID == "" {
			requestID = generateRequestID()
		}
		if requestID != "" {
			ctx = helper.ContextWithRequestID(ctx, requestID)
			c.Set(RequestIDHeader, requestID)
			c.Locals("request_id", requestID)
		}

		c.SetUserContext(ctx)
		c.Locals("request_context", ctx)
		return c.Next()
	}
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	if logger == nil {
		logger = Logger()
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		ctx := c.UserContext()
		route := c.Path()
		if current := c.Route(); current != nil && current.Path != "" {
			route = current.Path
		}
		if route == "/healthz" {
			return err
		}

		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("route", route),
			slog.Int("status", responseStatusCode(c, err)),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			logger.ErrorContext(ctx, "request failed", attrs...)
			return err
		}

		logger.InfoContext(ctx, "request completed", attrs...)
		return nil
	}
}

func NewHTTPServerTelemetry(serviceName string) *HTTPServerTelemetry {
	meter := otel.Meter(serviceName)

	requestsTotal, _ := meter.Int64Counter("gochat.http.server.requests")
	authFailuresTotal, _ := meter.Int64Counter("gochat.http.server.auth_failures")
	rateLimitTotal, _ := meter.Int64Counter("gochat.http.server.rate_limit_hits")
	idempotencyHits, _ := meter.Int64Counter("gochat.http.server.idempotency_hits")
	serverErrorsTotal, _ := meter.Int64Counter("gochat.http.server.errors")
	inflightRequests, _ := meter.Int64UpDownCounter("gochat.http.server.inflight")
	requestDuration, _ := meter.Float64Histogram("gochat.http.server.duration")

	return &HTTPServerTelemetry{
		serviceName:       serviceName,
		requestsTotal:     requestsTotal,
		authFailuresTotal: authFailuresTotal,
		rateLimitTotal:    rateLimitTotal,
		idempotencyHits:   idempotencyHits,
		serverErrorsTotal: serverErrorsTotal,
		inflightRequests:  inflightRequests,
		requestDuration:   requestDuration,
	}
}

func (t *HTTPServerTelemetry) Middleware() fiber.Handler {
	if t == nil {
		return func(c *fiber.Ctx) error { return c.Next() }
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/healthz" {
			return c.Next()
		}

		ctx := c.UserContext()
		route := path
		spanName := c.Method() + " " + path
		tracer := otel.Tracer(t.serviceName)
		ctx, span := tracer.Start(ctx, spanName)
		c.SetUserContext(ctx)
		c.Locals("request_context", ctx)

		t.inflightRequests.Add(ctx, 1, metric.WithAttributes(attribute.String("service", t.serviceName)))
		start := time.Now()
		err := c.Next()
		code := responseStatusCode(c, err)
		if current := c.Route(); current != nil && current.Path != "" {
			route = current.Path
		}

		attrs := []attribute.KeyValue{
			attribute.String("service", t.serviceName),
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
		}
		requestAttrs := append(append([]attribute.KeyValue(nil), attrs...),
			attribute.Int("http.status_code", code),
		)
		spanAttrs := append(append([]attribute.KeyValue(nil), requestAttrs...),
			semconv.HTTPRequestMethodKey.String(c.Method()),
			semconv.HTTPResponseStatusCode(code),
			semconv.URLPath(path),
		)

		duration := time.Since(start).Seconds()
		t.requestsTotal.Add(ctx, 1, metric.WithAttributes(requestAttrs...))
		t.requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
		t.inflightRequests.Add(ctx, -1, metric.WithAttributes(attribute.String("service", t.serviceName)))

		if code == fiber.StatusUnauthorized || code == fiber.StatusForbidden {
			t.authFailuresTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		}
		if code == fiber.StatusTooManyRequests {
			t.rateLimitTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		}
		if idempotency.IsFromCache(c) {
			t.idempotencyHits.Add(ctx, 1, metric.WithAttributes(attrs...))
			span.AddEvent("idempotency.cache_hit")
		}
		if code >= fiber.StatusInternalServerError {
			t.serverErrorsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
			if err != nil {
				span.RecordError(err)
			}
			span.SetStatus(codes.Error, strconv.Itoa(code))
		} else {
			span.SetStatus(codes.Ok, strconv.Itoa(code))
		}
		span.SetAttributes(spanAttrs...)
		span.End()
		return err
	}
}

func AttachUserToFiberContext(c *fiber.Ctx) {
	if c == nil {
		return
	}
	user, err := helper.GetUser(c)
	if err != nil || user == nil {
		return
	}
	ctx := helper.ContextWithUserID(c.UserContext(), user.Id)
	c.SetUserContext(ctx)
	c.Locals("request_context", ctx)
}

func textMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func responseStatusCode(c *fiber.Ctx, err error) int {
	if err == nil {
		return c.Response().StatusCode()
	}
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code
	}
	return fiber.StatusInternalServerError
}
