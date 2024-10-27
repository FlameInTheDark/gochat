package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	addAuditRecord         = `INSERT INTO gochat.audit_log (guild_id, created_at, changes) VALUES (?, toTimestamp(now()), ?)`
	removeRecordsByGuildId = `DELETE FROM gochat.audit_log WHERE guild_id = ?`
	removeRecordsBefore    = `DELETE FROM gochat.audit_log WHERE created_at < ?`
	getLastRecords         = `SELECT guild_id, created_at, changes FROM gochat.audit_log WHERE guild_id = ? ORDER BY created_at DESC LIMIT 10`
	getRecordsBefore       = `SELECT guild_id, created_at, changes FROM gochat.audit_log WHERE guild_id = ? AND created_at < ? ORDER BY created_at DESC LIMIT 10`
)

func (e *Entity) AddAuditRecord(ctx context.Context, guildID int64, changes map[string]string) error {
	err := e.c.Session().
		Query(addAuditRecord).
		WithContext(ctx).
		Bind(guildID, time.Now(), changes).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add audit record: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAuditRecordsByGuildId(ctx context.Context, guildID int64) error {
	err := e.c.Session().
		Query(removeRecordsByGuildId).
		WithContext(ctx).
		Bind(guildID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove audit records: %w", err)
	}
	return nil
}

func (e *Entity) RemoveRecordsBefore(ctx context.Context, before time.Time) error {
	err := e.c.Session().
		Query(removeRecordsBefore).
		WithContext(ctx).
		Bind(before).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove records: %w", err)
	}
	return nil
}

func (e *Entity) GetLastRecords(ctx context.Context, guild int64) ([]model.Audit, error) {
	var audits []model.Audit
	iter := e.c.Session().
		Query(getLastRecords).
		WithContext(ctx).
		Bind(guild).
		Iter()
	var a model.Audit
	for iter.Scan(&a.GuildId, &a.CreatedAt, &a.Changes) {
		audits = append(audits, a)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get audit records: %w", err)
	}
	return audits, nil
}

func (e *Entity) GetRecordsBefore(ctx context.Context, guild int64, before time.Time) ([]model.Audit, error) {
	var audits []model.Audit
	iter := e.c.Session().
		Query(getRecordsBefore).
		WithContext(ctx).
		Bind(guild).
		Iter()
	var a model.Audit
	for iter.Scan(&a.GuildId, &a.CreatedAt, &a.Changes) {
		audits = append(audits, a)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get audit records: %w", err)
	}
	return audits, nil
}
