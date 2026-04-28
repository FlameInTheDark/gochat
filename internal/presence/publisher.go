package presence

import (
	"context"
	"encoding/json"
	"fmt"
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

	msg, err := mqmsg.BuildEventMessage(&mqmsg.PresenceUpdate{
		UserID:           agg.UserID,
		Status:           agg.Status,
		CustomStatusText: agg.CustomStatusText,
		Since:            agg.Since,
		VoiceChannelID:   agg.VoiceChannelID,
		Mute:             agg.Mute,
		Deafen:           agg.Deafen,
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
