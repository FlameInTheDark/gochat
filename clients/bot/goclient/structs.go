package goclient

import (
	"encoding/json"
	"time"
)

// ChannelType identifies the kind of GoChat channel.
type ChannelType int

const (
	// ChannelTypeGuild is a guild text channel.
	ChannelTypeGuild ChannelType = iota
	// ChannelTypeGuildVoice is a guild voice channel.
	ChannelTypeGuildVoice
	// ChannelTypeGuildCategory is a guild category channel.
	ChannelTypeGuildCategory
	// ChannelTypeDM is a one-to-one direct message channel.
	ChannelTypeDM
	// ChannelTypeGroupDM is a group direct message channel.
	ChannelTypeGroupDM
	// ChannelTypeThread is a thread channel.
	ChannelTypeThread
)

// User is the public user profile payload returned to bots.
type User struct {
	ID            int64       `json:"id"`
	Name          string      `json:"name"`
	Discriminator string      `json:"discriminator"`
	Bio           *string     `json:"bio,omitempty"`
	BannerColor   *int        `json:"banner_color,omitempty"`
	PanelColor    *int        `json:"panel_color,omitempty"`
	Avatar        *AvatarData `json:"avatar,omitempty"`
	Banner        *BannerData `json:"banner,omitempty"`
	PersonalNote  *string     `json:"personal_note,omitempty"`
	IsBot         bool        `json:"is_bot"`
}

// AvatarData describes a user's active avatar.
type AvatarData struct {
	ID          int64   `json:"id"`
	URL         string  `json:"url"`
	ContentType *string `json:"content_type,omitempty"`
	Width       *int64  `json:"width,omitempty"`
	Height      *int64  `json:"height,omitempty"`
	Size        int64   `json:"size"`
}

// BannerData describes a user's active banner.
type BannerData struct {
	Exists      bool    `json:"exists"`
	ID          int64   `json:"id,omitempty"`
	URL         string  `json:"url,omitempty"`
	ContentType *string `json:"content_type,omitempty"`
	Width       *int64  `json:"width,omitempty"`
	Height      *int64  `json:"height,omitempty"`
	Size        int64   `json:"size,omitempty"`
}

// Icon describes a guild icon.
type Icon struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	Filesize int64  `json:"filesize"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
}

// Guild is a guild visible to the bot.
type Guild struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Icon            *Icon  `json:"icon,omitempty"`
	Owner           int64  `json:"owner"`
	Public          bool   `json:"public"`
	Permissions     int64  `json:"permissions"`
	SystemChannelID *int64 `json:"system_channel_id,omitempty"`
}

// GuildResponse is the bot guild list payload with granted install permissions.
type GuildResponse struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Owner              int64  `json:"owner"`
	Public             bool   `json:"public"`
	GrantedPermissions int64  `json:"granted_permissions"`
}

// Role describes a guild role.
type Role struct {
	ID          int64  `json:"id"`
	GuildID     int64  `json:"guild_id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Permissions int64  `json:"permissions"`
	Position    int    `json:"position"`
	Hoist       bool   `json:"hoist"`
}

// Member describes a guild member.
type Member struct {
	User     User      `json:"user"`
	Username *string   `json:"username,omitempty"`
	Avatar   *int64    `json:"avatar,omitempty"`
	JoinAt   time.Time `json:"join_at"`
	Roles    []int64   `json:"roles,omitempty"`
}

// ThreadMember describes a user's thread membership state.
type ThreadMember struct {
	UserID        int64     `json:"user_id"`
	JoinTimestamp time.Time `json:"join_timestamp"`
	Flags         int       `json:"flags"`
}

