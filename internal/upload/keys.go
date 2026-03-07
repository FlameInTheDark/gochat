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

func PublicURL(base, key string) string {
	if base == "" {
		return key
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(key, "/")
}
