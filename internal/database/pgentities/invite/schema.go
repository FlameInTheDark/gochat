package invite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

// CreateInvite inserts records into guild_invites and guild_invite_codes in a transaction.
// expiresAt must be a unix seconds timestamp, will be converted to timestamptz.
func (e *Entity) CreateInvite(ctx context.Context, code string, inviteID, guildID, authorID int64, expiresAt int64) (model.GuildInvite, error) {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return model.GuildInvite{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// Insert into guild_invites first
	// Using squirrel to avoid SQL injection and keep consistency
	q1 := squirrel.Insert("guild_invites").
		PlaceholderFormat(squirrel.Dollar).
		Columns("guild_id", "invite_id", "author_id", "created_at", "expires_at").
		Values(guildID, inviteID, authorID, time.Now(), time.Unix(expiresAt, 0))
	sql1, args1, err := q1.ToSql()
	if err != nil {
		return model.GuildInvite{}, fmt.Errorf("unable to create SQL query for invites: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sql1, args1...); err != nil {
		return model.GuildInvite{}, fmt.Errorf("unable to insert invite: %w", err)
	}

	// Insert into mapper table with the provided code
	q2 := squirrel.Insert("guild_invite_codes").
		PlaceholderFormat(squirrel.Dollar).
		Columns("invite_code", "invite_id", "guild_id").
		Values(code, inviteID, guildID)
	sql2, args2, err := q2.ToSql()
	if err != nil {
		return model.GuildInvite{}, fmt.Errorf("unable to create SQL query for invite codes: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sql2, args2...); err != nil {
		return model.GuildInvite{}, fmt.Errorf("unable to insert invite code: %w", err)
	}

	return model.GuildInvite{
		InviteCode: code,
		InviteId:   inviteID,
		GuildId:    guildID,
		AuthorId:   authorID,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Unix(expiresAt, 0),
	}, nil
}

// GetGuildInvites returns only non-expired invites using active_guild_invites view
func (e *Entity) GetGuildInvites(ctx context.Context, guildID int64) ([]model.GuildInvite, error) {
	var out []model.GuildInvite

	// Join active invites with their codes
	q := squirrel.Select(
		"ic.invite_code",
		"gi.invite_id",
		"gi.guild_id",
		"gi.author_id",
		"gi.created_at",
		"gi.expires_at",
	).
		PlaceholderFormat(squirrel.Dollar).
		From("active_guild_invites gi").
		Join("guild_invite_codes ic ON ic.invite_id = gi.invite_id AND ic.guild_id = gi.guild_id").
		Where(squirrel.Eq{"gi.guild_id": guildID}).
		OrderBy("gi.created_at DESC")

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to build SQL for list invites: %w", err)
	}
	if err = e.c.SelectContext(ctx, &out, sqlStr, args...); err != nil {
		return nil, fmt.Errorf("unable to get guild invites: %w", err)
	}
	return out, nil
}

// DeleteInviteByCode resolves code and deletes via helper function
func (e *Entity) DeleteInviteByCode(ctx context.Context, guildID int64, code string) error {
	var row struct {
		InviteId sql.NullInt64 `db:"invite_id"`
		GuildId  sql.NullInt64 `db:"guild_id"`
	}

	// Resolve code in mapper
	q := squirrel.Select("invite_id", "guild_id").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_invite_codes").
		Where(squirrel.Eq{"invite_code": code}).
		Limit(1)
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to build SQL for resolve invite code: %w", err)
	}
	if err = e.c.GetContext(ctx, &row, sqlStr, args...); err != nil {
		return fmt.Errorf("unable to resolve invite code: %w", err)
	}
	if !row.InviteId.Valid || !row.GuildId.Valid {
		return sql.ErrNoRows
	}
	// Ensure the code belongs to provided guild
	if row.GuildId.Int64 != guildID {
		return sql.ErrNoRows
	}

	// Use the helper function to delete both rows
	// SELECT delete_guild_invite($1, $2)
	if _, err = e.c.ExecContext(ctx, "SELECT delete_guild_invite($1, $2)", row.GuildId.Int64, row.InviteId.Int64); err != nil {
		return fmt.Errorf("unable to delete invite via function: %w", err)
	}
	return nil
}

// FetchInvite calls function fetch_guild_invite to return a valid invite (or 0 rows)
func (e *Entity) FetchInvite(ctx context.Context, code string) (model.GuildInvite, error) {
	var inv model.GuildInvite
	// The function returns: invite_code, invite_id, guild_id, author_id, created_at, expires_at
	err := e.c.GetContext(ctx, &inv, "SELECT * FROM fetch_guild_invite($1)", code)
	if err != nil {
		return model.GuildInvite{}, err
	}
	return inv, nil
}

// DeleteInviteByID removes the invite by composite key using the helper function
func (e *Entity) DeleteInviteByID(ctx context.Context, guildID, inviteID int64) error {
	if _, err := e.c.ExecContext(ctx, "SELECT delete_guild_invite($1, $2)", guildID, inviteID); err != nil {
		return fmt.Errorf("unable to delete invite via function: %w", err)
	}
	return nil
}