// Channel is a channel visible to the bot.
type Channel struct {
	ID            int64         `json:"id"`
	Type          ChannelType   `json:"type"`
	GuildID       *int64        `json:"guild_id,omitempty"`
	ParticipantID *int64        `json:"participant_id,omitempty"`
	CreatorID     *int64        `json:"creator_id,omitempty"`
	Member        *ThreadMember `json:"member,omitempty"`
	MemberIDs     []int64       `json:"member_ids,omitempty"`
	Name          string        `json:"name"`
	ParentID      *int64        `json:"parent_id,omitempty"`
	Position      int           `json:"position"`
	Topic         *string       `json:"topic,omitempty"`
	Permissions   *int64        `json:"permissions,omitempty"`
	Private       bool          `json:"private"`
	Closed        bool          `json:"closed"`
	Roles         []int64       `json:"roles,omitempty"`
	LastMessageID int64         `json:"last_message_id"`
	MessageCount  *int64        `json:"message_count,omitempty"`
	VoiceRegion   *string       `json:"voice_region,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// ChannelOrder describes an updated channel position.
type ChannelOrder struct {
	ID       int64 `json:"id"`
	Position int   `json:"position"`
}

// Message is a message visible to the bot.
type Message struct {
	ID                 int64               `json:"id"`
	ChannelID          int64               `json:"channel_id"`
	Author             User                `json:"author"`
	Content            string              `json:"content"`
	Position           *int64              `json:"position,omitempty"`
	Nonce              *string             `json:"nonce,omitempty"`
	Attachments        []Attachment        `json:"attachments,omitempty"`
	Embeds             []Embed             `json:"embeds,omitempty"`
	Flags              int                 `json:"flags,omitempty"`
	Type               int                 `json:"type"`
	Reference          *int64              `json:"reference,omitempty"`
	ReferenceChannelID *int64              `json:"reference_channel_id,omitempty"`
	ThreadID           *int64              `json:"thread_id,omitempty"`
	Thread             *Channel            `json:"thread,omitempty"`
	Reactions          []MessageReaction   `json:"reactions,omitempty"`
	UpdatedAt          *time.Time          `json:"updated_at,omitempty"`
	Interaction        *MessageInteraction `json:"interaction,omitempty"`
}

// MessageInteraction describes the application command interaction that produced a message.
type MessageInteraction struct {
	ID            int64  `json:"id"`
	ApplicationID int64  `json:"application_id"`
	CommandID     int64  `json:"command_id"`
	CommandName   string `json:"command_name"`
	UserID        int64  `json:"user_id"`
}

// Attachment describes a message attachment.
type Attachment struct {
	ContentType *string `json:"content_type,omitempty"`
	Filename    string  `json:"filename"`
	Height      *int64  `json:"height,omitempty"`
	Width       *int64  `json:"width,omitempty"`
	URL         string  `json:"url"`
	PreviewURL  *string `json:"preview_url,omitempty"`
	Size        int64   `json:"size"`
}

// MessageReaction is an aggregated message reaction.
type MessageReaction struct {
	Count int                  `json:"count"`
	Me    bool                 `json:"me"`
	Emoji MessageReactionEmoji `json:"emoji"`
}

// MessageReactionEmoji is a built-in or custom emoji reference.
type MessageReactionEmoji struct {
	ID   *int64 `json:"id"`
	Name string `json:"name"`
}

// MessageReactionUsersPage is a paginated list of reaction users.
type MessageReactionUsersPage struct {
	Items     []User `json:"items"`
	NextAfter *int64 `json:"next_after,omitempty"`
}

// GuildEmoji is a custom guild emoji.
type GuildEmoji struct {
	ID       int64  `json:"id,string"`
	GuildID  int64  `json:"guild_id,string"`
	Name     string `json:"name"`
	Animated bool   `json:"animated"`
}

// MessageSend is the request body for sending a bot message.
type MessageSend struct {
	Content     string  `json:"content"`
	Attachments []int64 `json:"attachments,omitempty"`
	Embeds      []Embed `json:"embeds,omitempty"`
	Reference   *int64  `json:"reference,omitempty"`
}

// MessageEdit is the request body for editing a bot message.
type MessageEdit struct {
	Content *string  `json:"content,omitempty"`
	Embeds  *[]Embed `json:"embeds,omitempty"`
	Flags   *int     `json:"flags,omitempty"`
}

// MeResponse is returned by /bot/api/v1/user/me.
type MeResponse struct {
	BotUserID          int64  `json:"bot_user_id"`
	OwnerUserID        int64  `json:"owner_user_id"`
	User               User   `json:"user"`
	Description        string `json:"description"`
	Public             bool   `json:"public"`
	DefaultPermissions int64  `json:"default_permissions"`
}

// ActiveStream describes an active screen/application stream.
type ActiveStream struct {
	ID         int64  `json:"id"`
	ChannelID  int64  `json:"channel_id"`
	SourceType string `json:"source_type"`
	AudioMode  string `json:"audio_mode"`
	StartedAt  int64  `json:"started_at"`
}

// RawMessage keeps unknown or forward-compatible event payloads available.
type RawMessage = json.RawMessage
