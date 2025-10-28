package mention

import (
	"context"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gocql/gocql"
)

const (
	addMention                = `INSERT INTO gochat.mentions (user_id, channel_id, message_id, author_id) VALUES (?, ?, ?, ?)`
	getMentionsAfter          = `SELECT user_id, channel_id, message_id, author_id FROM gochat.mentions WHERE user_id = ? AND channel_id = ? AND message_id > ?`
	getMentionsBefore         = `SELECT user_id, channel_id, message_id, author_id FROM gochat.mentions WHERE user_id = ? AND channel_id = ? AND message_id < ?`
	getMentionsAfterMany      = `SELECT user_id, channel_id, message_id, author_id FROM gochat.mentions WHERE user_id = ? AND channel_id IN ? AND message_id > ?`
	addGuildMention           = `INSERT INTO gochat.channel_mentions (guild_id, channel_id, message_id, role_id, author_id, type) VALUES (?, ?, ?, ?, ?, ?)`
	getGuildMentionsAfter     = `SELECT guild_id, channel_id, message_id, role_id, author_id, type FROM gochat.channel_mentions WHERE channel_id = ? AND message_id > ?`
	getGuildMentionsBefore    = `SELECT guild_id, channel_id, message_id, role_id, author_id, type FROM gochat.channel_mentions WHERE channel_id = ? AND message_id < ?`
	getGuildMentionsAfterMany = `SELECT guild_id, channel_id, message_id, role_id, author_id, type FROM gochat.channel_mentions WHERE channel_id IN ? AND message_id > ?`
)

func (e *Entity) AddMention(ctx context.Context, userId, channelId, messageId, authorId int64) error {
	err := e.c.Session().
		Query(addMention).
		WithContext(ctx).
		Bind(userId, channelId, messageId, authorId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add mention: %w", err)
	}
	return nil
}

func (e *Entity) GetMentionsBefore(ctx context.Context, userId, channelId, messageId int64) ([]model.Mention, error) {
	var mentions []model.Mention

	iter := e.c.Session().
		Query(getMentionsBefore).
		WithContext(ctx).
		Bind(userId, channelId, messageId).
		Iter()
	var m model.Mention
	for iter.Scan(&m.UserId, &m.ChannelId, &m.MessageId, &m.AuthorId) {
		mentions = append(mentions, m)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get mentions before: %w", err)
	}

	return mentions, nil
}

func (e *Entity) GetMentionsAfter(ctx context.Context, userId, channelId, messageId int64) ([]model.Mention, error) {
	var mentions []model.Mention

	iter := e.c.Session().
		Query(getMentionsAfter).
		WithContext(ctx).
		Bind(userId, channelId, messageId).
		Iter()
	var m model.Mention
	for iter.Scan(&m.UserId, &m.ChannelId, &m.MessageId, &m.AuthorId) {
		mentions = append(mentions, m)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get mentions after: %w", err)
	}

	return mentions, nil
}

func (e *Entity) GetMentionsAfterMany(ctx context.Context, userId, messageId int64, channelIds []int64) (map[int64][]model.Mention, error) {
	var mentions = make(map[int64][]model.Mention)

	iter := e.c.Session().
		Query(getMentionsAfterMany).
		WithContext(ctx).
		Bind(userId, channelIds, messageId).
		Iter()
	var m model.Mention
	for iter.Scan(&m.UserId, &m.ChannelId, &m.MessageId, &m.AuthorId) {
		mentions[m.ChannelId] = append(mentions[m.ChannelId], m)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get mentions after: %w", err)
	}

	return mentions, nil
}

func (e *Entity) AddChannelMention(ctx context.Context, guildId, channelId, messageId, authorId int64, roleId *int64, mtype model.ChannelMentionType) error {
	err := e.c.Session().
		Query(addGuildMention).
		WithContext(ctx).
		Bind(guildId, channelId, messageId, roleId, authorId, int(mtype)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add mention: %w", err)
	}
	return nil
}

func (e *Entity) GetChannelMentionsBefore(ctx context.Context, channelId, messageId int64) ([]model.ChannelMention, error) {
	var mentions []model.ChannelMention

	iter := e.c.Session().
		Query(getGuildMentionsBefore).
		WithContext(ctx).
		Bind(channelId, messageId).
		Iter()
	var m model.ChannelMention
	var mtype int
	for iter.Scan(&m.GuildId, &m.ChannelId, &m.MessageId, &m.RoleId, &m.AuthorId, &mtype) {
		m.Type = mtype
		mentions = append(mentions, m)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get mentions before: %w", err)
	}

	return mentions, nil
}

func (e *Entity) GetChannelMentionsAfter(ctx context.Context, channelId, messageId int64) ([]model.ChannelMention, error) {
	var mentions []model.ChannelMention

	iter := e.c.Session().
		Query(getGuildMentionsAfter).
		WithContext(ctx).
		Bind(channelId, messageId).
		Iter()
	var m model.ChannelMention
	var mtype int
	for iter.Scan(&m.GuildId, &m.ChannelId, &m.MessageId, &m.RoleId, &m.AuthorId, &mtype) {
		m.Type = mtype
		mentions = append(mentions, m)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get mentions after: %w", err)
	}

	return mentions, nil
}

func (e *Entity) GetChannelMentionsAfterMany(ctx context.Context, messageId int64, channelIds []int64) (map[int64][]model.ChannelMention, error) {
	var mentions = make(map[int64][]model.ChannelMention)

	iter := e.c.Session().
		Query(getGuildMentionsAfterMany).
		WithContext(ctx).
		Bind(channelIds, messageId).
		Iter()
	var m model.ChannelMention
	var mtype int
	for iter.Scan(&m.GuildId, &m.ChannelId, &m.MessageId, &m.RoleId, &m.AuthorId, &mtype) {
		m.Type = mtype
		mentions[m.ChannelId] = append(mentions[m.ChannelId], m)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get mentions after: %w", err)
	}

	return mentions, nil
}
