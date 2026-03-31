package wsmsg

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	reactionutil "github.com/FlameInTheDark/gochat/internal/reaction"
)

type ReactionLookupCache interface {
	HGet(ctx context.Context, key, field string) (string, error)
}

func PersonalizeMessageForRecipient(topic string, userID int64, data []byte) []byte {
	return PersonalizeMessageForRecipientWithCache(nil, topic, userID, data)
}

func PersonalizeMessageForRecipientWithCache(cache ReactionLookupCache, topic string, userID int64, data []byte) []byte {
	if !strings.HasPrefix(topic, "channel.") || userID == 0 {
		return cloneMessage(data)
	}

	var envelope mqmsg.Message
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.EventType == nil {
		return cloneMessage(data)
	}

	switch *envelope.EventType {
	case mqmsg.EventTypeMessageCreate, mqmsg.EventTypeMessageUpdate:
		var payload struct {
			GuildID *int64      `json:"guild_id"`
			Message dto.Message `json:"message"`
		}
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return cloneMessage(data)
		}
		if payload.Message.Nonce == nil || payload.Message.Author.Id == userID {
			return cloneMessage(data)
		}

		payload.Message.Nonce = nil
		redactedData, err := json.Marshal(payload)
		if err != nil {
			return cloneMessage(data)
		}
		envelope.Data = redactedData

	case mqmsg.EventTypeMessageReactionAdd, mqmsg.EventTypeMessageReactionRemove:
		var payload struct {
			GuildID   *int64              `json:"guild_id"`
			ChannelID int64               `json:"channel_id"`
			MessageID int64               `json:"message_id"`
			Reaction  dto.MessageReaction `json:"reaction"`
		}
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return cloneMessage(data)
		}
		if cache != nil {
			bucketKey := reactionutil.BucketKeyFromParts(payload.Reaction.Emoji.Id != nil, reactionDTOEmojiID(payload.Reaction), payload.Reaction.Emoji.Name)
			if reactionID, err := cache.HGet(context.Background(), reactionutil.UserKey(payload.MessageID, userID), bucketKey); err == nil {
				payload.Reaction.Me = reactionID != ""
			}
		}
		redactedData, err := json.Marshal(payload)
		if err != nil {
			return cloneMessage(data)
		}
		envelope.Data = redactedData

	default:
		return cloneMessage(data)
	}

	out, err := json.Marshal(envelope)
	if err != nil {
		return cloneMessage(data)
	}
	return out
}

func reactionDTOEmojiID(reaction dto.MessageReaction) int64 {
	if reaction.Emoji.Id == nil {
		return 0
	}
	return *reaction.Emoji.Id
}

func cloneMessage(data []byte) []byte {
	cp := make([]byte, len(data))
	copy(cp, data)
	return cp
}
