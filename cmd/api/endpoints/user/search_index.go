package user

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

func (e *entity) publishGuildSearchUpsert(ctx context.Context, guildID int64, logger *slog.Logger) {
	if e.smq == nil {
		return
	}
	if err := e.smq.UpsertGuild(ctx, dto.GuildIndexMessage{GuildId: guildID}); err != nil && logger != nil {
		logger.Error("unable to publish guild search upsert", slog.Int64("guild_id", guildID), slog.String("error", err.Error()))
	}
}
