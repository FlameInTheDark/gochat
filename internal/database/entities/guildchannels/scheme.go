package guildchannels

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	addChannel               = `INSERT INTO gochat.guild_channels (guild_id, channel_id, position) VALUES (?, ?, ?)`
	removeChannel            = `DELETE FROM gochat.guild_channels WHERE guild_id = ? AND channel_id = ?`
	getGuildChannel          = `SELECT guild_id, channel_id, position FROM gochat.guild_channels WHERE guild_id = ? AND channel_id = ?`
	getGuildChannels         = `SELECT guild_id, channel_id, position FROM gochat.guild_channels WHERE guild_id = ?`
	getGuildByChannel        = `SELECT guild_id, channel_id, position FROM gochat.guild_channels WHERE channel_id = ?`
	setChannelPosition       = `UPDATE gochat.guild_channels SET position = ? WHERE guild_id = ? AND channel_id = ?`
	resetChannelPositionBulk = `UPDATE gochat.guild_channels SET position = 0 WHERE guild_id = ? AND channel_id IN ?`
)

func (e *Entity) AddChannel(ctx context.Context, guildID, channelID int64, position int) error {
	err := e.c.Session().
		Query(addChannel).
		WithContext(ctx).
		Bind(guildID, channelID, position).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add channel: %w", err)
	}
	return nil
}

func (e *Entity) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	var ch model.GuildChannel
	err := e.c.Session().
		Query(getGuildChannel).
		WithContext(ctx).
		Bind(guildID, channelID).
		Scan(&ch.GuildId, &ch.ChannelId, &ch.Position)
	if err != nil {
		return ch, fmt.Errorf("unable to get guild channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	var chans []model.GuildChannel
	iter := e.c.Session().
		Query(getGuildChannels).
		WithContext(ctx).
		Bind(guildID).
		Iter()
	var ch model.GuildChannel
	for iter.Scan(&ch.GuildId, &ch.ChannelId, &ch.Position) {
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
		Scan(&ch.GuildId, &ch.ChannelId, &ch.Position)
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

func (e *Entity) SetGuildChannelPosition(ctx context.Context, chs []model.GuildChannelUpdatePosition) error {
	if len(chs) == 0 {
		return nil
	}
	b := e.c.Session().NewBatch(gocql.LoggedBatch).WithContext(ctx)
	for _, ch := range chs {
		b.Query(setChannelPosition, ch.Position, ch.GuildId, ch.ChannelId)
	}
	err := e.c.Session().ExecuteBatch(b)
	if err != nil {
		return fmt.Errorf("unable to set guild channel position: %w", err)
	}
	return nil
}

func (e *Entity) ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error {
	err := e.c.Session().
		Query(resetChannelPositionBulk).
		WithContext(ctx).
		Bind(guildId, chs).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to reset guild channel position bulk: %w", err)
	}
	return nil
}
