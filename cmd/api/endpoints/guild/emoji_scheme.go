package guild

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
)

const (
	ErrIncorrectEmojiID         = "incorrect emoji ID"
	ErrUnableToCreateEmoji      = "unable to create emoji"
	ErrUnableToGetEmojis        = "unable to get emojis"
	ErrUnableToUpdateEmoji      = "unable to update emoji"
	ErrUnableToDeleteEmoji      = "unable to delete emoji"
	ErrEmojiNameRequired        = "emoji name is required"
	ErrEmojiNameInvalid         = "emoji name can only contain latin letters, numbers, and hyphens"
	ErrEmojiNameTaken           = "emoji name already exists"
	ErrEmojiQuotaExceeded       = "emoji limit reached"
	ErrEmojiActiveLimitExceeded = "too many pending emoji uploads"
	ErrEmojiStorageUnavailable  = "emoji storage is not configured"
)

type CreateEmojiRequest struct {
	Name        string `json:"name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
}

func (r CreateEmojiRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error(ErrEmojiNameRequired),
			validation.By(func(v interface{}) error {
				name, _ := v.(string)
				if !emojiutil.NameRegex.MatchString(name) {
					return validation.NewError("validation", ErrEmojiNameInvalid)
				}
				return nil
			}),
		),
		validation.Field(&r.FileSize,
			validation.Min(int64(1)).Error(ErrFileIsTooBig),
			validation.Max(emojiutil.MaxUploadSizeBytes).Error(ErrFileIsTooBig),
		),
		validation.Field(&r.ContentType,
			validation.Required,
			validation.By(func(v interface{}) error {
				contentType, _ := v.(string)
				if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
					return fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported content type")
				}
				return nil
			}),
		),
	)
}

type UpdateEmojiRequest struct {
	Name string `json:"name"`
}

func (r UpdateEmojiRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error(ErrEmojiNameRequired),
			validation.By(func(v interface{}) error {
				name, _ := v.(string)
				if !emojiutil.NameRegex.MatchString(name) {
					return validation.NewError("validation", ErrEmojiNameInvalid)
				}
				return nil
			}),
		),
	)
}

func (e *entity) parseEmojiID(c *fiber.Ctx) (int64, error) {
	emojiID, err := c.ParamsInt("emoji_id")
	if err != nil || emojiID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectEmojiID)
	}
	return int64(emojiID), nil
}

func guildEmojiToDTO(emoji model.GuildEmoji) dto.GuildEmoji {
	return dto.GuildEmoji{
		Id:       emoji.Id,
		GuildId:  emoji.GuildId,
		Name:     emoji.Name,
		Animated: emoji.Animated,
	}
}

func guildEmojisToDTO(emojis []model.GuildEmoji) []dto.GuildEmoji {
	result := make([]dto.GuildEmoji, 0, len(emojis))
	for _, item := range emojis {
		result = append(result, guildEmojiToDTO(item))
	}
	return result
}
