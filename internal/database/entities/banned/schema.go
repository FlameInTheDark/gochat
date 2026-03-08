package banned

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	banUser      = `INSERT INTO gochat.banned (guild_id, user_id, reason) VALUES (?, ?, ?)`
	unbanUser    = `DELETE FROM gochat.banned WHERE guild_id = ? AND user_id = ?`
	isBanned     = `SELECT count(user_id) FROM gochat.banned WHERE guild_id = ? AND user_id = ? LIMIT 1`
	getGuildBans = `SELECT user_id, reason FROM gochat.banned WHERE guild_id = ?`
)

func (e *Entity) BanUser(ctx context.Context, guildID, userID int64, reason *string) error {
	err := e.c.Session().
		Query(banUser).
		WithContext(ctx).
		Bind(strconv.FormatInt(guildID, 10), strconv.FormatInt(userID, 10), reason).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to ban user: %w", err)
	}
	return nil
}

func (e *Entity) UnbanUser(ctx context.Context, guildID, userID int64) error {
	err := e.c.Session().
		Query(unbanUser).
		WithContext(ctx).
		Bind(strconv.FormatInt(guildID, 10), strconv.FormatInt(userID, 10)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to unban user: %w", err)
	}
	return nil
}

func (e *Entity) IsBanned(ctx context.Context, guildID, userID int64) (bool, error) {
	var count int
	err := e.c.Session().
		Query(isBanned).
		WithContext(ctx).
		Bind(strconv.FormatInt(guildID, 10), strconv.FormatInt(userID, 10)).
		Scan(&count)

	if err != nil {
		return false, fmt.Errorf("unable to check if user is banned: %w", err)
	}
	return count > 0, nil
}

func (e *Entity) GetGuildBans(ctx context.Context, guildID int64) ([]model.GuildBan, error) {
	iter := e.c.Session().
		Query(getGuildBans).
		WithContext(ctx).
		Bind(strconv.FormatInt(guildID, 10)).
		Iter()

	var bans []model.GuildBan
	var userIDRaw string
	var reason *string
	for iter.Scan(&userIDRaw, &reason) {
		userID, err := strconv.ParseInt(userIDRaw, 10, 64)
		if err != nil {
			_ = iter.Close()
			return nil, fmt.Errorf("unable to parse banned user id %q: %w", userIDRaw, err)
		}
		bans = append(bans, model.GuildBan{GuildId: guildID, UserId: userID, Reason: reason})
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("unable to get guild bans: %w", err)
	}
	return bans, nil
}
