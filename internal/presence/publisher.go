package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	natsio "github.com/nats-io/nats.go"
)

// Publish writes a presence update to the presence.user.<id> NATS subject.
func Publish(ctx context.Context, conn *natsio.Conn, agg Presence) (err error) {
	if conn == nil || agg.UserID == 0 {
		return nil
	}

	voiceID := agg.VoiceChannelID
	if agg.Status == StatusOffline && voiceID == nil {
		clearVoice := int64(0)
		voiceID = &clearVoice
	}

	msg, err := mqmsg.BuildEventMessage(&mqmsg.PresenceUpdate{
		UserID:           agg.UserID,
		Status:           agg.Status,
		CustomStatusText: agg.CustomStatusText,
		Since:            agg.Since,
		ClientStatus:     agg.ClientStatus,
		VoiceChannelID:   voiceID,
		Mute:             agg.Mute,
		Deafen:           agg.Deafen,
		SelfVideo:        agg.SelfVideo,
		ActiveStream:     agg.ActiveStream,
	})
	if err != nil {
		return err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal presence update: %w", err)
	}

	subject := fmt.Sprintf("presence.user.%d", agg.UserID)
	ctx, finish := observability.StartNATSPublishSpan(observability.BackgroundFromContext(ctx), subject)
	defer func() {
		finish(err)
	}()

	headers := observability.InjectNATSHeaders(ctx, nil)
	err = conn.PublishMsg(&natsio.Msg{
		Subject: subject,
		Header:  headers,
		Data:    body,
	})
	return err
}

// Refresh recomputes the aggregate presence, stores it, and publishes it.
func Refresh(ctx context.Context, store *Store, conn *natsio.Conn, userID, ttlSeconds int64) (Presence, error) {
	if store == nil || userID == 0 {
		return Presence{}, nil
	}

	agg, _, err := store.Aggregate(ctx, userID, time.Now().Unix())
	if err != nil {
		return Presence{}, err
	}
	if ttlSeconds > 0 {
		if err := store.SetAggregated(ctx, agg, ttlSeconds); err != nil {
			return Presence{}, err
		}
	}
	if err := Publish(ctx, conn, agg); err != nil {
		return Presence{}, err
	}
	return agg, nil
}

// ReconcileTouched recomputes recently touched users from Redis-backed leases
// and publishes only when the cached aggregate changed.
func ReconcileTouched(ctx context.Context, store *Store, conn *natsio.Conn, ttlSeconds, limit int64) (int, error) {
	if store == nil {
		return 0, nil
	}
	if ttlSeconds <= 0 {
		ttlSeconds = 60
	}
	if limit <= 0 {
		limit = 1000
	}

	now := time.Now().Unix()
	userIDs, err := store.ReconcileDue(ctx, now, limit)
	if err != nil {
		return 0, err
	}

	published := 0
	for _, userID := range userIDs {
		if _, err := store.PruneExpiredSessions(ctx, userID, now, 1000); err != nil {
			return published, err
		}

		prev, prevOK, err := store.CachedAggregate(ctx, userID)
		if err != nil {
			return published, err
		}

		agg, _, err := store.Aggregate(ctx, userID, now)
		if err != nil {
			return published, err
		}
		if err := store.SetAggregated(ctx, agg, ttlSeconds); err != nil {
			return published, err
		}
		if prevOK && presenceEqual(prev, agg) {
			continue
		}
		if err := Publish(ctx, conn, agg); err != nil {
			return published, err
		}
		published++
	}
	return published, nil
}

func presenceEqual(a, b Presence) bool {
	return a.UserID == b.UserID &&
		a.Status == b.Status &&
		a.CustomStatusText == b.CustomStatusText &&
		a.Since == b.Since &&
		equalInt64Ptr(a.VoiceChannelID, b.VoiceChannelID) &&
		a.Mute == b.Mute &&
		a.Deafen == b.Deafen &&
		a.SelfVideo == b.SelfVideo &&
		reflect.DeepEqual(a.ClientStatus, b.ClientStatus) &&
		reflect.DeepEqual(a.ActiveStream, b.ActiveStream)
}

func equalInt64Ptr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
