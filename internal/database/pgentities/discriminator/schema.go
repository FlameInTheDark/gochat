package discriminator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) CreateDiscriminator(ctx context.Context, userId int64, discriminator string) error {
	q := squirrel.Insert("discriminators").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "discriminator").
		Values(userId, discriminator)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to create discriminator: %w", err)
	}
	return nil
}

func (e *Entity) GetDiscriminatorByUserId(ctx context.Context, userId int64) (model.Discriminator, error) {
	var disc model.Discriminator
	q := squirrel.Select("user_id", "discriminator").
		PlaceholderFormat(squirrel.Dollar).
		From("discriminators").
		Where(squirrel.Eq{"user_id": userId}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return model.Discriminator{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &disc, raw, args...)
	if err != nil {
		return model.Discriminator{}, fmt.Errorf("unable to get discriminator by user id: %w", err)
	}
	return disc, nil
}

func (e *Entity) GetUserIdByDiscriminator(ctx context.Context, discriminator string) (model.Discriminator, error) {
	var disc model.Discriminator
	q := squirrel.Select("user_id", "discriminator").
		PlaceholderFormat(squirrel.Dollar).
		From("discriminators").
		Where(squirrel.Eq{"discriminator": discriminator}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return model.Discriminator{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &disc, raw, args...)
	if err != nil {
		return model.Discriminator{}, fmt.Errorf("unable to get discriminator by user id: %w", err)
	}
	return disc, nil
}

func (e *Entity) GetDiscriminatorsByUserIDs(ctx context.Context, userIDs []int64) ([]model.Discriminator, error) {
	var discs []model.Discriminator
	q := squirrel.Select("user_id", "discriminator").
		PlaceholderFormat(squirrel.Dollar).
		From("discriminators").
		Where(squirrel.Eq{"user_id": userIDs}).
		OrderBy("user_id ASC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &discs, raw, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get discriminators by user ids: %w", err)
	}
	return discs, nil
}
