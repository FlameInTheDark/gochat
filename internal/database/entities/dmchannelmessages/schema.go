package dmchannelmessages

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gocql/gocql"
)

const (
	getChannelsMessages   = `SELECT channel_id, last_message_id FROM gochat.dm_channels_last_messages WHERE channel_id IN ?`
	getChannelMessage     = `SELECT last_message_id FROM gochat.dm_channels_last_messages WHERE channel_id = ?`
	setChannelLastMessage = `UPDATE gochat.dm_channels_last_messages SET last_message_id = ? WHERE channel_id = ?`
)

func (e *Entity) GetChannelsMessages(ctx context.Context, channels []int64) (map[int64]int64, error) {
	var channelsMessages = make(map[int64]int64)
	iter := e.c.Session().
		Query(getChannelsMessages).
		WithContext(ctx).
		Bind(channels).
		Iter()
	var channel, msgId int64
	for iter.Scan(&channel, &msgId) {
		channelsMessages[channel] = msgId
		channel = 0
		msgId = 0
	}

	if err := iter.Close(); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil
		}
		slog.Error("unable to get dm last messages", slog.String("error", err.Error()))
		return nil, err
	}
	return channelsMessages, nil
}

func (e *Entity) GetChannelMessage(ctx context.Context, channelId int64) (int64, error) {
	var lastMessageId *int64
	err := e.c.Session().
		Query(getChannelMessage).
		WithContext(ctx).
		Bind(channelId).
		Scan(&lastMessageId)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if lastMessageId == nil {
		return 0, err
	}
	return *lastMessageId, nil
}

func (e *Entity) SetChannelLastMessage(ctx context.Context, channelId, lastMessageId int64) error {
	err := e.c.Session().
		Query(setChannelLastMessage).
		WithContext(ctx).
		Bind(lastMessageId, channelId).
		Exec()
	return err
}
