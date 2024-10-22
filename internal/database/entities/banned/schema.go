package banned

import (
	"context"
	"fmt"
)

const (
	banUser   = `INSERT INTO gochat.banned (guild_id, user_id) VALUES (?, ?)`
	unbanUser = `DELETE FROM gochat.banned WHERE guild_id = ? AND user_id = ?`
	isBanned  = `SELECT count(user_id) FROM gochat.banned WHERE guild_id = ? AND user_id = ? LIMIT 1`
)

func (e *Entity) BanUser(ctx context.Context, guildID, userID string) error {
	err := e.c.Session().
		Query(banUser).
		WithContext(ctx).
		Bind(guildID, userID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to ban user: %w", err)
	}
	return nil
}

func (e *Entity) UnbanUser(ctx context.Context, guildID, userID string) error {
	err := e.c.Session().
		Query(unbanUser).
		WithContext(ctx).
		Bind(guildID, userID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to unban user: %w", err)
	}
	return nil
}

func (e *Entity) IsBanned(ctx context.Context, guildID, userID string) (bool, error) {
	var count int
	err := e.c.Session().
		Query(isBanned).
		WithContext(ctx).
		Bind(guildID, userID).
		Scan(&count)

	if err != nil {
		return false, fmt.Errorf("unable to check if user is banned: %w", err)
	}
	return count > 0, nil
}
