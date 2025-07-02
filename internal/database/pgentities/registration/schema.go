package registration

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetRegistrationByUserId(ctx context.Context, userId int64) (model.Registration, error) {
	var r model.Registration
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("registrations").
		Where(squirrel.Eq{"user_id": userId}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.Registration{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &r, sql, args...)
	if err != nil {
		return r, fmt.Errorf("unable to get registration by id: %w", err)
	}
	return r, nil
}
