package guildchannels

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	addChannel        = `INSERT INTO gochat.guild_channels (guild_id, channel_id, permissions) VALUES (?, ?, ?)`
	removeChannel     = `DELETE FROM gochat.guild_channels WHERE guild_id = ? AND channel_id = ?`
	getGuildChannels  = `SELECT guild_id, channel_id, permissions FROM gochat.guild_channels WHERE guild_id = ?`
	getGuildByChannel = `SELECT guild_id, channel_id, permissions FROM gochat.guild_channels WHERE channel_id = ?`
)

func (e *Entity) AddChannel(ctx context.Context, guildID, channelID, permissions int64) error {
	err := e.c.Session().
		Query(addChannel).
		WithContext(ctx).
		Bind(guildID, channelID, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add channel: %w", err)
	}
	return nil
}

func (e *Entity) GetGuildChannels(ctx context.Context, guildID string) ([]model.GuildChannel, error) {
	var chans []model.GuildChannel
	iter := e.c.Session().
		Query(getGuildChannels).
		WithContext(ctx).
		Bind(guildID).
		Iter()
	var ch model.GuildChannel
	for iter.Scan(&ch.GuildId, &ch.ChannelId, &ch.Permissions) {
		chans = append(chans, ch)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get guild channels: %w", err)
	}
	return chans, nil
}

func (e *Entity) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	var ch model.GuildChannel
	err := e.c.Session().
		Query(getGuildByChannel).
		WithContext(ctx).
		Bind(channelID).
		Scan(&ch.GuildId, &ch.ChannelId, &ch.Permissions)
	if err != nil {
		return model.GuildChannel{}, fmt.Errorf("unable to get guild by channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) RemoveChannel(ctx context.Context, guildID, channelID string) error {
	err := e.c.Session().
		Query(removeChannel).
		WithContext(ctx).
		Bind(guildID, channelID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove channel: %w", err)
	}
	return nil
}
