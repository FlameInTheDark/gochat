package user

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) ModifyUser(ctx context.Context, userId int64, name *string, avatar *int64) error {
	q := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": userId})
	if name != nil {
		q = q.Set("name", *name)
	}
	if avatar != nil {
		q = q.Set("avatar", *avatar)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to modify user: %w", err)
	}
	return nil
}

func (e *Entity) GetUserById(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("users").
		Where(squirrel.Eq{"id": id}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return user, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &user, sql, args...)
	if err != nil {
		return user, fmt.Errorf("unable to get user: %w", err)
	}
	return user, nil
}

func (e *Entity) GetUsersList(ctx context.Context, ids []int64) ([]model.User, error) {
	var users []model.User
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("users").
		Where(squirrel.Eq{"id": ids})
	sql, args, err := q.ToSql()
	if err != nil {
		return users, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &users, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get users: %w", err)
	}
	return users, nil
}

func (e *Entity) CreateUser(ctx context.Context, id int64, name string) error {
	q := squirrel.Insert("users").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "name", "blocked").
		Values(id, name, false)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}
	return nil
}

func (e *Entity) SetUserAvatar(ctx context.Context, id, attachmentId int64) error {
	q := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("avatar", attachmentId)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set avatar Error: %w", err)
	}
	return nil
}

func (e *Entity) SetUsername(ctx context.Context, id, name string) error {
	q := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("name", name)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set username Error: %w", err)
	}
	return nil
}

func (e *Entity) SetUserBlocked(ctx context.Context, id int64, blocked bool) error {
	q := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("blocked", blocked)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set blocked Error: %w", err)
	}
	return nil
}

func (e *Entity) SetUploadLimit(ctx context.Context, id int64, uploadLimit int64) error {
	q := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("upload_limit", uploadLimit)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set upload limit Error: %w", err)
	}
	return nil
}
