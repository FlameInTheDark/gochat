package user

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/dto"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
)

func (e *entity) getGuildEmojiSettingsMap(ctx context.Context, guildIDs []int64) (map[int64][]dto.EmojiRef, error) {
	result := make(map[int64][]dto.EmojiRef, len(guildIDs))
	misses := make([]int64, 0, len(guildIDs))

	for _, guildID := range guildIDs {
		result[guildID] = []dto.EmojiRef{}
		if e.cache == nil {
			misses = append(misses, guildID)
			continue
		}

		var cached []dto.GuildEmoji
		if err := e.cache.GetJSON(ctx, emojiutil.GuildCacheKey(guildID), &cached); err == nil {
			result[guildID] = guildEmojiRefs(cached)
			continue
		}
		misses = append(misses, guildID)
	}

	if len(misses) == 0 {
		return result, nil
	}

	rows, err := e.emoji.ListReadyGuildEmojisByGuilds(ctx, misses)
	if err != nil {
		return nil, err
	}

	grouped := make(map[int64][]dto.GuildEmoji, len(misses))
	for _, row := range rows {
		grouped[row.GuildId] = append(grouped[row.GuildId], dto.GuildEmoji{
			Id:       row.Id,
			GuildId:  row.GuildId,
			Name:     row.Name,
			Animated: row.Animated,
		})
	}

	for _, guildID := range misses {
		payload := grouped[guildID]
		if payload == nil {
			payload = []dto.GuildEmoji{}
		}
		result[guildID] = guildEmojiRefs(payload)
		if e.cache != nil {
			_ = e.cache.SetTimedJSON(ctx, emojiutil.GuildCacheKey(guildID), payload, emojiutil.GuildCacheTTLSeconds)
		}
	}

	return result, nil
}

func guildEmojiRefs(emojis []dto.GuildEmoji) []dto.EmojiRef {
	refs := make([]dto.EmojiRef, 0, len(emojis))
	for _, emoji := range emojis {
		refs = append(refs, dto.EmojiRef{Name: emoji.Name, Id: emoji.Id})
	}
	if refs == nil {
		return []dto.EmojiRef{}
	}
	return refs
}
