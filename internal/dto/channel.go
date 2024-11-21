package dto

import (
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Channel struct {
	Id          int64             `json:"id"`
	Type        model.ChannelType `json:"type"`
	GuildId     *int64            `json:"guild_id,omitempty"`
	Name        string            `json:"name"`
	ParentId    *int64            `json:"parent_id,omitempty"`
	Position    int               `json:"position"`
	Topic       *string           `json:"topic"`
	Permissions *int64            `json:"permissions,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

type ChannelOrder struct {
	Id       int64 `json:"id"`
	Position int   `json:"position"`
}
