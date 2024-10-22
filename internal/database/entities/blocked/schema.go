package blocked

import (
	"context"
	"fmt"
)

const (
	blockUser   = `INSERT INTO gochat.blocked (user_id, blocked_id) VALUES (?, ?)`
	unblockUser = `DELETE FROM gochat.blocked WHERE user_id = ? AND blocked_id = ?`
	isBlocked   = `SELECT count(blocked_id) FROM gochat.blocked WHERE user_id = ? AND blocked_id = ? LIMIT 1`
)

func (e *Entity) BlockUser(ctx context.Context, guildID, userID string) error {
	err := e.c.Session().
		Query(blockUser).
		WithContext(ctx).
		Bind(guildID, userID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to ban user: %w", err)
	}
	return nil
}

func (e *Entity) UnblockUser(ctx context.Context, guildID, userID string) error {
	err := e.c.Session().
		Query(unblockUser).
		WithContext(ctx).
		Bind(guildID, userID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to unban user: %w", err)
	}
	return nil
}

func (e *Entity) IsBlocked(ctx context.Context, guildID, userID string) (bool, error) {
	var count int
	err := e.c.Session().
		Query(isBlocked).
		WithContext(ctx).
		Bind(guildID, userID).
		Scan(&count)

	if err != nil {
		return false, fmt.Errorf("unable to check if user is banned: %w", err)
	}
	return count > 0, nil
}
