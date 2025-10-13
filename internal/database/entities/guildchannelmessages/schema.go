package guildchannelmessages

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
)

const (
	getChannelsMessages       = `SELECT channels FROM gochat.guild_channels_last_messages WHERE guild_id = ?`
	getChannelMessage         = `SELECT channels[?] as last_message_id FROM gochat.guild_channels_last_messages WHERE guild_id = ?`
	setChannelLastMessage     = `UPDATE gochat.guild_channels_last_messages SET channels[?] = ? WHERE guild_id = ?`
	setChannelLastMessageMany = `UPDATE gochat.guild_channels_last_messages SET channels = channels + ? WHERE guild_id = ?`
	getGuildsChannelsMessages = `SELECT guild_id, channels FROM gochat.guild_channels_last_messages WHERE guild_id IN ?`
)

func (e *Entity) GetChannelsMessages(ctx context.Context, guildId int64) (map[int64]int64, error) {
	var channelMessages map[int64]int64
	err := e.c.Session().
		Query(getChannelsMessages).
		WithContext(ctx).
		Bind(guildId).
		Scan(&channelMessages)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return channelMessages, nil
}

func (e *Entity) GetChannelMessage(ctx context.Context, guildId, channelId int64) (int64, error) {
	var lastMessageId *int64
	err := e.c.Session().
		Query(getChannelMessage).
		WithContext(ctx).
		Bind(channelId, guildId).
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

func (e *Entity) SetChannelLastMessage(ctx context.Context, guildId, channelId, lastMessageId int64) error {
	err := e.c.Session().
		Query(setChannelLastMessage).
		WithContext(ctx).
		Bind(channelId, lastMessageId, guildId).
		Exec()
	return err
}

func (e *Entity) SetReadStateMany(ctx context.Context, guildId, values map[int64]int64) error {
	err := e.c.Session().
		Query(setChannelLastMessageMany).
		WithContext(ctx).
		Bind(values, guildId).
		Exec()
	return err
}

func (e *Entity) GetChannelsMessagesForGuilds(ctx context.Context, guildIDs []int64) (map[int64]map[int64]int64, error) {
	out := make(map[int64]map[int64]int64, len(guildIDs))
	if len(guildIDs) == 0 {
		return out, nil
	}

	iter := e.c.Session().
		Query(getGuildsChannelsMessages).
		WithContext(ctx).
		Bind(guildIDs).
		Iter()

	var gid int64
	var chans map[int64]int64

	for iter.Scan(&gid, &chans) {
		out[gid] = chans
		gid = 0
		chans = nil
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return out, nil
}
