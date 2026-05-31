package bot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) CreateBot(ctx context.Context, bot model.Bot) error {
	q := squirrel.Insert("bots").
		PlaceholderFormat(squirrel.Dollar).
		Columns("bot_user_id", "owner_user_id", "description", "public", "default_permissions", "disabled").
		Values(bot.BotUserId, bot.OwnerUserId, bot.Description, bot.Public, bot.DefaultPermissions, bot.Disabled)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to create bot: %w", err)
	}
	return nil
}

func (e *Entity) GetBot(ctx context.Context, botUserID int64) (model.Bot, error) {
	var bot model.Bot
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bots").
		Where(squirrel.Eq{"bot_user_id": botUserID}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return bot, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.GetContext(ctx, &bot, raw, args...); err != nil {
		return bot, fmt.Errorf("unable to get bot: %w", err)
	}
	return bot, nil
}

func (e *Entity) GetBotForOwner(ctx context.Context, ownerUserID, botUserID int64) (model.Bot, error) {
	var bot model.Bot
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bots").
		Where(squirrel.And{squirrel.Eq{"owner_user_id": ownerUserID}, squirrel.Eq{"bot_user_id": botUserID}}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return bot, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.GetContext(ctx, &bot, raw, args...); err != nil {
		return bot, fmt.Errorf("unable to get owner bot: %w", err)
	}
	return bot, nil
}

func (e *Entity) ListOwnerBots(ctx context.Context, ownerUserID int64) ([]model.Bot, error) {
	var bots []model.Bot
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bots").
		Where(squirrel.Eq{"owner_user_id": ownerUserID}).
		OrderBy("created_at DESC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &bots, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to list owner bots: %w", err)
	}
	return bots, nil
}

func (e *Entity) UpdateBot(ctx context.Context, botUserID int64, description *string, public *bool, defaultPermissions *int64, disabled *bool) error {
	q := squirrel.Update("bots").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"bot_user_id": botUserID}).
		Set("updated_at", time.Now())
	if description != nil {
		q = q.Set("description", *description)
	}
	if public != nil {
		q = q.Set("public", *public)
	}
	if defaultPermissions != nil {
		q = q.Set("default_permissions", *defaultPermissions)
	}
	if disabled != nil {
		q = q.Set("disabled", *disabled)
	}
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to update bot: %w", err)
	}
	return nil
}

func (e *Entity) DeleteBot(ctx context.Context, botUserID int64) error {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin bot delete transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, "DELETE FROM members WHERE user_id = $1", botUserID); err != nil {
		return fmt.Errorf("unable to delete bot memberships: %w", err)
	}
	for _, table := range []string{"bot_guilds", "bot_install_grants", "bot_tokens", "bots"} {
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE bot_user_id = $1", botUserID); err != nil {
			return fmt.Errorf("unable to delete %s: %w", table, err)
		}
	}
	return tx.Commit()
}

func (e *Entity) SearchPublicBots(ctx context.Context, query string, limit, offset uint64) ([]model.Bot, error) {
	var bots []model.Bot
	q := squirrel.Select("b.*").
		PlaceholderFormat(squirrel.Dollar).
		From("bots b").
		Join("users u ON u.id = b.bot_user_id").
		Where(squirrel.And{squirrel.Eq{"b.public": true}, squirrel.Eq{"b.disabled": false}}).
		OrderBy("b.created_at DESC").
		Limit(limit).
		Offset(offset)
	query = strings.TrimSpace(query)
	if query != "" {
		q = q.Where(squirrel.Or{
			squirrel.Like{"LOWER(u.name)": "%" + strings.ToLower(query) + "%"},
			squirrel.Like{"LOWER(b.description)": "%" + strings.ToLower(query) + "%"},
		})
	}
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &bots, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to search public bots: %w", err)
	}
	return bots, nil
}

func (e *Entity) CreateToken(ctx context.Context, token model.BotToken) error {
	q := squirrel.Insert("bot_tokens").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "bot_user_id", "name", "token_hash", "token_prefix").
		Values(token.Id, token.BotUserId, token.Name, token.TokenHash, token.TokenPrefix)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to create bot token: %w", err)
	}
	return nil
}

func (e *Entity) GetTokenByHash(ctx context.Context, tokenHash string) (model.BotToken, error) {
	var token model.BotToken
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_tokens").
		Where(squirrel.And{squirrel.Eq{"token_hash": tokenHash}, squirrel.Eq{"revoked_at": nil}}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return token, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.GetContext(ctx, &token, raw, args...); err != nil {
		return token, fmt.Errorf("unable to get bot token: %w", err)
	}
	return token, nil
}

func (e *Entity) ListTokens(ctx context.Context, botUserID int64) ([]model.BotToken, error) {
	var tokens []model.BotToken
	q := squirrel.Select("id", "bot_user_id", "name", "token_hash", "token_prefix", "revoked_at", "last_used_at", "created_at").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_tokens").
		Where(squirrel.Eq{"bot_user_id": botUserID}).
		OrderBy("created_at DESC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &tokens, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to list bot tokens: %w", err)
	}
	return tokens, nil
}

