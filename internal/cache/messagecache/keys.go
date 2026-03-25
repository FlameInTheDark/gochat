package messagecache

import (
	"fmt"
	"strconv"
)

const (
	// MessageTTLSeconds is how long an individual fully-built dto.Message lives in cache.
	// Short enough that reactions, embeds, and edits stay reasonably fresh on a miss.
	MessageTTLSeconds int64 = 300 // 5 minutes

	// IndexTTLSeconds is the safety-net TTL reset on the sorted-set index key on every write.
	// Inactive channels clean up automatically after this window.
	IndexTTLSeconds int64 = 86400 // 24 hours

	// WindowSize is the number of latest messages we keep in the window.
	// Must match MaxBatchSize in the message handler.
	WindowSize = 50
)

// IndexKey returns the Redis sorted-set key that tracks message IDs for a channel.
// Score = float64(message_id). Member = IDToMember(message_id).
//
// Note on float64 precision: snowflake IDs are ~61-bit integers. float64 has 53-bit mantissa,
// so IDs within the same ~512-ID batch share the same float64 score. Redis then breaks
// ties lexicographically by member. Since all IDs in this range have the same number of
// decimal digits (19), lexicographic order equals numeric order for same-score members.
// Ordering is therefore always correct.
func IndexKey(channelID int64) string {
	return fmt.Sprintf("channel:messages:index:%d", channelID)
}

// MessageKey returns the Redis string key for a single cached dto.Message.
func MessageKey(channelID, messageID int64) string {
	return fmt.Sprintf("channel:message:%d:%d", channelID, messageID)
}

// IDToMember converts a message ID to the sorted-set member string.
func IDToMember(id int64) string {
	return strconv.FormatInt(id, 10)
}

// MemberToID parses a sorted-set member back to int64.
func MemberToID(member string) (int64, error) {
	return strconv.ParseInt(member, 10, 64)
}
