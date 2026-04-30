package guild

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	cachepkg "github.com/FlameInTheDark/gochat/internal/cache"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

const (
	streamRouteInitialTTLSeconds = int64(180)
	streamStateTTLSeconds        = int64(180)
	streamRebindTTLSeconds       = int64(300)
)

func (e *entity) effectiveStreamRegion(ctx context.Context, channelID int64) string {
	if binding := e.cachedChannelBinding(ctx, channelID); binding.Region != "" {
		return binding.Region
	}
	return e.preferredVoiceRegion(ctx, channelID)
}

func (e *entity) cachedStreamBinding(ctx context.Context, streamID int64) streammeta.RouteBinding {
	var binding streammeta.RouteBinding
	if e == nil || e.cache == nil {
		return binding
	}
	_ = e.cache.GetJSON(ctx, streammeta.RouteKey(streamID), &binding)
	return binding
}

func (e *entity) bindStreamRoute(ctx context.Context, streamID int64, binding streammeta.RouteBinding) streammeta.RouteBinding {
	if e == nil || e.cache == nil || binding.URL == "" {
		return binding
	}

	set, err := e.cache.SetTimedJSONNX(ctx, streammeta.RouteKey(streamID), binding, streamRouteInitialTTLSeconds, cachepkg.NoneProactive())
	if err == nil && set {
		return binding
	}

	if winner := e.cachedStreamBinding(ctx, streamID); winner.URL != "" {
		return winner
	}

	return binding
}

func (e *entity) streamBindingForJoin(ctx context.Context, streamID, channelID int64) (streammeta.RouteBinding, error) {
	if binding := e.cachedStreamBinding(ctx, streamID); binding.URL != "" {
		return binding, nil
	}

	binding, err := e.selectStreamBinding(ctx, streamID, e.effectiveStreamRegion(ctx, channelID), false)
	if err != nil {
		return streammeta.RouteBinding{}, err
	}
	if binding.URL == "" {
		return streammeta.RouteBinding{}, nil
	}

	return e.bindStreamRoute(ctx, streamID, binding), nil
}

func (e *entity) selectStreamBinding(ctx context.Context, streamID int64, preferredRegion string, allowFallback bool) (streammeta.RouteBinding, error) {
	if e == nil || e.streamDisco == nil {
		return streammeta.RouteBinding{}, nil
	}

	regions, err := e.selectionRegions(ctx, preferredRegion, allowFallback)
	if err != nil {
		return streammeta.RouteBinding{}, err
	}

	var firstErr error
	for _, region := range regions {
		instances, err := e.cachedStreamRegionInstances(ctx, region)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if len(instances) == 0 {
			continue
		}
		if binding, ok := e.streamSelector.pickBinding(streamID, region, instances); ok {
			return streammeta.RouteBinding{ID: binding.ID, URL: binding.URL, Region: binding.Region}, nil
		}
	}

	return streammeta.RouteBinding{}, firstErr
}

func (e *entity) cachedStreamRegionInstances(ctx context.Context, region string) ([]discovery.Instance, error) {
	if e == nil || e.streamDisco == nil || e.streamSelector == nil || region == "" {
		return nil, nil
	}

	for attempts := 0; attempts < 2; attempts++ {
		if instances, ok := e.streamSelector.regionSnapshot(region, time.Now()); ok {
			return instances, nil
		}

		waitCh, refresh := e.streamSelector.beginRegionRefresh(region)
		if !refresh {
			<-waitCh
			continue
		}

		instances, err := e.streamDisco.List(ctx, region)
		if err != nil {
			e.streamSelector.finishRegionRefresh(region, nil, false)
			if stale, ok := e.streamSelector.staleRegionSnapshot(region, time.Now()); ok {
				if e.log != nil {
					e.log.Warn("stream discovery using stale region snapshot", "region", region, "error", err.Error())
				}
				return stale, nil
			}
			return nil, err
		}

		e.streamSelector.finishRegionRefresh(region, instances, true)
		return cloneDiscoveryInstances(instances), nil
	}

	return nil, nil
}

func (e *entity) listChannelStreams(ctx context.Context, channelID int64) ([]streammeta.Metadata, error) {
	if e == nil || e.cache == nil {
		return nil, nil
	}

	rows, err := e.cache.HGetAll(ctx, streammeta.ChannelKey(channelID))
	if err != nil {
		return nil, err
	}

	streams := make([]streammeta.Metadata, 0, len(rows))
	for _, raw := range rows {
		if raw == "" {
			continue
		}
		var meta streammeta.Metadata
		if err := json.Unmarshal([]byte(raw), &meta); err != nil {
			continue
		}
		if meta.ID == 0 {
			continue
		}
		streams = append(streams, meta)
	}

	sort.Slice(streams, func(i, j int) bool {
		if streams[i].StartedAt == streams[j].StartedAt {
			return streams[i].ID < streams[j].ID
		}
		return streams[i].StartedAt < streams[j].StartedAt
	})

	return streams, nil
}
