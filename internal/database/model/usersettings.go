package model

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type UserSettings struct {
	UserId   int64           `db:"user_id"`
	Settings json.RawMessage `db:"settings"`
	Version  int64           `db:"version"`
}

type UserSettingsData struct {
	Language      string                     `json:"language"`
	SelectedGuild int64                      `json:"selected_guild"`
	Appearance    UserSettingsAppearance     `json:"appearance"`
	GuildFolders  []UserSettingsGuildFolders `json:"guild_folders"`
	Guilds        []UserSettingsGuilds       `json:"guilds"`
	FavoriteGifs  []string                   `json:"favorite_gifs"`
}

func (s UserSettingsData) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Appearance),
		validation.Field(&s.FavoriteGifs, validation.Each(is.URL)),
	)
}

type UserSettingsGuilds struct {
	GuildId         int64                     `json:"guild_id"`
	Position        int64                     `json:"position"`
	SelectedChannel int64                     `json:"selected_channel"`
	Notifications   UserSettingsNotifications `json:"notifications"`
}

func (g UserSettingsGuilds) Validate() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.Notifications),
	)
}

type NotificationsType int

const (
	NotificationsAll NotificationsType = iota
	NotificationsMentions
	NotificationsNone
)

type UserSettingsNotifications struct {
	Muted         bool              `json:"muted"`
	Notifications NotificationsType `json:"notifications"`
	Global        bool              `json:"global"`
	Roles         bool              `json:"roles"`
}

func (n UserSettingsNotifications) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Notifications,
			validation.In(NotificationsAll, NotificationsMentions, NotificationsNone).Error("invalid notifications type"),
		),
	)
}

type UserSettingsAppearance struct {
	ColorScheme   string `json:"color_scheme"`
	ChatSpacing   int64  `json:"chat_spacing"`
	ChatFontScale int64  `json:"chat_font_scale"`
}

type UserSettingsGuildFolders struct {
	Name     string  `json:"name"`
	Color    int64   `json:"color"`
	Position int64   `json:"position"`
	Guilds   []int64 `json:"guilds"`
}
