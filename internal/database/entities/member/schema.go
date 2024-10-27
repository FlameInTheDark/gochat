package member

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	addMember       = `INSERT INTO gochat.members (user_id, guild_id, join_at) VALUES (?, ?, toTimestamp(now()))`
	removeMember    = `DELETE FROM gochat.members WHERE user_id = ? AND guild_id = ?`
	getMember       = `SELECT user_id, guild_id, username, avatar, join_at FROM gochat.members WHERE guild_id = ? AND user_id = ?`
	getGuildMembers = `SELECT user_id, guild_id, username, avatar, join_at FROM gochat.members WHERE guild_id = ?`
	isGuildMember   = `SELECT count(user_id) FROM gochat.members WHERE guild_id = ? AND user_id = ? LIMIT 1`
	getUserGuilds   = `SELECT user_id, guild_id FROM gochat.members WHERE user_id = ?`
)

func (e *Entity) AddMember(ctx context.Context, userID, guildID int64) error {
	err := e.c.Session().
		Query(addMember).
		WithContext(ctx).
		Bind(userID, guildID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add member: %w", err)
	}
	return nil
}

func (e *Entity) RemoveMember(ctx context.Context, userID, guildID int64) error {
	err := e.c.Session().
		Query(removeMember).
		WithContext(ctx).
		Bind(userID, guildID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove member: %w", err)
	}
	return nil
}

func (e *Entity) GetMember(ctx context.Context, userId, guildId int64) (model.Member, error) {
	var m model.Member
	err := e.c.Session().
		Query(getMember).
		WithContext(ctx).
		Bind(guildId, userId).
		Scan(&m.UserId, &m.GuildId, &m.Username, &m.Avatar, &m.JoinAt)
	if err != nil {
		return m, fmt.Errorf("unable to get member: %w", err)
	}
	return m, nil
}

func (e *Entity) GetGuildMembers(ctx context.Context, guildId int64) ([]model.Member, error) {
	var members []model.Member
	iter := e.c.Session().
		Query(getGuildMembers).
		WithContext(ctx).
		Bind(guildId).
		Iter()
	var m model.Member
	for iter.Scan(&m.UserId, &m.GuildId, &m.Username, &m.Avatar, &m.JoinAt) {
		members = append(members, m)
	}
	err := iter.Close()
	if err != nil {
		return members, fmt.Errorf("unable to get members: %w", err)
	}
	return members, nil
}

func (e *Entity) IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error) {
	var count int
	err := e.c.Session().
		Query(isGuildMember).
		WithContext(ctx).
		Bind(guildId, userId).
		Scan(&count)
	if err != nil {
		return false, fmt.Errorf("unable to check if guild member exists: %w", err)
	}
	return count > 0, nil
}

func (e *Entity) GetUserGuilds(ctx context.Context, userId int64) ([]model.UserGuild, error) {
	var guilds []model.UserGuild
	iter := e.c.Session().
		Query(getUserGuilds).
		WithContext(ctx).
		Bind(userId).
		Iter()
	var g model.UserGuild
	for iter.Scan(&g.UserId, &g.GuildId) {
		guilds = append(guilds, g)
	}
	err := iter.Close()
	if err != nil {
		return guilds, fmt.Errorf("unable to get user guilds: %w", err)
	}
	return guilds, nil
}
