package upload

import (
	"fmt"
	"strings"
)

func AttachmentOriginalKey(channelID, attachmentID int64) string {
	return fmt.Sprintf("media/%d/%d/original", channelID, attachmentID)
}

func AttachmentPreviewKey(channelID, attachmentID int64) string {
	return fmt.Sprintf("media/%d/%d/preview.webp", channelID, attachmentID)
}

func AvatarKey(userID, avatarID int64) string {
	return fmt.Sprintf("avatars/%d/%d.webp", userID, avatarID)
}

func IconKey(guildID, iconID int64) string {
	return fmt.Sprintf("icons/%d/%d.webp", guildID, iconID)
}

func EmojiMasterKey(emojiID int64) string {
	return fmt.Sprintf("emojis/%d/master.webp", emojiID)
}

func EmojiSizedKey(emojiID int64, size int) string {
	return fmt.Sprintf("emojis/%d/%d.webp", emojiID, size)
}

func EmojiVariantKey(emojiID int64, variant string) string {
	switch variant {
	case "44":
		return EmojiSizedKey(emojiID, 44)
	case "96":
		return EmojiSizedKey(emojiID, 96)
	default:
		return EmojiMasterKey(emojiID)
	}
}

func PublicURL(base, key string) string {
	if base == "" {
		return key
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(key, "/")
}
