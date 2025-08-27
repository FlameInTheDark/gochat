package audit

import (
	"context"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Audit interface {
	AddAuditRecord(ctx context.Context, guildID int64, changes map[string]string) error
	RemoveAuditRecordsByGuildId(ctx context.Context, guildID int64) error
	RemoveRecordsBefore(ctx context.Context, before time.Time) error
	GetLastRecords(ctx context.Context, guild int64) ([]model.Audit, error)
	GetRecordsBefore(ctx context.Context, guild int64, before time.Time) ([]model.Audit, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
