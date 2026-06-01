package applicationcommand

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type ApplicationCommand interface {
	CreateCommand(ctx context.Context, cmd appcmd.ApplicationCommand) error
	ReplaceCommand(ctx context.Context, cmd appcmd.ApplicationCommand) error
	GetCommand(ctx context.Context, commandID int64) (appcmd.ApplicationCommand, error)
	GetCommandForBot(ctx context.Context, botUserID, commandID int64) (appcmd.ApplicationCommand, error)
	ListBotCommands(ctx context.Context, botUserID int64, guildID *int64) ([]appcmd.ApplicationCommand, error)
	ListGuildCommandIndex(ctx context.Context, guildID int64) ([]appcmd.ApplicationCommand, error)
	ListVisibleGuildCommands(ctx context.Context, guildID int64, commandType appcmd.CommandType, query string, limit uint64) ([]appcmd.ApplicationCommand, error)
	BulkOverwrite(ctx context.Context, botUserID int64, guildID *int64, commands []appcmd.ApplicationCommand) error
	DeleteCommand(ctx context.Context, botUserID, commandID int64) error

	CreateInteraction(ctx context.Context, interaction appcmd.InteractionRecord) error
	GetInteractionByToken(ctx context.Context, applicationID int64, tokenHash string) (appcmd.InteractionRecord, error)
	AckInteraction(ctx context.Context, interactionID int64, state string, initialResponseID *int64) error
	SetInitialResponse(ctx context.Context, interactionID, initialResponseID int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) ApplicationCommand {
	return &Entity{c: c}
}

type commandRow struct {
	ID                       int64      `db:"id"`
	ApplicationID            int64      `db:"application_id"`
	GuildID                  *int64     `db:"guild_id"`
	Version                  int64      `db:"version"`
	Type                     int        `db:"type"`
	Name                     string     `db:"name"`
	NameLocalizations        string     `db:"name_localizations"`
	Description              string     `db:"description"`
	DescriptionLocalizations string     `db:"description_localizations"`
	Options                  string     `db:"options"`
	DefaultMemberPermissions *int64     `db:"default_member_permissions"`
	Contexts                 string     `db:"contexts"`
	IntegrationTypes         string     `db:"integration_types"`
	NSFW                     bool       `db:"nsfw"`
	BotName                  string     `db:"bot_name"`
	CreatedAt                *time.Time `db:"created_at"`
	UpdatedAt                *time.Time `db:"updated_at"`
}

func (r commandRow) command() appcmd.ApplicationCommand {
	var contexts []appcmd.InteractionContextType
	var integrationTypes []appcmd.IntegrationType
	return appcmd.ApplicationCommand{
		ID:                       r.ID,
		ApplicationID:            r.ApplicationID,
		GuildID:                  r.GuildID,
		Version:                  r.Version,
		Type:                     appcmd.CommandType(r.Type),
		Name:                     r.Name,
		NameLocalizations:        appcmd.UnmarshalJSONField(r.NameLocalizations, map[string]string{}),
		Description:              r.Description,
		DescriptionLocalizations: appcmd.UnmarshalJSONField(r.DescriptionLocalizations, map[string]string{}),
		Options:                  appcmd.UnmarshalJSONField(r.Options, []appcmd.ApplicationCommandOption{}),
		DefaultMemberPermissions: r.DefaultMemberPermissions,
		Contexts:                 appcmd.UnmarshalJSONField(r.Contexts, contexts),
		IntegrationTypes:         appcmd.UnmarshalJSONField(r.IntegrationTypes, integrationTypes),
		NSFW:                     r.NSFW,
		BotName:                  r.BotName,
		CreatedAt:                r.CreatedAt,
		UpdatedAt:                r.UpdatedAt,
	}
}

func commandSelect() squirrel.SelectBuilder {
	return squirrel.Select(
		"ac.id",
		"ac.application_id",
		"ac.guild_id",
		"ac.version",
		"ac.type",
		"ac.name",
		"COALESCE(ac.name_localizations::text, '{}') AS name_localizations",
		"ac.description",
		"COALESCE(ac.description_localizations::text, '{}') AS description_localizations",
		"COALESCE(ac.options::text, '[]') AS options",
		"ac.default_member_permissions",
		"COALESCE(ac.contexts::text, '[]') AS contexts",
		"COALESCE(ac.integration_types::text, '[]') AS integration_types",
		"ac.nsfw",
		"COALESCE(u.name, '') AS bot_name",
		"ac.created_at",
		"ac.updated_at",
	).PlaceholderFormat(squirrel.Dollar).
		From("application_commands ac").
		LeftJoin("users u ON u.id = ac.application_id")
}

func (e *Entity) CreateCommand(ctx context.Context, cmd appcmd.ApplicationCommand) error {
	nameLoc, descLoc, options, contexts, integrationTypes, err := marshalCommandFields(cmd)
	if err != nil {
		return err
	}
	q := squirrel.Insert("application_commands").
		PlaceholderFormat(squirrel.Dollar).
		Columns(
			"id", "application_id", "guild_id", "version", "type", "name",
			"name_localizations", "description", "description_localizations",
			"options", "default_member_permissions", "contexts", "integration_types", "nsfw",
		).
		Values(
			cmd.ID, cmd.ApplicationID, cmd.GuildID, cmd.Version, int(cmd.Type), cmd.Name,
			squirrel.Expr("?::jsonb", nameLoc), cmd.Description, squirrel.Expr("?::jsonb", descLoc),
			squirrel.Expr("?::jsonb", options), cmd.DefaultMemberPermissions, squirrel.Expr("?::jsonb", contexts),
			squirrel.Expr("?::jsonb", integrationTypes), cmd.NSFW,
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build create command SQL: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("create application command: %w", err)
	}
	return nil
}

func (e *Entity) ReplaceCommand(ctx context.Context, cmd appcmd.ApplicationCommand) error {
	nameLoc, descLoc, options, contexts, integrationTypes, err := marshalCommandFields(cmd)
	if err != nil {
		return err
	}
	q := squirrel.Update("application_commands").
		PlaceholderFormat(squirrel.Dollar).
		Set("version", cmd.Version).
		Set("type", int(cmd.Type)).
		Set("name", cmd.Name).
		Set("name_localizations", squirrel.Expr("?::jsonb", nameLoc)).
		Set("description", cmd.Description).
		Set("description_localizations", squirrel.Expr("?::jsonb", descLoc)).
		Set("options", squirrel.Expr("?::jsonb", options)).
		Set("default_member_permissions", cmd.DefaultMemberPermissions).
		Set("contexts", squirrel.Expr("?::jsonb", contexts)).
		Set("integration_types", squirrel.Expr("?::jsonb", integrationTypes)).
		Set("nsfw", cmd.NSFW).
		Set("updated_at", time.Now()).
		Where(squirrel.And{
			squirrel.Eq{"id": cmd.ID},
			squirrel.Eq{"application_id": cmd.ApplicationID},
		})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build replace command SQL: %w", err)
	}
	res, err := e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("replace application command: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (e *Entity) GetCommand(ctx context.Context, commandID int64) (appcmd.ApplicationCommand, error) {
	q := commandSelect().
		Where(squirrel.Eq{"ac.id": commandID}).
		Limit(1)
	return e.getCommand(ctx, q)
}

func (e *Entity) GetCommandForBot(ctx context.Context, botUserID, commandID int64) (appcmd.ApplicationCommand, error) {
	q := commandSelect().
		Where(squirrel.And{squirrel.Eq{"ac.id": commandID}, squirrel.Eq{"ac.application_id": botUserID}}).
		Limit(1)
	return e.getCommand(ctx, q)
}

func (e *Entity) ListBotCommands(ctx context.Context, botUserID int64, guildID *int64) ([]appcmd.ApplicationCommand, error) {
	where := squirrel.And{squirrel.Eq{"ac.application_id": botUserID}}
	if guildID == nil {
		where = append(where, squirrel.Eq{"ac.guild_id": nil})
	} else {
		where = append(where, squirrel.Eq{"ac.guild_id": *guildID})
	}
	q := commandSelect().
		Where(where).
		OrderBy("ac.type ASC", "ac.name ASC")
	return e.listCommands(ctx, q)
}

func (e *Entity) ListGuildCommandIndex(ctx context.Context, guildID int64) ([]appcmd.ApplicationCommand, error) {
	q := commandSelect().
		Join("bot_guilds bg ON bg.bot_user_id = ac.application_id AND bg.guild_id = ?", guildID).
		Join("bots b ON b.bot_user_id = ac.application_id").
		Where(squirrel.And{
			squirrel.Eq{"b.disabled": false},
			squirrel.Or{squirrel.Eq{"ac.guild_id": nil}, squirrel.Eq{"ac.guild_id": guildID}},
		}).
		OrderBy("ac.application_id ASC", "ac.type ASC", "ac.guild_id NULLS FIRST", "ac.name ASC")
	commands, err := e.listCommands(ctx, q)
	if err != nil {
		return nil, err
	}
	return dedupeGuildOverrides(commands), nil
}

func (e *Entity) ListVisibleGuildCommands(ctx context.Context, guildID int64, commandType appcmd.CommandType, query string, limit uint64) ([]appcmd.ApplicationCommand, error) {
	q := commandSelect().
		Join("bot_guilds bg ON bg.bot_user_id = ac.application_id AND bg.guild_id = ?", guildID).
		Join("bots b ON b.bot_user_id = ac.application_id").
		Where(squirrel.And{
			squirrel.Eq{"b.disabled": false},
			squirrel.Eq{"ac.type": int(commandType)},
			squirrel.Or{squirrel.Eq{"ac.guild_id": nil}, squirrel.Eq{"ac.guild_id": guildID}},
		}).
		OrderBy("ac.guild_id NULLS FIRST", "ac.name ASC")
	query = strings.ToLower(strings.TrimSpace(query))
	if query != "" {
		q = q.Where(squirrel.Like{"LOWER(ac.name)": query + "%"})
	}
	if limit == 0 || limit > 100 {
		limit = 100
	}
	q = q.Limit(limit)
	commands, err := e.listCommands(ctx, q)
	if err != nil {
		return nil, err
	}
	return dedupeGuildOverrides(commands), nil
}

func (e *Entity) BulkOverwrite(ctx context.Context, botUserID int64, guildID *int64, commands []appcmd.ApplicationCommand) error {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin command overwrite: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	del := squirrel.Delete("application_commands").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"application_id": botUserID})
	if guildID == nil {
		del = del.Where(squirrel.Eq{"guild_id": nil})
	} else {
		del = del.Where(squirrel.Eq{"guild_id": *guildID})
	}
	raw, args, err := del.ToSql()
	if err != nil {
		return fmt.Errorf("build overwrite delete SQL: %w", err)
	}
	if _, err = tx.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("delete existing commands: %w", err)
	}
	for _, cmd := range commands {
		nameLoc, descLoc, options, contexts, integrationTypes, err := marshalCommandFields(cmd)
		if err != nil {
			return err
		}
		ins := squirrel.Insert("application_commands").
			PlaceholderFormat(squirrel.Dollar).
			Columns("id", "application_id", "guild_id", "version", "type", "name", "name_localizations", "description", "description_localizations", "options", "default_member_permissions", "contexts", "integration_types", "nsfw").
			Values(cmd.ID, botUserID, guildID, cmd.Version, int(cmd.Type), cmd.Name, squirrel.Expr("?::jsonb", nameLoc), cmd.Description, squirrel.Expr("?::jsonb", descLoc), squirrel.Expr("?::jsonb", options), cmd.DefaultMemberPermissions, squirrel.Expr("?::jsonb", contexts), squirrel.Expr("?::jsonb", integrationTypes), cmd.NSFW)
		raw, args, err = ins.ToSql()
		if err != nil {
			return fmt.Errorf("build overwrite insert SQL: %w", err)
		}
		if _, err = tx.ExecContext(ctx, raw, args...); err != nil {
			return fmt.Errorf("insert overwritten command: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit command overwrite: %w", err)
	}
	return nil
}

