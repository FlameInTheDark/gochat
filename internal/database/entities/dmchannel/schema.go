package dmchannel

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getDmChannel    = `SELECT user_id, participant_id, channel_id FROM gochat.dm_channels WHERE user_id = ? AND participant_id = ?`
	createDmChannel = `INSERT INTO gochat.dm_channels(user_id, participant_id, channel_id) VALUES(?, ?, ?);`
)

func (e *Entity) GetDmChannel(ctx context.Context, userId, participantId int64) (model.DMChannel, error) {
	var ch model.DMChannel
	err := e.c.Session().
		Query(getDmChannel).
		WithContext(ctx).
		Bind(userId, participantId).
		Scan(&ch.UserId, &ch.ParticipantId, &ch.ChannelId)
	if err != nil {
		return model.DMChannel{}, fmt.Errorf("unable to get dm channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) CreateDmChannel(ctx context.Context, userId, participantId, channelId int64) error {
	b := e.c.Session().NewBatch(gocql.UnloggedBatch).WithContext(ctx)
	b.Query(createDmChannel, userId, participantId, channelId)
	b.Query(createDmChannel, participantId, userId, channelId)
	err := e.c.Session().ExecuteBatch(b)
	if err != nil {
		return fmt.Errorf("unable to create dm channel: %w", err)
	}
	return nil
}
