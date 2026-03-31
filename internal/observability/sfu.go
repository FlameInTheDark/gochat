package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type SFUTelemetry struct {
	activeChannels     metric.Int64UpDownCounter
	activePeers        metric.Int64UpDownCounter
	activeTracks       metric.Int64UpDownCounter
	joins              metric.Int64Counter
	leaves             metric.Int64Counter
	renegotiations     metric.Int64Counter
	offers             metric.Int64Counter
	answers            metric.Int64Counter
	candidates         metric.Int64Counter
	bitrateDisconnects metric.Int64Counter
	heartbeats         metric.Int64Counter
	heartbeatLatency   metric.Float64Histogram
	adminCloseLatency  metric.Float64Histogram
	connectionStates   metric.Int64Counter
	peerDuration       metric.Float64Histogram
	baseAttrs          []attribute.KeyValue
}

func NewSFUTelemetry(serviceName string, attrs ...attribute.KeyValue) *SFUTelemetry {
	meter := otel.Meter(serviceName + "/sfu")
	activeChannels, _ := meter.Int64UpDownCounter("gochat.sfu.channels.active")
	activePeers, _ := meter.Int64UpDownCounter("gochat.sfu.peers.active")
	activeTracks, _ := meter.Int64UpDownCounter("gochat.sfu.tracks.active")
	joins, _ := meter.Int64Counter("gochat.sfu.joins")
	leaves, _ := meter.Int64Counter("gochat.sfu.leaves")
	renegotiations, _ := meter.Int64Counter("gochat.sfu.renegotiations")
	offers, _ := meter.Int64Counter("gochat.sfu.offers")
	answers, _ := meter.Int64Counter("gochat.sfu.answers")
	candidates, _ := meter.Int64Counter("gochat.sfu.candidates")
	bitrateDisconnects, _ := meter.Int64Counter("gochat.sfu.bitrate_disconnects")
	heartbeats, _ := meter.Int64Counter("gochat.sfu.heartbeats")
	heartbeatLatency, _ := meter.Float64Histogram("gochat.sfu.heartbeat.duration")
	adminCloseLatency, _ := meter.Float64Histogram("gochat.sfu.admin_close.duration")
	connectionStates, _ := meter.Int64Counter("gochat.sfu.connection_state_changes")
	peerDuration, _ := meter.Float64Histogram("gochat.sfu.peer.duration")

	return &SFUTelemetry{
		baseAttrs:          append([]attribute.KeyValue(nil), attrs...),
		activeChannels:     activeChannels,
		activePeers:        activePeers,
		activeTracks:       activeTracks,
		joins:              joins,
		leaves:             leaves,
		renegotiations:     renegotiations,
		offers:             offers,
		answers:            answers,
		candidates:         candidates,
		bitrateDisconnects: bitrateDisconnects,
		heartbeats:         heartbeats,
		heartbeatLatency:   heartbeatLatency,
		adminCloseLatency:  adminCloseLatency,
		connectionStates:   connectionStates,
		peerDuration:       peerDuration,
	}
}

func (t *SFUTelemetry) ChannelDelta(ctx context.Context, delta int64) {
	if t == nil || delta == 0 {
		return
	}
	t.activeChannels.Add(ctx, delta)
}

func (t *SFUTelemetry) PeerDelta(ctx context.Context, delta int64, attrs ...attribute.KeyValue) {
	if t == nil || delta == 0 {
		return
	}
	attrs = t.withBaseAttrs(attrs...)
	t.activePeers.Add(ctx, delta, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) TrackDelta(ctx context.Context, kind string, delta int64, attrs ...attribute.KeyValue) {
	if t == nil || delta == 0 {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("track.kind", kind))
	t.activeTracks.Add(ctx, delta, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) Join(ctx context.Context, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = t.withBaseAttrs(attrs...)
	t.joins.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) Leave(ctx context.Context, started time.Time, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = t.withBaseAttrs(attrs...)
	t.leaves.Add(ctx, 1, metric.WithAttributes(attrs...))
	if !started.IsZero() {
		t.peerDuration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	}
}

func (t *SFUTelemetry) Renegotiation(ctx context.Context, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = t.withBaseAttrs(attrs...)
	t.renegotiations.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) Offer(ctx context.Context, direction string, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("direction", direction))
	t.offers.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) Answer(ctx context.Context, direction string, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("direction", direction))
	t.answers.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) Candidate(ctx context.Context, direction string, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("direction", direction))
	t.candidates.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) BitrateDisconnect(ctx context.Context, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = t.withBaseAttrs(attrs...)
	t.bitrateDisconnects.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) Heartbeat(ctx context.Context, started time.Time, status string, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("status", status))
	t.heartbeats.Add(ctx, 1, metric.WithAttributes(attrs...))
	if !started.IsZero() {
		t.heartbeatLatency.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	}
}

func (t *SFUTelemetry) AdminClose(ctx context.Context, started time.Time, status string, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("status", status))
	if !started.IsZero() {
		t.adminCloseLatency.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	}
}

func (t *SFUTelemetry) ConnectionState(ctx context.Context, state string, attrs ...attribute.KeyValue) {
	if t == nil {
		return
	}
	attrs = append(t.withBaseAttrs(attrs...), attribute.String("state", state))
	t.connectionStates.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (t *SFUTelemetry) withBaseAttrs(attrs ...attribute.KeyValue) []attribute.KeyValue {
	if t == nil || len(t.baseAttrs) == 0 {
		return attrs
	}
	out := make([]attribute.KeyValue, 0, len(t.baseAttrs)+len(attrs))
	out = append(out, t.baseAttrs...)
	out = append(out, attrs...)
	return out
}
