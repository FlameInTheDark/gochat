package goclient

import "encoding/json"

// OPCodeType identifies a gateway operation.
type OPCodeType int

const (
	// OpCodeDispatch carries a server-dispatched event.
	OpCodeDispatch OPCodeType = iota
	// OPCodeHello identifies a bot shard when sent by the client and carries
	// heartbeat configuration when sent by the server.
	OPCodeHello
	// OPCodeHeartBeat keeps the gateway session alive.
	OPCodeHeartBeat
	// OPCodePresenceUpdate updates or dispatches user presence.
	OPCodePresenceUpdate
	// OPCodeGuildUpdateSubscription is reserved by the core gateway protocol.
	OPCodeGuildUpdateSubscription
	// OPCodeChannelSubscription is reserved by the core gateway protocol.
	OPCodeChannelSubscription
	// OPCodePresenceSubscription is reserved by the core gateway protocol.
	OPCodePresenceSubscription
	// OPCodeRTC is reserved for RTC signaling frames.
	OPCodeRTC
	// OPCodeHeartbeatAck acknowledges a heartbeat.
	OPCodeHeartbeatAck
)

// EventType identifies a gateway dispatch payload.
type EventType int

const (
	// EventTypeGatewayReady is sent after a shard identifies successfully.
	EventTypeGatewayReady EventType = 1
)

const (
	// EventTypeMessageCreate is dispatched when a visible message is created.
	EventTypeMessageCreate EventType = 100 + iota
	// EventTypeMessageUpdate is dispatched when a visible message is edited.
	EventTypeMessageUpdate
	// EventTypeMessageDelete is dispatched when a visible message is deleted.
	EventTypeMessageDelete
	// EventTypeMessageReactionAdd is dispatched when a visible reaction is added.
	EventTypeMessageReactionAdd
	// EventTypeMessageReactionRemove is dispatched when a visible reaction is removed.
	EventTypeMessageReactionRemove
	// EventTypeGuildCreate is reserved for guild create events.
	EventTypeGuildCreate
	// EventTypeGuildUpdate is dispatched when a visible guild changes.
	EventTypeGuildUpdate
	// EventTypeGuildDelete is reserved for guild delete events.
	EventTypeGuildDelete
	// EventTypeChannelCreate is dispatched when a visible channel is created.
	EventTypeChannelCreate
	// EventTypeChannelUpdate is dispatched when a visible channel changes.
	EventTypeChannelUpdate
	// EventTypeChannelOrderUpdate is dispatched when visible channel order changes.
	EventTypeChannelOrderUpdate
	// EventTypeChannelDelete is dispatched when a visible channel is deleted.
	EventTypeChannelDelete
	// EventTypeGuildRoleCreate is dispatched when a visible guild role is created.
	EventTypeGuildRoleCreate
	// EventTypeGuildRoleUpdate is dispatched when a visible guild role changes.
	EventTypeGuildRoleUpdate
	// EventTypeGuildRoleDelete is dispatched when a visible guild role is deleted.
	EventTypeGuildRoleDelete
	// EventTypeThreadCreate is dispatched when a visible thread is created.
	EventTypeThreadCreate
	// EventTypeThreadUpdate is dispatched when a visible thread changes.
	EventTypeThreadUpdate
	// EventTypeThreadDelete is dispatched when a visible thread is deleted.
	EventTypeThreadDelete
	// EventTypeGuildEmojiCreate is dispatched when a guild emoji is created.
	EventTypeGuildEmojiCreate
	// EventTypeGuildEmojiUpdate is dispatched when a guild emoji changes.
	EventTypeGuildEmojiUpdate
	// EventTypeGuildEmojiDelete is dispatched when a guild emoji is deleted.
	EventTypeGuildEmojiDelete
)

