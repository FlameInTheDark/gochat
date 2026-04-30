package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type UserSettings struct {
	UserId   int64           `db:"user_id,string"`
	Settings json.RawMessage `db:"settings"`
	Version  int64           `db:"version,string"`
}

type UserSettingsData struct {
	Language         string                     `json:"language"`
	Appearance       UserSettingsAppearance     `json:"appearance"`
	GuildFolders     []UserSettingsGuildFolders `json:"guild_folders"`
	Guilds           []UserSettingsGuilds       `json:"guilds"`
	ChannelsSettings []UserSettingsChannel      `json:"channels"`
	UsersSettings    []UserSettingsUsers        `json:"users"`
	FavoriteGifs     []string                   `json:"favorite_gifs"`
	ForcedPresence   string                     `json:"forced_presence"`
	Status           Status                     `json:"status"`
	DMChannels       []UserDMChannels           `json:"dm_channels"`
	Devices          Devices                    `json:"devices"`
	DevicesByKey     map[string]Devices         `json:"devices_by_key,omitempty"`
	UISounds         UserUISounds               `json:"ui_sounds"`

	deviceUsageOrder []string
}

func (s *UserSettingsData) NormalizeCollections() {
	if s == nil {
		return
	}
	if s.GuildFolders == nil {
		s.GuildFolders = []UserSettingsGuildFolders{}
	}
	if s.Guilds == nil {
		s.Guilds = []UserSettingsGuilds{}
	}
	if s.ChannelsSettings == nil {
		s.ChannelsSettings = []UserSettingsChannel{}
	}
	if s.UsersSettings == nil {
		s.UsersSettings = []UserSettingsUsers{}
	}
	if s.FavoriteGifs == nil {
		s.FavoriteGifs = []string{}
	}
	if s.DMChannels == nil {
		s.DMChannels = []UserDMChannels{}
	}
}

func (s UserSettingsData) Validate() error {
	if err := validation.ValidateStruct(&s,
		validation.Field(&s.Appearance),
		validation.Field(&s.Status),
		validation.Field(&s.Devices),
		validation.Field(&s.FavoriteGifs, validation.Each(is.URL)),
	); err != nil {
		return err
	}

	for key, devices := range s.DevicesByKey {
		if strings.TrimSpace(key) == "" {
			return validation.NewError("VALIDATION_DEVICE_KEY_REQUIRED", "devices_by_key contains an empty key")
		}
		if len(key) > 128 {
			return validation.NewError("VALIDATION_DEVICE_KEY_TOO_LONG", fmt.Sprintf("devices_by_key[%q] is too long", key))
		}
		if err := devices.Validate(); err != nil {
			return fmt.Errorf("devices_by_key[%q]: %w", key, err)
		}
	}

	return nil
}

func (s UserSettingsData) DevicesForKey(deviceKey string) Devices {
	if deviceKey == "" {
		return s.Devices
	}
	if devices, ok := s.DevicesByKey[deviceKey]; ok {
		return devices
	}
	return s.Devices
}

func (s *UserSettingsData) SetDevicesForKey(deviceKey string, devices Devices) {
	if s == nil || deviceKey == "" {
		return
	}
	if s.DevicesByKey == nil {
		s.DevicesByKey = make(map[string]Devices)
	}
	s.DevicesByKey[deviceKey] = devices
}

func (s UserSettingsData) DeviceUsageOrder() []string {
	if len(s.deviceUsageOrder) == 0 {
		return nil
	}

	order := make([]string, len(s.deviceUsageOrder))
	copy(order, s.deviceUsageOrder)
	return order
}

func (s *UserSettingsData) SetDeviceUsageOrder(order []string) {
	if s == nil {
		return
	}
	if len(order) == 0 {
		s.deviceUsageOrder = nil
		return
	}

	s.deviceUsageOrder = append([]string(nil), order...)
}

type Devices struct {
	AudioInputDevice    string  `json:"audio_input_device"`
	AudioOutputDevice   string  `json:"audio_output_device"`
	VideoDevice         string  `json:"video_device"`
	NoiseSuppression    bool    `json:"noise_suppression"`
	DenoiserType        string  `json:"denoiser_type"`
	EchoCancellation    bool    `json:"echo_cancellation"`
	AudioInputLevel     float64 `json:"audio_input_level"`
	AudioOutputLevel    float64 `json:"audio_output_level"`
	AudioInputThreshold float64 `json:"audio_input_threshold"`
	AutoGainControl     bool    `json:"auto_gain_control"`
	InputMode           string  `json:"input_mode,omitempty"`
	PushToTalkKey       string  `json:"push_to_talk_key,omitempty"`
}

func (d Devices) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.AudioInputLevel, validation.Min(0.0), validation.Max(200.0)),
		validation.Field(&d.AudioOutputLevel, validation.Min(0.0), validation.Max(200.0)),
		validation.Field(&d.InputMode, validation.In("", "voice_activity", "push_to_talk")),
		validation.Field(&d.PushToTalkKey, validation.Length(0, 64)),
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
	UserId      helper.StringInt64 `json:"user_id"`
	ChannelId   helper.StringInt64 `json:"channel_id"`
	Hidden      bool               `json:"hidden"`
	HiddenAfter int64              `json:"hidden_after"`
}

type UserSettingsGuilds struct {
	GuildId         helper.StringInt64        `json:"guild_id"`
	Position        int64                     `json:"position"`
	SelectedChannel helper.StringInt64        `json:"selected_channel"`
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
	Muted                    bool              `json:"muted"`
	MutedUntil               *time.Time        `json:"muted_until,omitempty"`
	Notifications            NotificationsType `json:"notifications"`
	SuppressUserMentions     bool              `json:"suppress_user_mentions"`
	SuppressRoleMentions     bool              `json:"suppress_role_mentions"`
	SuppressEveryoneMentions bool              `json:"suppress_everyone_mentions"`
	SuppressHereMentions     bool              `json:"suppress_here_mentions"`
}

func (n UserSettingsNotifications) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.Notifications,
			validation.In(NotificationsAll, NotificationsMentions, NotificationsNone).Error("invalid notifications type"),
		),
	)
}

type UserSettingsAppearance struct {
	ColorScheme   string  `json:"color_scheme"`
	ChatSpacing   float32 `json:"chat_spacing"`
	ChatFontScale float32 `json:"chat_font_scale"`
}

type UserUISounds struct {
	Mute         *bool `json:"mute,omitempty"`
	Deafen       *bool `json:"deafen,omitempty"`
	VoiceChannel *bool `json:"voice_channel,omitempty"`
	Notification *bool `json:"notification,omitempty"`
}

type UserSettingsGuildFolders struct {
	Name     string                  `json:"name"`
	Color    int64                   `json:"color"`
	Position int64                   `json:"position"`
	Guilds   helper.StringInt64Array `json:"guilds"`
}

type UserSettingsChannel struct {
	ChannelId     helper.StringInt64        `json:"channel_id"`
	Notifications UserSettingsNotifications `json:"notifications"`
}

type UserSettingsUsers struct {
	UserId        helper.StringInt64        `json:"user_id"`
	Notifications UserSettingsNotifications `json:"notifications"`
}
