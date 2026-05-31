package developer

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

func (e *entity) publishBotSearchUpsert(ctx context.Context, botUserID int64) {
	if e.smq == nil {
		return
	}
	if err := e.smq.UpsertBot(ctx, dto.BotIndexMessage{BotUserId: botUserID}); err != nil && e.log != nil {
		e.log.Error("unable to publish bot search upsert", slog.Int64("bot_user_id", botUserID), slog.String("error", err.Error()))
	}
}

func (e *entity) publishBotSearchDelete(ctx context.Context, botUserID int64) {
	if e.smq == nil {
		return
	}
	if err := e.smq.DeleteBot(ctx, dto.BotIndexDeleteMessage{BotUserId: botUserID}); err != nil && e.log != nil {
		e.log.Error("unable to publish bot search delete", slog.Int64("bot_user_id", botUserID), slog.String("error", err.Error()))
	}
}