const (
	// EventTypeGuildMemberAdd is dispatched when a guild member is added.
	EventTypeGuildMemberAdd EventType = 200 + iota
	// EventTypeGuildMemberUpdate is dispatched when a guild member changes.
	EventTypeGuildMemberUpdate
	// EventTypeGuildMemberRemove is dispatched when a guild member is removed.
	EventTypeGuildMemberRemove
	// EventTypeGuildMemberAddRole is dispatched when a role is added to a member.
	EventTypeGuildMemberAddRole
	// EventTypeGuildMemberRemoveRole is dispatched when a role is removed from a member.
	EventTypeGuildMemberRemoveRole
	// EventTypeGuildMemberJoinVoice is dispatched when a member joins voice.
	EventTypeGuildMemberJoinVoice
	// EventTypeGuildMemberLeaveVoice is dispatched when a member leaves voice.
	EventTypeGuildMemberLeaveVoice
	// EventTypeGuildMemberModeration is dispatched for moderation actions.
	EventTypeGuildMemberModeration
)

const (
	// EventTypeGuildVoiceRegionChanging is dispatched before voice region migration.
	EventTypeGuildVoiceRegionChanging EventType = 208
	// EventTypeVoiceStateUpdate is dispatched when a user's voice state changes.
	EventTypeVoiceStateUpdate EventType = 209
)

const (
	// EventTypeGuildMemberStartStream is dispatched when a member starts streaming.
	EventTypeGuildMemberStartStream EventType = 210 + iota
	// EventTypeGuildMemberStopStream is dispatched when a member stops streaming.
	EventTypeGuildMemberStopStream
	// EventTypeGuildStreamsRebind is dispatched when active streams should reconnect.
	EventTypeGuildStreamsRebind
)

const (
	// EventTypeGuildChannelMessage is a compact message notification.
	EventTypeGuildChannelMessage EventType = 300 + iota
	// EventTypeChannelUserTyping is dispatched when a visible user starts typing.
	EventTypeChannelUserTyping
	// EventTypeMention is dispatched when a visible mention is created.
	EventTypeMention
)

const (
	// EventTypeUserUpdateReadState is reserved for user read-state updates.
	EventTypeUserUpdateReadState EventType = 400 + iota
	// EventTypeUserUpdateSettings is reserved for user settings updates.
	EventTypeUserUpdateSettings
	// EventTypeUserFriendRequest is reserved for friend request events.
	EventTypeUserFriendRequest
	// EventTypeUserFriendAdded is reserved for friendship events.
	EventTypeUserFriendAdded
	// EventTypeUserFriendRemoved is reserved for friendship removal events.
	EventTypeUserFriendRemoved
	// EventTypeUserDMMessage is dispatched when the bot receives a DM.
	EventTypeUserDMMessage
	// EventTypeUserUpdate is reserved for user profile updates.
	EventTypeUserUpdate
	// EventTypeUserAuthRevoked is reserved for auth revocation events.
	EventTypeUserAuthRevoked
	// EventTypeUserDMCallStarted is reserved for DM call events.
	EventTypeUserDMCallStarted
	// EventTypeUserDMCallJoined is reserved for DM call events.
	EventTypeUserDMCallJoined
	// EventTypeUserDMCallDeclined is reserved for DM call events.
	EventTypeUserDMCallDeclined
	// EventTypeUserDMCallLeft is reserved for DM call events.
	EventTypeUserDMCallLeft
	// EventTypeUserDMCallEnded is reserved for DM call events.
	EventTypeUserDMCallEnded
	// EventTypeUserDMCallStreamStarted is reserved for DM call stream events.
	EventTypeUserDMCallStreamStarted
	// EventTypeUserDMCallStreamStopped is reserved for DM call stream events.
	EventTypeUserDMCallStreamStopped
)

const (
	// EventTypeRTCJoin identifies an RTC join signal.
	EventTypeRTCJoin EventType = 500 + iota
	// EventTypeRTCOffer identifies an RTC offer signal.
	EventTypeRTCOffer
	// EventTypeRTCAnswer identifies an RTC answer signal.
	EventTypeRTCAnswer
	// EventTypeRTCCandidate identifies an RTC ICE candidate signal.
	EventTypeRTCCandidate
	// EventTypeRTCLeave identifies an RTC leave signal.
	EventTypeRTCLeave
)

