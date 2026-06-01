package applicationcommanddispatch

import (
	"context"
	"encoding/json"
	"fmt"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/botgateway"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	natsio "github.com/nats-io/nats.go"
)

type BotSessionRegistry interface {
	BotSessions(ctx context.Context, botUserID int64) ([]botgateway.Session, error)
}

func DispatchInteraction(ctx context.Context, nc *natsio.Conn, registry BotSessionRegistry, botUserID int64, guildID *int64, interaction appcmd.Interaction) error {
	if nc == nil {
		return fmt.Errorf("bot gateway nats connection is nil")
	}
	sessions, err := registry.BotSessions(ctx, botUserID)
	if err != nil {
		return fmt.Errorf("list bot sessions: %w", err)
	}
	if guildID != nil {
		sessions = botgateway.TargetGuildSessions(sessions, *guildID)
	} else {
		sessions = botgateway.TargetDMSessions(sessions)
	}
	if len(sessions) == 0 {
		return nil
	}
	envelope, err := mqmsg.BuildEventMessage(&mqmsg.ApplicationCommandInteractionCreate{Interaction: interaction})
	if err != nil {
		return fmt.Errorf("build interaction event: %w", err)
	}
	event, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal interaction event: %w", err)
	}
	for instanceID, sessionIDs := range botgateway.GroupByInstance(sessions) {
		body, err := json.Marshal(botgateway.DeliveryMessage{SessionIDs: sessionIDs, Event: event})
		if err != nil {
			return fmt.Errorf("marshal delivery message: %w", err)
		}
		subject := botgateway.InstanceDeliverySubject(instanceID)
		headers := observability.InjectNATSHeaders(ctx, nil)
		if err := nc.PublishMsg(&natsio.Msg{Subject: subject, Header: headers, Data: body}); err != nil {
			return fmt.Errorf("publish interaction delivery: %w", err)
		}
	}
	return nil
}
