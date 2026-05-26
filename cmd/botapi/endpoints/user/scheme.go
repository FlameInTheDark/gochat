package user

import "github.com/FlameInTheDark/gochat/internal/dto"

type MeResponse struct {
	BotUserId          int64    `json:"bot_user_id"`
	OwnerUserId        int64    `json:"owner_user_id"`
	User               dto.User `json:"user"`
	Description        string   `json:"description"`
	Public             bool     `json:"public"`
	DefaultPermissions int64    `json:"default_permissions"`
}
