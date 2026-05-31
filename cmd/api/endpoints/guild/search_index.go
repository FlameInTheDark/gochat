package guild

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

func (e *entity) adjustDiscoveryMembers(ctx context.Context, guildID, delta int64) {
	if e.gd == nil {
		return
	}
	if err := e.gd.AdjustMembersCount(ctx, guildID, delta); err != nil {
		_ = e.gd.RecountMembers(ctx, guildID)
	}
}

func (e *entity) publishGuildSearchUpsert(ctx context.Context, guildID int64, logger *slog.Logger) {
	if e.smq == nil {
		return
	}
	if err := e.smq.UpsertGuild(ctx, dto.GuildIndexMessage{GuildId: guildID}); err != nil && logger != nil {
		logger.Error("unable to publish guild search upsert", slog.Int64("guild_id", guildID), slog.String("error", err.Error()))
	}
}

func (e *entity) publishGuildSearchDelete(ctx context.Context, guildID int64, logger *slog.Logger) {
	if e.smq == nil {
		return
	}
	if err := e.smq.DeleteGuild(ctx, dto.GuildIndexDeleteMessage{GuildId: guildID}); err != nil && logger != nil {
		logger.Error("unable to publish guild search delete", slog.Int64("guild_id", guildID), slog.String("error", err.Error()))
	}
}

func (e *entity) publishBotSearchUpsert(ctx context.Context, botUserID int64, logger *slog.Logger) {
	if e.smq == nil {
		return
	}
	if err := e.smq.UpsertBot(ctx, dto.BotIndexMessage{BotUserId: botUserID}); err != nil && logger != nil {
		logger.Error("unable to publish bot search upsert", slog.Int64("bot_user_id", botUserID), slog.String("error", err.Error()))
	}
}
