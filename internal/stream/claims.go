package stream

import "github.com/FlameInTheDark/gochat/internal/helper"

type Claims struct {
	helper.Claims
	StreamID   int64  `json:"stream_id"`
	ChannelID  int64  `json:"channel_id"`
	GuildID    int64  `json:"guild_id"`
	Role       string `json:"role"`
	SourceType string `json:"source_type,omitempty"`
	AudioMode  string `json:"audio_mode,omitempty"`
}
