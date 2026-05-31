package model

import "time"

type Bot struct {
	BotUserId          int64     `json:"bot_user_id" db:"bot_user_id"`
	OwnerUserId        int64     `json:"owner_user_id" db:"owner_user_id"`
	Description        string    `json:"description" db:"description"`
	Public             bool      `json:"public" db:"public"`
	DefaultPermissions int64     `json:"default_permissions" db:"default_permissions"`
	Disabled           bool      `json:"disabled" db:"disabled"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type BotToken struct {
	Id          int64      `json:"id" db:"id"`
	BotUserId   int64      `json:"bot_user_id" db:"bot_user_id"`
	Name        string     `json:"name" db:"name"`
	TokenHash   string     `json:"-" db:"token_hash"`
	TokenPrefix string     `json:"token_prefix" db:"token_prefix"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type BotInstallGrant struct {
	Id                   int64      `json:"id" db:"id"`
	BotUserId            int64      `json:"bot_user_id" db:"bot_user_id"`
	OwnerUserId          int64      `json:"owner_user_id" db:"owner_user_id"`
	TokenHash            string     `json:"-" db:"token_hash"`
	TokenPrefix          string     `json:"token_prefix" db:"token_prefix"`
	RequestedPermissions int64      `json:"requested_permissions" db:"requested_permissions"`
	ExpiresAt            time.Time  `json:"expires_at" db:"expires_at"`
	MaxUses              int        `json:"max_uses" db:"max_uses"`
	Uses                 int        `json:"uses" db:"uses"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

type BotGuild struct {
	BotUserId          int64     `json:"bot_user_id" db:"bot_user_id"`
	GuildId            int64     `json:"guild_id" db:"guild_id"`
	GrantedPermissions int64     `json:"granted_permissions" db:"granted_permissions"`
	InstallerUserId    int64     `json:"installer_user_id" db:"installer_user_id"`
	GrantId            *int64    `json:"grant_id,omitempty" db:"grant_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}
