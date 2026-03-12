package dto

import (
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Channel struct {
	Id            int64             `json:"id" example:"2230469276416868352"`                       // Channel ID
	Type          model.ChannelType `json:"type" example:"0"`                                       // Channel type
	GuildId       *int64            `json:"guild_id,omitempty" example:"2230469276416868352"`       // Guild ID channel was created in
	ParticipantId *int64            `json:"participant_id,omitempty" example:"2230469276416868352"` // For DM channels: the other participant's user ID
	CreatorId     *int64            `json:"creator_id,omitempty" example:"2230469276416868352"`     // For threads: user who created the thread
	Member        *ThreadMember     `json:"member,omitempty"`                                       // For threads: current user's membership state when returned via HTTP.
	MemberIds     []int64           `json:"member_ids,omitempty" example:"2230469276416868352"`     // For threads: IDs of users who have joined the thread.
	Name          string            `json:"name" example:"channel-name"`                            // Channel name, without spaces
	ParentId      *int64            `json:"parent_id,omitempty" example:"2230469276416868352"`      // Parent channel id
	Position      int               `json:"position" example:"4"`                                   // Channel position
	Topic         *string           `json:"topic" example:"Just a channel topic"`                   // Channel topic.
	Permissions   *int64            `json:"permissions,omitempty"`                                  // Permissions. Check the permissions documentation for more info.
	Private       bool              `json:"private" default:"false"`                                // Whether the channel is private. Private channels can only be seen by users with roles assigned to this channel.
	Closed        bool              `json:"closed"`                                                 // Whether the thread is closed for new messages.
	Roles         []int64           `json:"roles,omitempty" example:"2230469276416868352"`          // Roles IDs
	LastMessageId int64             `json:"last_message_id" example:"2230469276416868352"`          // ID of the last message in the channel
	MessageCount  *int64            `json:"message_count,omitempty" example:"42"`                   // For threads: approximate stored message count.
	VoiceRegion   *string           `json:"voice_region,omitempty" example:"us-east"`               // Voice channel region
	CreatedAt     time.Time         `json:"created_at"`                                             // Timestamp of channel creation
}

type ChannelOrder struct {
	Id       int64 `json:"id"`       // Channel ID.
	Position int   `json:"position"` // New channel position.
}
