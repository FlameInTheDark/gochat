package banner

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createBanner = `INSERT INTO gochat.banners (id, user_id, done, filesize) VALUES (?, ?, false, ?) USING TTL ?`
	getBanner    = `SELECT id, user_id, url, content_type, width, height, filesize, done FROM gochat.banners WHERE user_id = ? AND id = ?`
	doneBanner   = `UPDATE gochat.banners USING TTL 0 SET done = true, content_type = ?, url = ?, height = ?, width = ?, filesize = ? WHERE user_id = ? AND id = ?`
	removeBanner = `DELETE FROM gochat.banners WHERE user_id = ? AND id = ?`
)

func (e *Entity) CreateBanner(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error {
	err := e.c.Session().
		Query(createBanner).
		WithContext(ctx).
		Bind(id, userId, fileSize, ttlSeconds).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create banner: %w", err)
	}
	return nil
}

func (e *Entity) GetBanner(ctx context.Context, id, userId int64) (model.Banner, error) {
	var b model.Banner
	err := e.c.Session().
		Query(getBanner).
		WithContext(ctx).
		Bind(userId, id).
		Scan(&b.Id, &b.UserId, &b.URL, &b.ContentType, &b.Width, &b.Height, &b.FileSize, &b.Done)
	if err != nil {
		return b, fmt.Errorf("unable to get banner: %w", err)
	}
	return b, nil
}

func (e *Entity) DoneBanner(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error {
	err := e.c.Session().
		Query(doneBanner).
		WithContext(ctx).
		Bind(contentType, url, height, width, fileSize, userId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to done banner: %w", err)
	}
	return nil
}

func (e *Entity) RemoveBanner(ctx context.Context, id, userId int64) error {
	err := e.c.Session().
		Query(removeBanner).
		WithContext(ctx).
		Bind(userId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove banner: %w", err)
	}
	return nil
}
