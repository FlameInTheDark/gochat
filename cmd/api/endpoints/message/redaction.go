package message

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

func (e *entity) redactBannedMessages(ctx context.Context, guildId int64, rawMessages []model.Message, messages []dto.Message) error {
	if e.ban == nil || len(rawMessages) == 0 || len(messages) == 0 {
		return nil
	}

	bannedAuthors := make(map[int64]bool, len(rawMessages))
	for _, message := range rawMessages {
		if _, seen := bannedAuthors[message.UserId]; seen {
			continue
		}
		banned, err := e.ban.IsBanned(ctx, guildId, message.UserId)
		if err != nil {
			return err
		}
		bannedAuthors[message.UserId] = banned
	}

	for i, raw := range rawMessages {
		if !bannedAuthors[raw.UserId] {
			continue
		}
		messages[i].Content = ""
		messages[i].Attachments = nil
		messages[i].Embeds = nil
		messages[i].Flags |= model.MessageFlagBannedAuthor
	}

	return nil
}