func (e *Entity) DeleteCommand(ctx context.Context, botUserID, commandID int64) error {
	q := squirrel.Delete("application_commands").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.And{squirrel.Eq{"application_id": botUserID}, squirrel.Eq{"id": commandID}})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build delete command SQL: %w", err)
	}
	res, err := e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("delete application command: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type interactionRow struct {
	ID                int64      `db:"id"`
	ApplicationID     int64      `db:"application_id"`
	CommandID         int64      `db:"command_id"`
	Type              int        `db:"type"`
	GuildID           *int64     `db:"guild_id"`
	ChannelID         int64      `db:"channel_id"`
	InvokerUserID     int64      `db:"invoker_user_id"`
	TokenHash         string     `db:"token_hash"`
	TokenPrefix       string     `db:"token_prefix"`
	AppPermissions    int64      `db:"app_permissions"`
	Locale            string     `db:"locale"`
	GuildLocale       string     `db:"guild_locale"`
	Context           int        `db:"context"`
	AckState          string     `db:"ack_state"`
	InitialResponseID *int64     `db:"initial_response_message_id"`
	ExpiresAt         time.Time  `db:"expires_at"`
	CreatedAt         time.Time  `db:"created_at"`
	RespondedAt       *time.Time `db:"responded_at"`
}

