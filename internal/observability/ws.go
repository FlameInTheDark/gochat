package observability

import (
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type WSTelemetry struct {
	activeConnections metric.Int64UpDownCounter
	activeSubs        metric.Int64UpDownCounter
	authFailures      metric.Int64Counter
	heartbeats        metric.Int64Counter
	heartbeatTimeouts metric.Int64Counter
	messagesIn        metric.Int64Counter
	messagesOut       metric.Int64Counter
	outboundDrops     metric.Int64Counter
	connectionSeconds metric.Float64Histogram
}

func NewWSTelemetry(serviceName string) *WSTelemetry {
	meter := otel.Meter(serviceName + "/ws")
	activeConnections, _ := meter.Int64UpDownCounter("gochat.ws.connections.active")
	activeSubs, _ := meter.Int64UpDownCounter("gochat.ws.subscriptions.active")
	authFailures, _ := meter.Int64Counter("gochat.ws.auth.failures")
	heartbeats, _ := meter.Int64Counter("gochat.ws.heartbeats")
	heartbeatTimeouts, _ := meter.Int64Counter("gochat.ws.heartbeat.timeouts")
	messagesIn, _ := meter.Int64Counter("gochat.ws.messages.in")
	messagesOut, _ := meter.Int64Counter("gochat.ws.messages.out")
	outboundDrops, _ := meter.Int64Counter("gochat.ws.messages.dropped")
	connectionSeconds, _ := meter.Float64Histogram("gochat.ws.connection.duration")

	return &WSTelemetry{
		activeConnections: activeConnections,
		activeSubs:        activeSubs,
		authFailures:      authFailures,
		heartbeats:        heartbeats,
		heartbeatTimeouts: heartbeatTimeouts,
		messagesIn:        messagesIn,
		messagesOut:       messagesOut,
		outboundDrops:     outboundDrops,
		connectionSeconds: connectionSeconds,
	}
}

func (t *WSTelemetry) ConnectionOpened(ctx context.Context, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	t.activeConnections.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *WSTelemetry) ConnectionClosed(ctx context.Context, started time.Time, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	t.activeConnections.Add(ctx, -1, metric.WithAttributes(attrs...))
	if !started.IsZero() {
		t.connectionSeconds.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	}
}

func (t *WSTelemetry) SubscriptionDelta(ctx context.Context, kind string, delta int64) {
	if t == nil || delta == 0 {
		return
	}
	t.activeSubs.Add(ctx, delta, metric.WithAttributes(attribute.String("subscription.kind", kind)))
}

func (t *WSTelemetry) AuthFailure(ctx context.Context, reason string) {
	if t == nil {
		return
	}
	t.authFailures.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
}

func (t *WSTelemetry) Heartbeat(ctx context.Context, result string) {
	if t == nil {
		return
	}
	t.heartbeats.Add(ctx, 1, metric.WithAttributes(attribute.String("result", result)))
}

func (t *WSTelemetry) HeartbeatTimeout(ctx context.Context) {
	if t == nil {
		return
	}
	t.heartbeatTimeouts.Add(ctx, 1)
}

func (t *WSTelemetry) MessageIn(ctx context.Context, operation int, eventType *int) {
	if t == nil {
		return
	}
	attrs := []attribute.KeyValue{
		attribute.String("message.operation", strconv.Itoa(operation)),
	}
	if eventType != nil {
		attrs = append(attrs, attribute.String("message.event_type", strconv.Itoa(*eventType)))
	}
	t.messagesIn.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *WSTelemetry) MessageOut(ctx context.Context, topic string, status string) {
	if t == nil {
		return
	}
	t.messagesOut.Add(ctx, 1, metric.WithAttributes(
		attribute.String("topic", topic),
		attribute.String("status", status),
	))
}

func (t *WSTelemetry) OutboundDrop(ctx context.Context, topic string) {
	if t == nil {
		return
	}
	t.outboundDrops.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topic)))
}
