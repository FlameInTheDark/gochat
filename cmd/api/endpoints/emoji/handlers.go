package emoji

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"

	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	"github.com/FlameInTheDark/gochat/internal/dto"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
	"github.com/FlameInTheDark/gochat/internal/helper"
)

// GetInfo
//
//	@Summary	Get emoji info
//	@Produce	json
//	@Tags		Emoji
//	@Param		emoji_id	path		int64			true	"Emoji ID"
//	@Success	200			{object}	dto.EmojiInfo	"Emoji info"
//	@failure	400			{string}	string			"Bad request"
//	@failure	401			{string}	string			"Unauthorized"
//	@failure	404			{string}	string			"Not found"
//	@failure	500			{string}	string			"Internal server error"
//	@Router		/info/emoji/{emoji_id} [get]
func (e *entity) GetInfo(c *fiber.Ctx) error {
	emojiID, err := e.parseEmojiID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	lookup, err := e.getEmojiLookupCached(c.UserContext(), emojiID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetEmojiInfo)
	}
	if lookup == nil || !lookup.Done {
		return fiber.NewError(fiber.StatusNotFound, ErrEmojiNotFound)
	}

	guild, err := e.guild.GetGuildById(c.UserContext(), lookup.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetEmojiInfo)
	}

	var gico *dto.Icon

	if guild.Icon != nil {
		key := fmt.Sprintf("icons:%d:%d", guild.Id, *guild.Icon)
		var cached dto.Icon
		if err := e.cache.GetJSON(c.UserContext(), key, &cached); err == nil && cached.URL != "" {
			gico = &cached
		} else {
			if ic, err := e.icon.GetIcon(c.UserContext(), *guild.Icon, guild.Id); err == nil && ic.URL != nil {
				var w, h, size int64
				if ic.Width != nil {
					w = *ic.Width
				}
				if ic.Height != nil {
					h = *ic.Height
				}
				size = ic.FileSize
				var urlStr string
				if ic.URL != nil {
					urlStr = *ic.URL
				}
				ico := dto.Icon{Id: *guild.Icon, URL: urlStr, Filesize: size, Width: w, Height: h}
				gico = &ico
				_ = e.cache.SetJSON(c.UserContext(), key, ico)
			}
		}
	}

	resp := dto.EmojiInfo{
		Name:          lookup.Name,
		ServerPrivate: !guild.Public,
		Icon:          gico,
	}
	if guild.Public {
		resp.ServerName = &guild.Name
		return c.JSON(resp)
	}

	isMember, err := e.member.IsGuildMember(c.UserContext(), guild.Id, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetEmojiInfo)
	}
	if isMember {
		resp.ServerName = &guild.Name
	}

	return c.JSON(resp)
}

func (e *entity) getEmojiLookupCached(ctx context.Context, emojiID int64) (*emojiutil.LookupCacheEntry, error) {
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
