package guild

import (
	"context"
	"time"

	cachepkg "github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

func (e *entity) scheduleStreamRebinds(ctx context.Context, guildID, channelID int64, region string, delayMs int64) {
	if e == nil || e.cache == nil || e.streamDisco == nil {
		return
	}

	streams, err := e.listChannelStreams(ctx, channelID)
	if err != nil || len(streams) == 0 {
		return
	}

	type plannedRebind struct {
		streamID int64
		binding  streammeta.RouteBinding
	}

	planned := make([]plannedRebind, 0, len(streams))
	streamIDs := make([]int64, 0, len(streams))
	for _, stream := range streams {
		binding, err := e.selectStreamBinding(ctx, stream.ID, region, false)
		if err != nil || binding.URL == "" {
			continue
		}
		planned = append(planned, plannedRebind{streamID: stream.ID, binding: binding})
		streamIDs = append(streamIDs, stream.ID)
		_ = e.cache.SetTimedJSON(ctx, streammeta.RebindKey(stream.ID), binding, streamRebindTTLSeconds, cachepkg.NoneProactive())
	}

	if len(planned) == 0 {
		return
	}

	asyncCtx := observability.BackgroundFromContext(ctx)
	go func() {
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
		for _, item := range planned {
			_ = e.cache.SetTimedJSON(asyncCtx, streammeta.RouteKey(item.streamID), item.binding, streamRouteInitialTTLSeconds, cachepkg.NoneProactive())
			_ = e.cache.SetTimedJSON(asyncCtx, streammeta.RebindKey(item.streamID), item.binding, streamRebindTTLSeconds, cachepkg.NoneProactive())
		}
		_ = mq.SendGuildUpdate(asyncCtx, e.mqt, guildID, &mqmsg.GuildStreamsRebind{
			GuildId:   guildID,
			ChannelId: channelID,
			StreamIds: streamIDs,
			JitterMs:  int(delayMs),
		})
	}()
}
