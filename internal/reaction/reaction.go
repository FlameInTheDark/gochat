package reaction

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	UsersPageDefaultLimit   = 50
	UsersPageMaxLimit       = 100
	FlushStreamKey          = "reaction:flush"
	FlushStreamApproxMaxLen = 100000
	WarmMarkerField         = "_warm"
)

type ParsedName struct {
	Raw       string
	BucketKey string
	Custom    bool
	EmojiId   int64
	EmojiName string
}

type CacheSummaryMeta struct {
	Custom    bool   `json:"custom"`
	EmojiId   int64  `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name"`
}

type FlushEvent struct {
	Op        string `json:"op"`
	MessageId int64  `json:"message_id"`
	Reaction  model.Reaction
	Count     int `json:"count"`
}

func ParseReactionName(raw string) (ParsedName, error) {
	decoded, err := url.PathUnescape(strings.TrimSpace(raw))
	if err != nil {
		return ParsedName{}, fmt.Errorf("unable to decode reaction name: %w", err)
	}
	if decoded == "" {
		return ParsedName{}, fmt.Errorf("reaction name is required")
	}

	if idx := strings.LastIndex(decoded, ":"); idx > 0 && idx < len(decoded)-1 {
		id, err := strconv.ParseInt(decoded[idx+1:], 10, 64)
		if err == nil && id > 0 {
			name := decoded[:idx]
			return ParsedName{
				Raw:       decoded,
				BucketKey: CustomBucketKey(id),
				Custom:    true,
				EmojiId:   id,
				EmojiName: name,
			}, nil
		}
	}

	return ParsedName{
		Raw:       decoded,
		BucketKey: SystemBucketKey(decoded),
		EmojiName: decoded,
	}, nil
}

func SystemBucketKey(name string) string {
	return "u:" + name
}

func CustomBucketKey(emojiId int64) string {
	return fmt.Sprintf("c:%d", emojiId)
}

func BucketKeyFromParts(custom bool, emojiId int64, emojiName string) string {
	if custom {
		return CustomBucketKey(emojiId)
	}
	return SystemBucketKey(emojiName)
}

func BucketKeyEmoji(bucketKey string) (custom bool, emojiId int64, emojiName string) {
	switch {
	case strings.HasPrefix(bucketKey, "c:"):
		id, _ := strconv.ParseInt(strings.TrimPrefix(bucketKey, "c:"), 10, 64)
		return true, id, ""
	case strings.HasPrefix(bucketKey, "u:"):
		return false, 0, strings.TrimPrefix(bucketKey, "u:")
	default:
		return false, 0, bucketKey
	}
}

func EncodedBucketKey(bucketKey string) string {
	return url.QueryEscape(bucketKey)
}

func SummaryCountKey(messageId int64) string {
	return fmt.Sprintf("reaction:summary:count:%d", messageId)
}

func SummaryMetaKey(messageId int64) string {
	return fmt.Sprintf("reaction:summary:meta:%d", messageId)
}

func UserKey(messageId, userId int64) string {
	return fmt.Sprintf("reaction:user:%d:%d", messageId, userId)
}

func BucketSortedSetKey(messageId int64, bucketKey string) string {
	return fmt.Sprintf("reaction:bucket:%d:%s:z", messageId, EncodedBucketKey(bucketKey))
}

func BucketRowsKey(messageId int64, bucketKey string) string {
	return fmt.Sprintf("reaction:bucket:%d:%s:rows", messageId, EncodedBucketKey(bucketKey))
}

func ParseLimit(raw int) int {
	if raw <= 0 {
		return UsersPageDefaultLimit
	}
	if raw > UsersPageMaxLimit {
		return UsersPageMaxLimit
	}
	return raw
}

func EncodeReactionRow(row model.Reaction) (string, error) {
	b, err := json.Marshal(row)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
