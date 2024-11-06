package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getUserByID    = `SELECT id, name, avatar, blocked, upload_limit, created_at FROM gochat.users WHERE id = ?;`
	createUser     = `INSERT INTO gochat.users (id,  name, blocked, created_at) VALUES (?, ?, false, toTimestamp(now()));`
	setAvatar      = `UPDATE gochat.users SET avatar = ? WHERE id = ?;`
	setUsername    = `UPDATE gochat.users SET name = ? WHERE id = ?;`
	setBlocked     = `UPDATE gochat.users SET blocked = ? WHERE id = ?;`
	updateUser     = `UPDATE gochat.users SET %s WHERE id = ?`
	setUploadLimit = `UPDATE gochat.users SET upload_limit = ? WHERE id = ?`
)

func (e *Entity) ModifyUser(ctx context.Context, userId int64, name *string, avatar *int64) error {
	var arg []any
	var params []string
	if avatar != nil {
		arg = append(arg, *avatar)
		params = append(params, "avatar = ?")
	}
	if name != nil {
		arg = append(arg, *name)
		params = append(params, "name = ?")
	}
	arg = append(arg, userId)
	err := e.c.Session().
		Query(fmt.Sprintf(updateUser, strings.Join(params, ","))).
		WithContext(ctx).
		Bind(arg...).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to modify user: %w", err)
	}
	return nil
}

func (e *Entity) GetUserById(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := e.c.Session().
		Query(getUserByID).
		WithContext(ctx).
		Bind(id).
		Scan(
			&user.Id,
			&user.Name,
			&user.Avatar,
			&user.Blocked,
			&user.UploadLimit,
			&user.CreatedAt,
		)
	if err != nil {
		return user, fmt.Errorf("unable to get user: %w", err)
	}
	return user, nil
}

func (e *Entity) CreateUser(ctx context.Context, id int64, name string) error {
	err := e.c.Session().
		Query(createUser).
		WithContext(ctx).
		Bind(id, name).
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
		Bind(blocked, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set blocked Error: %w", err)
	}
	return nil
}

func (e *Entity) SetUploadLimit(ctx context.Context, id int64, limit int64) error {
	err := e.c.Session().
		Query(setUploadLimit).
		WithContext(ctx).
		Bind(limit, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set upload limit Error: %w", err)
	}
	return nil
}