func (e *Entity) RevokeToken(ctx context.Context, botUserID, tokenID int64) error {
	q := squirrel.Update("bot_tokens").
		PlaceholderFormat(squirrel.Dollar).
		Set("revoked_at", time.Now()).
		Where(squirrel.And{squirrel.Eq{"bot_user_id": botUserID}, squirrel.Eq{"id": tokenID}})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to revoke bot token: %w", err)
	}
	return nil
}

func (e *Entity) TouchToken(ctx context.Context, tokenID int64) error {
	q := squirrel.Update("bot_tokens").
		PlaceholderFormat(squirrel.Dollar).
		Set("last_used_at", time.Now()).
		Where(squirrel.Eq{"id": tokenID})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to touch bot token: %w", err)
	}
	return nil
}

func (e *Entity) CreateGrant(ctx context.Context, grant model.BotInstallGrant) error {
	q := squirrel.Insert("bot_install_grants").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "bot_user_id", "owner_user_id", "token_hash", "token_prefix", "requested_permissions", "expires_at", "max_uses").
		Values(grant.Id, grant.BotUserId, grant.OwnerUserId, grant.TokenHash, grant.TokenPrefix, grant.RequestedPermissions, grant.ExpiresAt, grant.MaxUses)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to create bot install grant: %w", err)
	}
	return nil
}

func (e *Entity) GetGrantByHash(ctx context.Context, tokenHash string) (model.BotInstallGrant, error) {
	var grant model.BotInstallGrant
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_install_grants").
		Where(squirrel.Eq{"token_hash": tokenHash}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return grant, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.GetContext(ctx, &grant, raw, args...); err != nil {
		return grant, fmt.Errorf("unable to get bot install grant: %w", err)
	}
	return grant, nil
}

func (e *Entity) ListGrants(ctx context.Context, botUserID int64) ([]model.BotInstallGrant, error) {
	var grants []model.BotInstallGrant
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_install_grants").
		Where(squirrel.Eq{"bot_user_id": botUserID}).
		OrderBy("created_at DESC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &grants, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to list bot install grants: %w", err)
	}
	return grants, nil
}

func (e *Entity) RevokeGrant(ctx context.Context, botUserID, grantID int64) error {
	q := squirrel.Update("bot_install_grants").
		PlaceholderFormat(squirrel.Dollar).
		Set("revoked_at", time.Now()).
		Where(squirrel.And{squirrel.Eq{"bot_user_id": botUserID}, squirrel.Eq{"id": grantID}})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to revoke bot install grant: %w", err)
	}
	return nil
}

func (e *Entity) UseGrant(ctx context.Context, grantID int64) error {
	q := squirrel.Update("bot_install_grants").
		PlaceholderFormat(squirrel.Dollar).
		Set("uses", squirrel.Expr("uses + 1")).
		Where(squirrel.Eq{"id": grantID})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to use bot install grant: %w", err)
	}
	return nil
}

func (e *Entity) UpsertBotGuild(ctx context.Context, guild model.BotGuild) error {
	raw := `INSERT INTO bot_guilds (bot_user_id, guild_id, granted_permissions, installer_user_id, grant_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (bot_user_id, guild_id) DO UPDATE SET
    granted_permissions = EXCLUDED.granted_permissions,
    installer_user_id = EXCLUDED.installer_user_id,
    grant_id = EXCLUDED.grant_id`
	if _, err := e.c.ExecContext(ctx, raw, guild.BotUserId, guild.GuildId, guild.GrantedPermissions, guild.InstallerUserId, guild.GrantId); err != nil {
		return fmt.Errorf("unable to upsert bot guild: %w", err)
	}
	return nil
}

func (e *Entity) GetBotGuild(ctx context.Context, botUserID, guildID int64) (model.BotGuild, error) {
	var guild model.BotGuild
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_guilds").
		Where(squirrel.And{squirrel.Eq{"bot_user_id": botUserID}, squirrel.Eq{"guild_id": guildID}}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return guild, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.GetContext(ctx, &guild, raw, args...); err != nil {
		return guild, fmt.Errorf("unable to get bot guild: %w", err)
	}
	return guild, nil
}

func (e *Entity) ListBotGuilds(ctx context.Context, botUserID int64) ([]model.BotGuild, error) {
	var guilds []model.BotGuild
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_guilds").
		Where(squirrel.Eq{"bot_user_id": botUserID}).
		OrderBy("guild_id ASC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &guilds, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to list bot guilds: %w", err)
	}
	return guilds, nil
}

func (e *Entity) ListGuildBots(ctx context.Context, guildID int64) ([]model.BotGuild, error) {
	var bots []model.BotGuild
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("bot_guilds").
		Where(squirrel.Eq{"guild_id": guildID}).
		OrderBy("created_at DESC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &bots, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to list guild bots: %w", err)
	}
	return bots, nil
}

func (e *Entity) DeleteBotGuild(ctx context.Context, botUserID, guildID int64) error {
	q := squirrel.Delete("bot_guilds").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.And{squirrel.Eq{"bot_user_id": botUserID}, squirrel.Eq{"guild_id": guildID}})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to delete bot guild: %w", err)
	}
	return nil
}