const (
	// EventTypeRTCMuteSelf identifies a local mute state update.
	EventTypeRTCMuteSelf EventType = 505 + iota
	// EventTypeRTCMuteUser identifies a local per-user mute state update.
	EventTypeRTCMuteUser
	// EventTypeRTCServerMuteUser identifies a server-side mute command.
	EventTypeRTCServerMuteUser
	// EventTypeRTCServerDeafenUser identifies a server-side deafen command.
	EventTypeRTCServerDeafenUser
	// EventTypeRTCBindingAlive keeps an RTC binding alive.
	EventTypeRTCBindingAlive
	// EventTypeRTCServerKickUser identifies a server-side kick command.
	EventTypeRTCServerKickUser
	// EventTypeRTCServerBlockUser identifies a server-side block command.
	EventTypeRTCServerBlockUser
	// EventTypeRTCMoved tells a client to move to another RTC channel.
	EventTypeRTCMoved
	// EventTypeRTCServerRebind tells clients to reconnect to the current RTC channel.
	EventTypeRTCServerRebind
	// EventTypeRTCSpeaking identifies a speaking-state update.
	EventTypeRTCSpeaking
)

const (
	// EventTypeRTCIdentify identifies an RTC v2 identify signal.
	EventTypeRTCIdentify EventType = 530 + iota
	// EventTypeRTCReady identifies an RTC v2 ready signal.
	EventTypeRTCReady
	// EventTypeRTCSelectProtocol identifies an RTC v2 select-protocol signal.
	EventTypeRTCSelectProtocol
	// EventTypeRTCSessionDescription identifies an RTC v2 session description.
	EventTypeRTCSessionDescription
)

const (
	// EventTypeRTCError identifies an RTC error signal.
	EventTypeRTCError EventType = 539
)

// GatewayMessage is the JSON envelope used by the bot gateway.
type GatewayMessage struct {
	Operation OPCodeType      `json:"op"`
	Data      json.RawMessage `json:"d"`
	EventType *EventType      `json:"t,omitempty"`
}

