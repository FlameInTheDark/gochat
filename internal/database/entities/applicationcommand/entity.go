package applicationcommand

import (
	"context"
	"time"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/database/db"
)

type ApplicationCommandInteraction interface {
	CreateInteractionPayload(ctx context.Context, record appcmd.InteractionPayloadRecord) error
	GetInteractionPayload(ctx context.Context, interactionID int64) (appcmd.InteractionPayloadRecord, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) ApplicationCommandInteraction {
	return &Entity{c: c}
}

const (
	createInteractionPayload = `INSERT INTO gochat.application_command_interactions (id, application_id, command_id, type, guild_id, channel_id, invoker_user_id, data, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	getInteractionPayload    = `SELECT id, application_id, command_id, type, guild_id, channel_id, invoker_user_id, data, created_at FROM gochat.application_command_interactions WHERE id = ?`
)

func (e *Entity) CreateInteractionPayload(ctx context.Context, record appcmd.InteractionPayloadRecord) error {
	createdAt := record.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return e.c.Session().Query(createInteractionPayload).
		WithContext(ctx).
		Bind(record.ID, record.ApplicationID, record.CommandID, int(record.Type), record.GuildID, record.ChannelID, record.InvokerUserID, record.DataJSON, createdAt).
		Exec()
}

func (e *Entity) GetInteractionPayload(ctx context.Context, interactionID int64) (appcmd.InteractionPayloadRecord, error) {
	var record appcmd.InteractionPayloadRecord
	var t int
	if err := e.c.Session().Query(getInteractionPayload).
		WithContext(ctx).
		Bind(interactionID).
		Scan(&record.ID, &record.ApplicationID, &record.CommandID, &t, &record.GuildID, &record.ChannelID, &record.InvokerUserID, &record.DataJSON, &record.CreatedAt); err != nil {
		return appcmd.InteractionPayloadRecord{}, err
	}
	record.Type = appcmd.InteractionType(t)
	return record, nil
}
