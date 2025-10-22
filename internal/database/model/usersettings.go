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
	Language       string                     `json:"language"`
	Appearance     UserSettingsAppearance     `json:"appearance"`
	GuildFolders   []UserSettingsGuildFolders `json:"guild_folders"`
	Guilds         []UserSettingsGuilds       `json:"guilds"`
	FavoriteGifs   []string                   `json:"favorite_gifs"`
	ForcedPresence string                     `json:"forced_presence"`
	Status         Status                     `json:"status"`
	DMChannels     []UserDMChannels           `json:"dm_channels"`
	Devices        Devices                    `json:"devices"`
}

func (s UserSettingsData) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Appearance),
		validation.Field(&s.Status),
		validation.Field(&s.Devices),
		validation.Field(&s.FavoriteGifs, validation.Each(is.URL)),
	)
}

type Devices struct {
	AudioInputDevice    string  `json:"audio_input_device"`
	AudioOutputDevice   string  `json:"audio_output_device"`
	VideoDevice         string  `json:"video_device"`
	NoiseSuppression    bool    `json:"noise_suppression"`
	EchoCancellation    bool    `json:"echo_cancellation"`
	AudioInputLevel     float64 `json:"audio_input_level"`
	AudioOutputLevel    float64 `json:"audio_output_level"`
	AudioInputThreshold float64 `json:"audio_input_threshold"`
	AutoGainControl     bool    `json:"auto_gain_control"`
}

func (d Devices) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.AudioInputLevel, validation.Min(0.0), validation.Max(100.0)),
		validation.Field(&d.AudioOutputLevel, validation.Min(0.0), validation.Max(150.0)),
	)
}

type Status struct {
	Status           string `json:"status"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
}

func (s Status) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Status, validation.In("online", "idle", "dnd", "offline")),
		validation.Field(&s.CustomStatusText, validation.Length(0, 255)),
	)
}

type UserDMChannels struct {
	UserId      int64 `json:"user_id"`
	ChannelId   int64 `json:"channel_id"`
	Hidden      bool  `json:"hidden"`
	HiddenAfter int64 `json:"hidden_after"`
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
