package observability

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

type natsHeaderCarrier struct {
	headers nats.Header
}

func (c natsHeaderCarrier) Get(key string) string {
	if c.headers == nil {
		return ""
	}
	return c.headers.Get(key)
}

func (c natsHeaderCarrier) Set(key, value string) {
	if c.headers == nil {
		return
	}
	c.headers.Set(key, value)
}

func (c natsHeaderCarrier) Keys() []string {
	if c.headers == nil {
		return nil
	}
	keys := make([]string, 0, len(c.headers))
	for key := range c.headers {
		keys = append(keys, key)
	}
	return keys
}

func InjectNATSHeaders(ctx context.Context, headers nats.Header) nats.Header {
	if headers == nil {
		headers = nats.Header{}
	}
	textMapPropagator().Inject(ctx, natsHeaderCarrier{headers: headers})
	if rid, ok := helper.RequestIDFromContext(ctx); ok {
		headers.Set(RequestIDHeader, rid)
	}
	return headers
}

func ExtractNATSContext(ctx context.Context, msg *nats.Msg) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if msg == nil {
		return ctx
	}
	ctx = textMapPropagator().Extract(ctx, natsHeaderCarrier{headers: msg.Header})
	if rid := msg.Header.Get(RequestIDHeader); rid != "" {
		ctx = helper.ContextWithRequestID(ctx, rid)
	}
	return ctx
}

func StartNATSPublishSpan(ctx context.Context, subject string) (context.Context, func(error)) {
	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKey.String("nats"),
		semconv.MessagingDestinationName(subject),
		attribute.String("messaging.operation", "publish"),
	}
	ctx, span := otel.Tracer("gochat/nats").Start(ctx, fmt.Sprintf("nats publish %s", subject))
	span.SetAttributes(attrs...)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "published")
		}
		span.End()
	}
}

func StartNATSConsumeSpan(ctx context.Context, subject string) (context.Context, func(error)) {
	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKey.String("nats"),
		semconv.MessagingDestinationName(subject),
		attribute.String("messaging.operation", "process"),
	}
	ctx, span := otel.Tracer("gochat/nats").Start(ctx, fmt.Sprintf("nats consume %s", subject))
	span.SetAttributes(attrs...)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "processed")
		}
		span.End()
	}
}

func InjectMapCarrier(ctx context.Context, carrier propagation.TextMapCarrier) {
	textMapPropagator().Inject(ctx, carrier)
}