// UnmarshalJSON supports gateway frames that spell the payload field as "data".
func (m *GatewayMessage) UnmarshalJSON(b []byte) error {
	var aux struct {
		Operation OPCodeType      `json:"op"`
		D         json.RawMessage `json:"d"`
		Data      json.RawMessage `json:"data"`
		EventType *EventType      `json:"t"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	m.Operation = aux.Operation
	m.EventType = aux.EventType
	if len(aux.D) > 0 {
		m.Data = aux.D
	} else {
		m.Data = aux.Data
	}
	return nil
}

// Event is emitted to generic handlers for every decoded gateway dispatch.
type Event struct {
	Operation OPCodeType      `json:"op"`
	Type      *EventType      `json:"t,omitempty"`
	RawData   json.RawMessage `json:"d"`
	Struct    any             `json:"-"`
}

// Ready is dispatched after a bot shard identifies successfully.
type Ready struct {
	Bot                User            `json:"bot"`
	SessionID          string          `json:"session_id"`
	ShardID            int             `json:"shard_id"`
	ShardCount         int             `json:"shard_count"`
	GuildIDs           []int64         `json:"guild_ids"`
	DMChannelIDs       []int64         `json:"dm_channel_ids"`
	GroupDMChannelIDs  []int64         `json:"group_dm_channel_ids"`
	ReceivesDMEvents   bool            `json:"receives_dm_events"`
	GrantedPermissions map[int64]int64 `json:"granted_permissions"`
}

// HeartbeatInterval is sent by the gateway after identify.
type HeartbeatInterval struct {
	HeartbeatInterval int64  `json:"heartbeat_interval"`
	SessionID         string `json:"session_id,omitempty"`
	ConnectionID      string `json:"connection_id,omitempty"`
	Generation        int64  `json:"generation,omitempty"`
	ProtocolVersion   int    `json:"protocol_version,omitempty"`
}

// HeartbeatAck acknowledges a heartbeat.
type HeartbeatAck struct {
	LastEventID  int64  `json:"e"`
	ServerTime   int64  `json:"server_time"`
	ConnectionID string `json:"connection_id,omitempty"`
}

// Presence represents an initial bot presence payload.
type Presence struct {
	Status           string `json:"status"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
}

// PresenceUpdateRequest updates the bot user's gateway presence.
type PresenceUpdateRequest struct {
	Status           string `json:"status"`
	Platform         string `json:"platform,omitempty"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
	VoiceChannelID   *int64 `json:"voice_channel_id,omitempty"`
	Mute             *bool  `json:"mute,omitempty"`
	Deafen           *bool  `json:"deafen,omitempty"`
	SelfVideo        *bool  `json:"self_video,omitempty"`
}

// PresenceUpdate describes a dispatched presence update.
type PresenceUpdate struct {
	UserID           int64             `json:"user_id"`
	Status           string            `json:"status"`
	CustomStatusText string            `json:"custom_status_text,omitempty"`
	Since            int64             `json:"since"`
	ClientStatus     map[string]string `json:"client_status,omitempty"`
	VoiceChannelID   *int64            `json:"voice_channel_id,omitempty"`
	Mute             bool              `json:"mute,omitempty"`
	Deafen           bool              `json:"deafen,omitempty"`
	SelfVideo        bool              `json:"self_video,omitempty"`
	ActiveStream     *ActiveStream     `json:"active_stream,omitempty"`
}

// MessageCreate is dispatched when a visible message is created.
type MessageCreate struct {
	GuildID *int64  `json:"guild_id"`
	Message Message `json:"message"`
}

// MessageUpdate is dispatched when a visible message is edited.
type MessageUpdate struct {
	GuildID *int64  `json:"guild_id"`
	Message Message `json:"message"`
}

// MessageDelete is dispatched when a visible message is deleted.
type MessageDelete struct {
	GuildID   *int64 `json:"guild_id"`
	ChannelID int64  `json:"channel_id"`
	MessageID int64  `json:"message_id"`
}

// MessageReactionAdd is dispatched when a reaction is added.
type MessageReactionAdd struct {
	GuildID   *int64          `json:"guild_id"`
	ChannelID int64           `json:"channel_id"`
	MessageID int64           `json:"message_id"`
	Reaction  MessageReaction `json:"reaction"`
}

// MessageReactionRemove is dispatched when a reaction is removed.
type MessageReactionRemove struct {
	GuildID   *int64          `json:"guild_id"`
	ChannelID int64           `json:"channel_id"`
	MessageID int64           `json:"message_id"`
	Reaction  MessageReaction `json:"reaction"`
}

// GuildCreate is dispatched when a guild becomes visible to the bot.
type GuildCreate struct {
	Guild Guild `json:"guild"`
}

// GuildUpdate is dispatched when a guild visible to the bot changes.
type GuildUpdate struct {
	Guild Guild `json:"guild"`
}

// GuildDelete is dispatched when a guild becomes unavailable to the bot.
type GuildDelete struct {
	GuildID int64 `json:"guild_id"`
}

// ChannelCreate is dispatched when a visible channel is created.
type ChannelCreate struct {
	GuildID *int64  `json:"guild_id"`
	Channel Channel `json:"channel"`
}

// ChannelUpdate is dispatched when a visible channel changes.
type ChannelUpdate struct {
	GuildID *int64  `json:"guild_id"`
	Channel Channel `json:"channel"`
}

// ChannelOrderUpdate is dispatched when visible channel order changes.
type ChannelOrderUpdate struct {
	GuildID  *int64         `json:"guild_id"`
	Channels []ChannelOrder `json:"channels"`
}

// ChannelDelete is dispatched when a visible channel is deleted.
type ChannelDelete struct {
	GuildID     *int64      `json:"guild_id"`
	ChannelType ChannelType `json:"channel_type"`
	ChannelID   int64       `json:"channel_id"`
}

// GuildRoleCreate is dispatched when a guild role is created.
type GuildRoleCreate struct {
	Role Role `json:"role"`
}

// GuildRoleUpdate is dispatched when a guild role changes.
type GuildRoleUpdate struct {
	GuildID int64 `json:"guild_id"`
	Role    Role  `json:"role"`
}

// GuildRoleDelete is dispatched when a guild role is deleted.
type GuildRoleDelete struct {
	GuildID int64 `json:"guild_id"`
	RoleID  int64 `json:"role_id"`
}

// ThreadCreate is dispatched when a visible thread is created.
type ThreadCreate struct {
	GuildID *int64  `json:"guild_id"`
	Thread  Channel `json:"thread"`
}

// ThreadUpdate is dispatched when a visible thread changes.
type ThreadUpdate struct {
	GuildID *int64  `json:"guild_id"`
	Thread  Channel `json:"thread"`
}

// ThreadDelete is dispatched when a visible thread is deleted.
type ThreadDelete struct {
	GuildID  *int64 `json:"guild_id"`
	ThreadID int64  `json:"thread_id"`
}

// GuildEmojiCreate is dispatched when a custom emoji is created.
type GuildEmojiCreate struct {
	Emoji GuildEmoji `json:"emoji"`
}

// GuildEmojiUpdate is dispatched when a custom emoji changes.
type GuildEmojiUpdate struct {
	Emoji GuildEmoji `json:"emoji"`
}

// GuildEmojiDelete is dispatched when a custom emoji is deleted.
type GuildEmojiDelete struct {
	GuildID int64 `json:"guild_id,string"`
	EmojiID int64 `json:"emoji_id,string"`
}

// GuildMemberAdd is dispatched when a guild member is added.
type GuildMemberAdd struct {
	GuildID int64  `json:"guild_id"`
	UserID  int64  `json:"user_id"`
	Member  Member `json:"member"`
}

// GuildMemberUpdate is dispatched when a guild member changes.
type GuildMemberUpdate struct {
	GuildID int64  `json:"guild_id"`
	Member  Member `json:"member"`
}

// GuildMemberRemove is dispatched when a guild member is removed.
type GuildMemberRemove struct {
	GuildID int64 `json:"guild_id"`
	UserID  int64 `json:"user_id"`
}

// GuildMemberAddRole is dispatched when a role is added to a member.
type GuildMemberAddRole struct {
	GuildID int64 `json:"guild_id"`
	RoleID  int64 `json:"role_id"`
	UserID  int64 `json:"user_id"`
}

// GuildMemberRemoveRole is dispatched when a role is removed from a member.
type GuildMemberRemoveRole struct {
	GuildID int64 `json:"guild_id"`
	RoleID  int64 `json:"role_id"`
	UserID  int64 `json:"user_id"`
}

// GuildMemberJoinVoice is dispatched when a member joins voice.
type GuildMemberJoinVoice struct {
	GuildID   int64 `json:"guild_id"`
	UserID    int64 `json:"user_id"`
	ChannelID int64 `json:"channel_id"`
}

// GuildMemberLeaveVoice is dispatched when a member leaves voice.
type GuildMemberLeaveVoice struct {
	GuildID   int64 `json:"guild_id"`
	UserID    int64 `json:"user_id"`
	ChannelID int64 `json:"channel_id"`
}

// GuildMemberModerationAction describes a moderation event action.
type GuildMemberModerationAction string

const (
	// GuildMemberModerationKick is a kick action.
	GuildMemberModerationKick GuildMemberModerationAction = "kick"
	// GuildMemberModerationBan is a ban action.
	GuildMemberModerationBan GuildMemberModerationAction = "ban"
	// GuildMemberModerationUnban is an unban action.
	GuildMemberModerationUnban GuildMemberModerationAction = "unban"
)

// GuildMemberModeration is dispatched for guild moderation actions.
type GuildMemberModeration struct {
	GuildID int64                       `json:"guild_id"`
	UserID  int64                       `json:"user_id"`
	ActorID int64                       `json:"actor_id"`
	Action  GuildMemberModerationAction `json:"action"`
	Reason  *string                     `json:"reason,omitempty"`
}

// VoiceStateUpdate is dispatched when a user's voice state changes.
type VoiceStateUpdate struct {
	GuildID   int64 `json:"guild_id"`
	UserID    int64 `json:"user_id"`
	ChannelID int64 `json:"channel_id"`
	Mute      bool  `json:"mute"`
	Deafen    bool  `json:"deafen"`
	SelfVideo bool  `json:"self_video,omitempty"`
}

// VoiceRegionChanging is dispatched before voice region migration begins.
type VoiceRegionChanging struct {
	ChannelID int64  `json:"channel_id"`
	Region    string `json:"region"`
	DelayMs   int    `json:"delay_ms"`
}

// GuildMemberStartStream is dispatched when a member starts streaming.
type GuildMemberStartStream struct {
	GuildID   int64        `json:"guild_id"`
	ChannelID int64        `json:"channel_id"`
	UserID    int64        `json:"user_id"`
	Stream    ActiveStream `json:"stream"`
}

// GuildMemberStopStream is dispatched when a member stops streaming.
type GuildMemberStopStream struct {
	GuildID   int64  `json:"guild_id"`
	ChannelID int64  `json:"channel_id"`
	UserID    int64  `json:"user_id"`
	StreamID  int64  `json:"stream_id"`
	Reason    string `json:"reason,omitempty"`
}

// GuildStreamsRebind is dispatched when stream viewers should reconnect.
type GuildStreamsRebind struct {
	GuildID   int64   `json:"guild_id"`
	ChannelID int64   `json:"channel_id"`
	StreamIDs []int64 `json:"stream_ids"`
	JitterMs  int     `json:"jitter_ms,omitempty"`
}

// GuildChannelMessage is a compact channel message notification.
type GuildChannelMessage struct {
	GuildID   *int64 `json:"guild_id"`
	ChannelID int64  `json:"channel_id"`
	MessageID int64  `json:"message_id"`
}

// ChannelUserTyping is dispatched when a user starts typing.
type ChannelUserTyping struct {
	GuildID   *int64 `json:"guild_id,omitempty"`
	ChannelID int64  `json:"channel_id"`
	UserID    int64  `json:"user_id"`
}

// Mention is dispatched when a mention is created.
type Mention struct {
	GuildID   *int64 `json:"guild_id"`
	ChannelID int64  `json:"channel_id"`
	MessageID int64  `json:"message_id"`
	AuthorID  int64  `json:"author_id"`
	Type      int    `json:"type"`
}

// UserBrief is a lightweight user payload used in DM events.
type UserBrief struct {
	ID            int64       `json:"id"`
	Name          string      `json:"name"`
	Discriminator string      `json:"discriminator"`
	Avatar        *int64      `json:"avatar,omitempty"`
	AvatarData    *AvatarData `json:"avatar_data,omitempty"`
}

// DMMessage is dispatched when the bot receives a direct message.
type DMMessage struct {
	ChannelID int64     `json:"channel_id"`
	MessageID int64     `json:"message_id"`
	From      UserBrief `json:"from"`
	Message   *Message  `json:"message,omitempty"`
}

// UpdateReadState is dispatched when a user's read state changes.
type UpdateReadState struct {
	ChannelID int64 `json:"channel_id"`
	MessageID int64 `json:"message_id"`
}

// UpdateUserSettings is dispatched when user settings change.
type UpdateUserSettings struct {
	Settings json.RawMessage `json:"settings"`
}

// IncomingFriendRequest is dispatched when a friend request is received.
type IncomingFriendRequest struct {
	From UserBrief `json:"from"`
}

// FriendAdded is dispatched when a friendship is established.
type FriendAdded struct {
	Friend UserBrief `json:"friend"`
}

// FriendRemoved is dispatched when a friendship is removed.
type FriendRemoved struct {
	Friend UserBrief `json:"friend"`
}

// UpdateUser is dispatched when a user's public profile changes.
type UpdateUser struct {
	User User `json:"user"`
}

// UserAuthRevoked is dispatched when user auth sessions are revoked.
type UserAuthRevoked struct {
	SessionVersion int64 `json:"session_version"`
}

// DMCallSummary describes a DM call state snapshot.
type DMCallSummary struct {
	CallID       int64           `json:"call_id"`
	ChannelID    int64           `json:"channel_id"`
	CallerID     int64           `json:"caller_id"`
	RecipientID  int64           `json:"recipient_id"`
	Region       string          `json:"region,omitempty"`
	Participants map[int64]int64 `json:"participants,omitempty"`
	StartedAt    int64           `json:"started_at"`
	SoloSince    int64           `json:"solo_since,omitempty"`
	Dismissed    bool            `json:"dismissed,omitempty"`
}

// DMCallStarted is dispatched when a DM call starts.
type DMCallStarted struct {
	Call DMCallSummary `json:"call"`
}

// DMCallJoined is dispatched when a user joins a DM call.
type DMCallJoined struct {
	Call   DMCallSummary `json:"call"`
	UserID int64         `json:"user_id,omitempty"`
}

// DMCallDeclined is dispatched when a user declines a DM call.
type DMCallDeclined struct {
	Call   DMCallSummary `json:"call"`
	UserID int64         `json:"user_id,omitempty"`
}

// DMCallLeft is dispatched when a user leaves a DM call.
type DMCallLeft struct {
	Call   DMCallSummary `json:"call"`
	UserID int64         `json:"user_id,omitempty"`
}

// DMCallEnded is dispatched when a DM call ends.
type DMCallEnded struct {
	Call   DMCallSummary `json:"call"`
	Reason string        `json:"reason,omitempty"`
}

// DMCallStreamStarted is dispatched when a user starts streaming in a DM call.
type DMCallStreamStarted struct {
	Call   DMCallSummary `json:"call"`
	UserID int64         `json:"user_id"`
	Stream ActiveStream  `json:"stream,omitempty"`
}

// DMCallStreamStopped is dispatched when a user stops streaming in a DM call.
type DMCallStreamStopped struct {
	Call     DMCallSummary `json:"call"`
	UserID   int64         `json:"user_id"`
	StreamID int64         `json:"stream_id,omitempty"`
	Reason   string        `json:"reason,omitempty"`
}

// VoiceMove is an RTC command to move to another voice channel.
type VoiceMove struct {
	Channel  int64  `json:"channel"`
	SFUURL   string `json:"sfu_url"`
	SFUToken string `json:"sfu_token"`
}

// VoiceRebind is an RTC command to reconnect to the current voice channel.
type VoiceRebind struct {
	Channel  int64 `json:"channel"`
	JitterMs int   `json:"jitter_ms,omitempty"`
}

var eventConstructors = map[EventType]func() any{
	EventTypeGatewayReady:             func() any { return &Ready{} },
	EventTypeMessageCreate:            func() any { return &MessageCreate{} },
	EventTypeMessageUpdate:            func() any { return &MessageUpdate{} },
	EventTypeMessageDelete:            func() any { return &MessageDelete{} },
	EventTypeMessageReactionAdd:       func() any { return &MessageReactionAdd{} },
	EventTypeMessageReactionRemove:    func() any { return &MessageReactionRemove{} },
	EventTypeGuildCreate:              func() any { return &GuildCreate{} },
	EventTypeGuildUpdate:              func() any { return &GuildUpdate{} },
	EventTypeGuildDelete:              func() any { return &GuildDelete{} },
	EventTypeChannelCreate:            func() any { return &ChannelCreate{} },
	EventTypeChannelUpdate:            func() any { return &ChannelUpdate{} },
	EventTypeChannelOrderUpdate:       func() any { return &ChannelOrderUpdate{} },
	EventTypeChannelDelete:            func() any { return &ChannelDelete{} },
	EventTypeGuildRoleCreate:          func() any { return &GuildRoleCreate{} },
	EventTypeGuildRoleUpdate:          func() any { return &GuildRoleUpdate{} },
	EventTypeGuildRoleDelete:          func() any { return &GuildRoleDelete{} },
	EventTypeThreadCreate:             func() any { return &ThreadCreate{} },
	EventTypeThreadUpdate:             func() any { return &ThreadUpdate{} },
	EventTypeThreadDelete:             func() any { return &ThreadDelete{} },
	EventTypeGuildEmojiCreate:         func() any { return &GuildEmojiCreate{} },
	EventTypeGuildEmojiUpdate:         func() any { return &GuildEmojiUpdate{} },
	EventTypeGuildEmojiDelete:         func() any { return &GuildEmojiDelete{} },
	EventTypeGuildMemberAdd:           func() any { return &GuildMemberAdd{} },
	EventTypeGuildMemberUpdate:        func() any { return &GuildMemberUpdate{} },
	EventTypeGuildMemberRemove:        func() any { return &GuildMemberRemove{} },
	EventTypeGuildMemberAddRole:       func() any { return &GuildMemberAddRole{} },
	EventTypeGuildMemberRemoveRole:    func() any { return &GuildMemberRemoveRole{} },
	EventTypeGuildMemberJoinVoice:     func() any { return &GuildMemberJoinVoice{} },
	EventTypeGuildMemberLeaveVoice:    func() any { return &GuildMemberLeaveVoice{} },
	EventTypeGuildMemberModeration:    func() any { return &GuildMemberModeration{} },
	EventTypeGuildVoiceRegionChanging: func() any { return &VoiceRegionChanging{} },
	EventTypeVoiceStateUpdate:         func() any { return &VoiceStateUpdate{} },
	EventTypeGuildMemberStartStream:   func() any { return &GuildMemberStartStream{} },
	EventTypeGuildMemberStopStream:    func() any { return &GuildMemberStopStream{} },
	EventTypeGuildStreamsRebind:       func() any { return &GuildStreamsRebind{} },
	EventTypeGuildChannelMessage:      func() any { return &GuildChannelMessage{} },
	EventTypeChannelUserTyping:        func() any { return &ChannelUserTyping{} },
	EventTypeMention:                  func() any { return &Mention{} },
	EventTypeUserUpdateReadState:      func() any { return &UpdateReadState{} },
	EventTypeUserUpdateSettings:       func() any { return &UpdateUserSettings{} },
	EventTypeUserFriendRequest:        func() any { return &IncomingFriendRequest{} },
	EventTypeUserFriendAdded:          func() any { return &FriendAdded{} },
	EventTypeUserFriendRemoved:        func() any { return &FriendRemoved{} },
	EventTypeUserDMMessage:            func() any { return &DMMessage{} },
	EventTypeUserUpdate:               func() any { return &UpdateUser{} },
	EventTypeUserAuthRevoked:          func() any { return &UserAuthRevoked{} },
	EventTypeUserDMCallStarted:        func() any { return &DMCallStarted{} },
	EventTypeUserDMCallJoined:         func() any { return &DMCallJoined{} },
	EventTypeUserDMCallDeclined:       func() any { return &DMCallDeclined{} },
	EventTypeUserDMCallLeft:           func() any { return &DMCallLeft{} },
	EventTypeUserDMCallEnded:          func() any { return &DMCallEnded{} },
	EventTypeUserDMCallStreamStarted:  func() any { return &DMCallStreamStarted{} },
	EventTypeUserDMCallStreamStopped:  func() any { return &DMCallStreamStopped{} },
	EventTypeRTCMoved:                 func() any { return &VoiceMove{} },
	EventTypeRTCServerRebind:          func() any { return &VoiceRebind{} },
}