func (r interactionRow) record() appcmd.InteractionRecord {
	return appcmd.InteractionRecord{
		ID:                r.ID,
		ApplicationID:     r.ApplicationID,
		CommandID:         r.CommandID,
		Type:              appcmd.InteractionType(r.Type),
		GuildID:           r.GuildID,
		ChannelID:         r.ChannelID,
		InvokerUserID:     r.InvokerUserID,
		TokenHash:         r.TokenHash,
		TokenPrefix:       r.TokenPrefix,
		AppPermissions:    r.AppPermissions,
		Locale:            r.Locale,
		GuildLocale:       r.GuildLocale,
		Context:           appcmd.InteractionContextType(r.Context),
		AckState:          r.AckState,
		InitialResponseID: r.InitialResponseID,
		ExpiresAt:         r.ExpiresAt,
		CreatedAt:         r.CreatedAt,
		RespondedAt:       r.RespondedAt,
	}
}

func (e *Entity) CreateInteraction(ctx context.Context, interaction appcmd.InteractionRecord) error {
	q := squirrel.Insert("application_command_interactions").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "application_id", "command_id", "type", "guild_id", "channel_id", "invoker_user_id", "token_hash", "token_prefix", "app_permissions", "locale", "guild_locale", "context", "ack_state", "expires_at").
		Values(interaction.ID, interaction.ApplicationID, interaction.CommandID, int(interaction.Type), interaction.GuildID, interaction.ChannelID, interaction.InvokerUserID, interaction.TokenHash, interaction.TokenPrefix, interaction.AppPermissions, interaction.Locale, interaction.GuildLocale, int(interaction.Context), interaction.AckState, interaction.ExpiresAt)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build create interaction SQL: %w", err)
	}
	if _, err = e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("create interaction: %w", err)
	}
	return nil
}

