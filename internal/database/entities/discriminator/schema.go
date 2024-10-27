package discriminator

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createDiscriminator      = `INSERT INTO gochat.discriminators (user_id, discriminator) VALUES (?, ?)`
	getDiscriminatorByUserId = `SELECT user_id, discriminator FROM gochat.discriminators WHERE user_id = ?`
	getUserIdByDiscriminator = `SELECT user_id, discriminator FROM gochat.discriminators WHERE discriminator = ?`
)

func (e *Entity) CreateDiscriminator(ctx context.Context, userId int64, discriminator string) error {
	err := e.c.Session().
		Query(createDiscriminator).
		WithContext(ctx).
		Bind(userId, discriminator).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create discriminator: %w", err)
	}
	return nil
}

func (e *Entity) GetDiscriminatorByUserId(ctx context.Context, userId int64) (model.Discriminator, error) {
	var disc model.Discriminator
	err := e.c.Session().
		Query(getDiscriminatorByUserId).
		WithContext(ctx).
		Bind(userId).
		Scan(&disc.UserId, &disc.Discriminator)
	if err != nil {
		return model.Discriminator{}, fmt.Errorf("unable to get discriminator by user id: %w", err)
	}
	return disc, nil
}

func (e *Entity) GetUserIdByDiscriminator(ctx context.Context, discriminator string) (model.Discriminator, error) {
	var disc model.Discriminator
	err := e.c.Session().
		Query(getUserIdByDiscriminator).
		WithContext(ctx).
		Bind(discriminator).
		Scan(&disc.UserId, &disc.Discriminator)
	if err != nil {
		return model.Discriminator{}, fmt.Errorf("unable to get discriminator by user id: %w", err)
	}
	return disc, nil
}
