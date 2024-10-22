package user

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getUserByID       = `SELECT id, name, determinator, avatar, blocked, created_at FROM gochat.users WHERE id = ?;`
	getByDeterminator = `SELECT id, name, determinator, avatar, blocked, created_at FROM gochat.users WHERE determinator = ?;`
	createUser        = `INSERT INTO gochat.users (id, determinator,  name, blocked, created_at) VALUES (?, ?, ?, false, toTimestamp(now()));`
	setAvatar         = `UPDATE gochat.users SET avatar = ? WHERE id = ?;`
	setUsername       = `UPDATE gochat.users SET name = ? WHERE id = ?;`
	setBlocked        = `UPDATE gochat.users SET blocked = ? WHERE id = ?;`
)

func (e *Entity) GetUserById(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := e.c.Session().
		Query(getUserByID).
		WithContext(ctx).
		Bind(id).
		Scan(
			&user.Id,
			&user.Name,
			&user.Determinator,
			&user.Avatar,
			&user.Blocked,
			&user.CreatedAt,
		)
	if err != nil {
		return user, fmt.Errorf("unable to get user: %w", err)
	}
	return user, nil
}

func (e *Entity) GetUserByDeterminator(ctx context.Context, determinator string) (model.User, error) {
	var u model.User
	err := e.c.Session().
		Query(getByDeterminator).
		WithContext(ctx).
		Bind(determinator).
		Scan(&u.Id, &u.Name, &u.Determinator, &u.Avatar, &u.Blocked, &u.CreatedAt)
	if err != nil {
		return u, fmt.Errorf("unable to get user by determinator: %w", err)
	}
	return u, nil
}

func (e *Entity) CreateUser(ctx context.Context, id int64, name, determinator string) error {
	err := e.c.Session().
		Query(createUser).
		WithContext(ctx).
		Bind(id, determinator, name).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}
	return nil
}

func (e *Entity) SetUserAvatar(ctx context.Context, id, attachmentId int64) error {
	err := e.c.Session().
		Query(setAvatar).
		WithContext(ctx).
		Bind(attachmentId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set avatar Error: %w", err)
	}
	return nil
}

func (e *Entity) SetUsername(ctx context.Context, id, name string) error {
	err := e.c.Session().
		Query(setUsername).
		WithContext(ctx).
		Bind(name, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set username Error: %w", err)
	}
	return nil
}

func (e *Entity) SetUserBlocked(ctx context.Context, id int64, blocked bool) error {
	err := e.c.Session().
		Query(setBlocked).
		WithContext(ctx).
		Bind(id, blocked).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set blocked Error: %w", err)
	}
	return nil
}