func (e *Entity) GetInteractionByToken(ctx context.Context, applicationID int64, tokenHash string) (appcmd.InteractionRecord, error) {
	var row interactionRow
	q := squirrel.Select("id", "application_id", "command_id", "type", "guild_id", "channel_id", "invoker_user_id", "token_hash", "token_prefix", "app_permissions", "locale", "guild_locale", "context", "ack_state", "initial_response_message_id", "expires_at", "created_at", "responded_at").
		PlaceholderFormat(squirrel.Dollar).
		From("application_command_interactions").
		Where(squirrel.And{squirrel.Eq{"application_id": applicationID}, squirrel.Eq{"token_hash": tokenHash}}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return appcmd.InteractionRecord{}, fmt.Errorf("build get interaction SQL: %w", err)
	}
	if err = e.c.GetContext(ctx, &row, raw, args...); err != nil {
		return appcmd.InteractionRecord{}, fmt.Errorf("get interaction: %w", err)
	}
	return row.record(), nil
}

func (e *Entity) AckInteraction(ctx context.Context, interactionID int64, state string, initialResponseID *int64) error {
	q := squirrel.Update("application_command_interactions").
		PlaceholderFormat(squirrel.Dollar).
		Set("ack_state", state).
		Set("responded_at", time.Now()).
		Set("initial_response_message_id", initialResponseID).
		Where(squirrel.And{
			squirrel.Eq{"id": interactionID},
			squirrel.Eq{"ack_state": appcmd.AckStatePending},
		})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build ack interaction SQL: %w", err)
	}
	res, err := e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("ack interaction: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (e *Entity) SetInitialResponse(ctx context.Context, interactionID, initialResponseID int64) error {
	q := squirrel.Update("application_command_interactions").
		PlaceholderFormat(squirrel.Dollar).
		Set("ack_state", appcmd.AckStateResponded).
		Set("responded_at", time.Now()).
		Set("initial_response_message_id", initialResponseID).
		Where(squirrel.And{
			squirrel.Eq{"id": interactionID},
			squirrel.Eq{"ack_state": appcmd.AckStateDeferred},
			squirrel.Eq{"initial_response_message_id": nil},
		})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build set initial response SQL: %w", err)
	}
	res, err := e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("set initial response: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (e *Entity) getCommand(ctx context.Context, q squirrel.SelectBuilder) (appcmd.ApplicationCommand, error) {
	var row commandRow
	raw, args, err := q.ToSql()
	if err != nil {
		return appcmd.ApplicationCommand{}, fmt.Errorf("build get command SQL: %w", err)
	}
	if err = e.c.GetContext(ctx, &row, raw, args...); err != nil {
		return appcmd.ApplicationCommand{}, fmt.Errorf("get application command: %w", err)
	}
	return row.command(), nil
}

