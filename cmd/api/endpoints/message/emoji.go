package message

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
)

var customEmojiTagRegex = regexp.MustCompile(`<:([A-Za-z0-9-]+):([0-9]+)>`)

func (e *entity) sanitizeEmojiContent(ctx context.Context, userID int64, content string) (string, error) {
	if content == "" {
		return content, nil
	}

	var firstErr error
	sanitized := customEmojiTagRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := customEmojiTagRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		originalName := parts[1]
		emojiID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return ":" + originalName + ":"
		}

		lookup, err := e.lookupEmojiCached(ctx, emojiID)
		if err != nil {
			firstErr = err
			return match
		}
		if lookup == nil || !lookup.Done {
			return ":" + originalName + ":"
		}

		ok, err := e.m.IsGuildMember(ctx, lookup.GuildId, userID)
		if err != nil {
			firstErr = err
			return match
		}
		if !ok {
			return ":" + originalName + ":"
		}
		return fmt.Sprintf("<:%s:%d>", lookup.Name, emojiID)
	})
	if firstErr != nil {
		return "", firstErr
	}
	return sanitized, nil
}

func (e *entity) lookupEmojiCached(ctx context.Context, emojiID int64) (*emojiutil.LookupCacheEntry, error) {
	if e.cache != nil {
		var cached emojiutil.LookupCacheEntry
		if err := e.cache.GetJSON(ctx, emojiutil.LookupCacheKey(emojiID), &cached); err == nil {
			if cached.Missing {
				return nil, nil
			}
			return &cached, nil
		}
	}

	lookup, err := e.emoji.GetEmojiLookup(ctx, emojiID)
	if err != nil {
		if errors.Is(err, emojirepo.ErrEmojiNotFound) {
			if e.cache != nil {
				_ = e.cache.SetTimedJSON(ctx, emojiutil.LookupCacheKey(emojiID), emojiutil.LookupCacheEntry{Missing: true}, emojiutil.NegativeCacheTTLSeconds)
			}
			return nil, nil
		}
		return nil, err
	}

	cached := emojiutil.LookupCacheEntry{
		Id:       lookup.Id,
		GuildId:  lookup.GuildId,
		Name:     lookup.Name,
		Done:     lookup.Done,
		Animated: lookup.Animated,
	}
	if lookup.Width != nil {
		cached.Width = *lookup.Width
	}
	if lookup.Height != nil {
		cached.Height = *lookup.Height
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(ctx, emojiutil.LookupCacheKey(emojiID), cached, emojiutil.LookupCacheTTLSeconds)
	}
	return &cached, nil
}
