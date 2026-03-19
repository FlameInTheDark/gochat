package threadcount

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DeltaTTLSeconds = 60 * 60 * 24 * 30
	DeltaKeyPattern = "thread:message_count_delta:*"
)

func DeltaKey(threadID int64) string {
	return fmt.Sprintf("thread:message_count_delta:%d", threadID)
}

func ParseDeltaKey(key string) (int64, error) {
	threadIDStr := strings.TrimPrefix(key, "thread:message_count_delta:")
	return strconv.ParseInt(threadIDStr, 10, 64)
}