func (e *Entity) listCommands(ctx context.Context, q squirrel.SelectBuilder) ([]appcmd.ApplicationCommand, error) {
	var rows []commandRow
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list commands SQL: %w", err)
	}
	if err = e.c.SelectContext(ctx, &rows, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("list application commands: %w", err)
	}
	out := make([]appcmd.ApplicationCommand, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.command())
	}
	return out, nil
}

func marshalCommandFields(cmd appcmd.ApplicationCommand) (nameLoc, descLoc, options, contexts, integrationTypes string, err error) {
	if nameLoc, err = appcmd.MarshalJSONField(emptyMap(cmd.NameLocalizations)); err != nil {
		return
	}
	if descLoc, err = appcmd.MarshalJSONField(emptyMap(cmd.DescriptionLocalizations)); err != nil {
		return
	}
	if options, err = appcmd.MarshalJSONField(emptySlice(cmd.Options)); err != nil {
		return
	}
	if contexts, err = appcmd.MarshalJSONField(emptySlice(cmd.Contexts)); err != nil {
		return
	}
	if integrationTypes, err = appcmd.MarshalJSONField(emptySlice(cmd.IntegrationTypes)); err != nil {
		return
	}
	return
}

func emptyMap(values map[string]string) map[string]string {
	if values == nil {
		return map[string]string{}
	}
	return values
}

func emptySlice[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}

func dedupeGuildOverrides(commands []appcmd.ApplicationCommand) []appcmd.ApplicationCommand {
	seen := make(map[string]int, len(commands))
	out := make([]appcmd.ApplicationCommand, 0, len(commands))
	for _, command := range commands {
		key := fmt.Sprintf("%d:%d:%s", command.ApplicationID, command.Type, command.Name)
		if idx, ok := seen[key]; ok {
			if command.GuildID != nil {
				out[idx] = command
			}
			continue
		}
		seen[key] = len(out)
		out = append(out, command)
	}
	return out
}
