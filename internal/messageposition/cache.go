package messageposition

import "fmt"

const (
	BlockSize       int64 = 256
	CacheTTLSeconds int64 = 60 * 60 * 24 * 30
	LockTTLSeconds  int64 = 5
)

func CurrentKey(channelID int64) string {
	return fmt.Sprintf("channel:message_position:current:%d", channelID)
}

func ReservedMaxKey(channelID int64) string {
	return fmt.Sprintf("channel:message_position:reserved_max:%d", channelID)
}

func LockKey(channelID int64) string {
	return fmt.Sprintf("channel:message_position:lock:%d", channelID)
}
