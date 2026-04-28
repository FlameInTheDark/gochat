package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache/testutil"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type userSettingsRepoMock struct {
	mu       sync.Mutex
	settings map[int64]model.UserSettings
}

func newUserSettingsRepoMock() *userSettingsRepoMock {
	return &userSettingsRepoMock{
		settings: make(map[int64]model.UserSettings),
	}
}

func (m *userSettingsRepoMock) GetUserSettings(_ context.Context, userId, version int64) (model.UserSettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, ok := m.settings[userId]
	if !ok || current.Version <= version {
		return model.UserSettings{}, nil
	}
	return current, nil
}

func (m *userSettingsRepoMock) SetUserSettings(_ context.Context, userId int64, settings model.UserSettingsData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := model.MarshalStoredUserSettingsData(settings)
	if err != nil {
		return err
	}

	current := m.settings[userId]
	nextVersion := current.Version + 1
	if nextVersion == 0 {
		nextVersion = 1
	}

	m.settings[userId] = model.UserSettings{
		UserId:   userId,
		Settings: data,
		Version:  nextVersion,
	}
	return nil
}

type memberRepoMock struct {
	guilds []model.UserGuild
}

func (m *memberRepoMock) AddMember(context.Context, int64, int64) error     { return nil }
func (m *memberRepoMock) RemoveMember(context.Context, int64, int64) error  { return nil }
func (m *memberRepoMock) RemoveMembersByGuild(context.Context, int64) error { return nil }
func (m *memberRepoMock) GetMember(context.Context, int64, int64) (model.Member, error) {
	return model.Member{}, nil
}
func (m *memberRepoMock) GetMembersList(context.Context, int64, []int64) ([]model.Member, error) {
	return nil, nil
}
func (m *memberRepoMock) GetGuildMembers(context.Context, int64) ([]model.Member, error) {
	return nil, nil
}
func (m *memberRepoMock) IsGuildMember(context.Context, int64, int64) (bool, error) {
	return false, nil
}
func (m *memberRepoMock) GetUserGuilds(context.Context, int64) ([]model.UserGuild, error) {
	if m.guilds == nil {
		return []model.UserGuild{}, nil
	}
	return m.guilds, nil
}
func (m *memberRepoMock) SetTimeout(context.Context, int64, int64, *time.Time) error { return nil }
func (m *memberRepoMock) CountGuildMembers(context.Context, int64) (int64, error)    { return 0, nil }

type guildRepoMock struct {
	guilds []model.Guild
}

func (m *guildRepoMock) GetGuildById(context.Context, int64) (model.Guild, error) {
	return model.Guild{}, nil
}
func (m *guildRepoMock) CreateGuild(context.Context, int64, string, int64, int64) error {
	return nil
}
func (m *guildRepoMock) DeleteGuild(context.Context, int64) error             { return nil }
func (m *guildRepoMock) SetGuildIcon(context.Context, int64, int64) error     { return nil }
func (m *guildRepoMock) SetGuildPublic(context.Context, int64, bool) error    { return nil }
func (m *guildRepoMock) ChangeGuildOwner(context.Context, int64, int64) error { return nil }
func (m *guildRepoMock) GetGuildsList(context.Context, []int64) ([]model.Guild, error) {
	if m.guilds == nil {
		return []model.Guild{}, nil
	}
	return m.guilds, nil
}
func (m *guildRepoMock) SetGuildPermissions(context.Context, int64, int64) error { return nil }
func (m *guildRepoMock) UpdateGuild(context.Context, int64, *string, *int64, *bool, *int64) error {
	return nil
}
func (m *guildRepoMock) SetSystemMessagesChannel(context.Context, int64, *int64) error { return nil }

type emojiRepoMock struct{}

func (m *emojiRepoMock) CountActiveGuildEmojis(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *emojiRepoMock) CreatePlaceholder(context.Context, model.GuildEmoji) error { return nil }
func (m *emojiRepoMock) ReusePendingPlaceholder(context.Context, model.GuildEmoji) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) GetGuildEmoji(context.Context, int64, int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) GetEmojiLookup(context.Context, int64) (model.EmojiLookup, error) {
	return model.EmojiLookup{}, nil
}
func (m *emojiRepoMock) ListReadyGuildEmojis(context.Context, int64) ([]model.GuildEmoji, error) {
	return []model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) ListReadyGuildEmojisByGuilds(context.Context, []int64) ([]model.GuildEmoji, error) {
	return []model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) MarkReady(context.Context, int64, int64, bool, int64, int64, int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) Rename(context.Context, int64, int64, string, string) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) Delete(context.Context, int64, int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) DeleteGuildEmojis(context.Context, int64) ([]model.GuildEmoji, error) {
	return []model.GuildEmoji{}, nil
}
func (m *emojiRepoMock) PruneExpired(context.Context, int64) error { return nil }

type readStatesRepoMock struct{}

func (m *readStatesRepoMock) GetReadStates(context.Context, int64) (map[int64]int64, error) {
	return map[int64]int64{}, nil
}
func (m *readStatesRepoMock) GetReadState(context.Context, int64, int64) (int64, error) {
	return 0, nil
}
func (m *readStatesRepoMock) SetReadState(context.Context, int64, int64, int64) error { return nil }
func (m *readStatesRepoMock) SetReadStateMany(context.Context, map[int64]int64, map[int64]int64) error {
	return nil
}

type guildChannelMessagesRepoMock struct{}

func (m *guildChannelMessagesRepoMock) GetChannelsMessages(context.Context, int64) (map[int64]int64, error) {
	return map[int64]int64{}, nil
}
func (m *guildChannelMessagesRepoMock) GetChannelMessage(context.Context, int64, int64) (int64, error) {
	return 0, nil
}
func (m *guildChannelMessagesRepoMock) SetChannelLastMessage(context.Context, int64, int64, int64) error {
	return nil
}
func (m *guildChannelMessagesRepoMock) SetReadStateMany(context.Context, map[int64]int64, map[int64]int64) error {
	return nil
}
func (m *guildChannelMessagesRepoMock) GetChannelsMessagesForGuilds(context.Context, []int64) (map[int64]map[int64]int64, error) {
	return map[int64]map[int64]int64{}, nil
}

type guildChannelsRepoMock struct{}

func (m *guildChannelsRepoMock) AddChannel(context.Context, int64, int64, string, model.ChannelType, *int64, bool, int, *string, *int64, bool) error {
	return nil
}
func (m *guildChannelsRepoMock) GetGuildChannel(context.Context, int64, int64) (model.GuildChannel, error) {
	return model.GuildChannel{}, nil
}
func (m *guildChannelsRepoMock) GetGuildChannels(context.Context, int64) ([]model.GuildChannel, error) {
	return []model.GuildChannel{}, nil
}
func (m *guildChannelsRepoMock) GetGuildByChannel(context.Context, int64) (model.GuildChannel, error) {
	return model.GuildChannel{}, nil
}
func (m *guildChannelsRepoMock) GetGuildChannelsByChannelIDs(context.Context, []int64) ([]model.GuildChannel, error) {
	return []model.GuildChannel{}, nil
}
func (m *guildChannelsRepoMock) RemoveChannel(context.Context, int64, int64) error { return nil }
func (m *guildChannelsRepoMock) SetGuildChannelPosition(context.Context, []model.GuildChannelUpdatePosition) error {
	return nil
}
func (m *guildChannelsRepoMock) ResetGuildChannelPositionBulk(context.Context, []int64, int64) error {
	return nil
}
func (m *guildChannelsRepoMock) GetGuildsChannelsIDsMany(context.Context, []int64) ([]int64, error) {
	return []int64{}, nil
}

type threadMemberRepoMock struct{}

func (m *threadMemberRepoMock) AddThreadMember(context.Context, int64, int64) (model.ThreadMember, error) {
	return model.ThreadMember{}, nil
}
func (m *threadMemberRepoMock) RemoveThreadMember(context.Context, int64, int64) error { return nil }
func (m *threadMemberRepoMock) RemoveThreadMembers(context.Context, int64) error       { return nil }
func (m *threadMemberRepoMock) GetThreadMember(context.Context, int64, int64) (model.ThreadMember, error) {
	return model.ThreadMember{}, nil
}
func (m *threadMemberRepoMock) GetThreadMembers(context.Context, int64) ([]model.ThreadMember, error) {
	return []model.ThreadMember{}, nil
}
func (m *threadMemberRepoMock) GetThreadMembersBulk(context.Context, []int64) ([]model.ThreadMember, error) {
	return []model.ThreadMember{}, nil
}
func (m *threadMemberRepoMock) GetThreadMembersByUser(context.Context, int64, []int64) ([]model.ThreadMember, error) {
	return []model.ThreadMember{}, nil
}
func (m *threadMemberRepoMock) GetUserThreadMembers(context.Context, int64) ([]model.ThreadMember, error) {
	return []model.ThreadMember{}, nil
}

type channelRepoMock struct{}

func (m *channelRepoMock) GetChannel(context.Context, int64) (model.Channel, error) {
	return model.Channel{}, nil
}
func (m *channelRepoMock) GetChannelsBulk(context.Context, []int64) ([]model.Channel, error) {
	return []model.Channel{}, nil
}
func (m *channelRepoMock) GetChannelThreads(context.Context, int64) ([]model.Channel, error) {
	return []model.Channel{}, nil
}
func (m *channelRepoMock) GetChannelMessagePosition(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *channelRepoMock) CreateChannel(context.Context, int64, string, model.ChannelType, *int64, *int64, bool) error {
	return nil
}
func (m *channelRepoMock) DeleteChannel(context.Context, int64) error                  { return nil }
func (m *channelRepoMock) RenameChannel(context.Context, int64, string) error          { return nil }
func (m *channelRepoMock) SetChannelPermissions(context.Context, int64, int) error     { return nil }
func (m *channelRepoMock) SetChannelPrivate(context.Context, int64, bool) error        { return nil }
func (m *channelRepoMock) SetChannelTopic(context.Context, int64, *string) error       { return nil }
func (m *channelRepoMock) SetChannelParent(context.Context, int64, *int64) error       { return nil }
func (m *channelRepoMock) SetChannelParentBulk(context.Context, []int64, *int64) error { return nil }
func (m *channelRepoMock) SetLastMessage(context.Context, int64, int64) error          { return nil }
func (m *channelRepoMock) AdjustMessageCount(context.Context, int64, int64) error      { return nil }
func (m *channelRepoMock) ReserveMessagePositions(context.Context, int64, int64) (int64, error) {
	return 0, nil
}
func (m *channelRepoMock) UpdateChannel(context.Context, int64, *int64, *bool, *string, *string, *bool) (model.Channel, error) {
	return model.Channel{}, nil
}
func (m *channelRepoMock) SetChannelVoiceRegion(context.Context, int64, *string) error { return nil }
func (m *channelRepoMock) GetChannelVoiceRegion(context.Context, int64) (*string, error) {
	return nil, nil
}

func newSettingsTestApp(e *entity) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: 1}})
		return c.Next()
	})
	app.Get("/user/me/settings", e.GetUserSettings)
	app.Post("/user/me/settings", e.SetUserSettings)
	return app
}

func newSettingsTestEntity(repo *userSettingsRepoMock) *entity {
	return &entity{
		uset:         repo,
		member:       &memberRepoMock{},
		guild:        &guildRepoMock{},
		emoji:        &emojiRepoMock{},
		rs:           &readStatesRepoMock{},
		gclm:         &guildChannelMessagesRepoMock{},
		gc:           &guildChannelsRepoMock{},
		tm:           &threadMemberRepoMock{},
		ch:           &channelRepoMock{},
		cache:        testutil.Noop{},
		contentHosts: []string{"https://cdn.example.com"},
	}
}

func TestUserSettingsRoundTripIncludesGuildChannelAndUserNotifications(t *testing.T) {
	app := newSettingsTestApp(newSettingsTestEntity(newUserSettingsRepoMock()))

	req := httptest.NewRequest(http.MethodPost, "/user/me/settings", strings.NewReader(`{
		"status":{"status":"online"},
		"guilds":[
			{
				"guild_id":10,
				"position":1,
				"selected_channel":100,
				"notifications":{
					"muted":true,
					"notifications":2,
					"suppress_user_mentions":true,
					"suppress_role_mentions":true,
					"suppress_everyone_mentions":false,
					"suppress_here_mentions":true
				}
			}
		],
		"channels":[
			{
				"channel_id":"1234",
				"notifications":{
					"muted":false,
					"notifications":1,
					"suppress_user_mentions":false,
					"suppress_role_mentions":true,
					"suppress_everyone_mentions":true,
					"suppress_here_mentions":false
				}
			}
		],
		"users":[
			{
				"user_id":"77",
				"notifications":{
					"muted":false,
					"notifications":1,
					"suppress_user_mentions":false,
					"suppress_role_mentions":false,
					"suppress_everyone_mentions":false,
					"suppress_here_mentions":false
				}
			}
		],
		"dm_channels":[
			{
				"user_id":"88",
				"channel_id":"188",
				"hidden":true,
				"hidden_after":55
			}
		]
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("expected POST request to complete, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected POST status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	getResp, err := app.Test(getReq, -1)
	if err != nil {
		t.Fatalf("expected GET request to complete, got %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected GET status %d, got %d", fiber.StatusOK, getResp.StatusCode)
	}

	var got UserSettingsResponse
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}

	if got.Settings == nil {
		t.Fatal("expected settings payload in GET response")
	}
	if len(got.Settings.Guilds) != 1 {
		t.Fatalf("expected one guild settings entry, got %#v", got.Settings.Guilds)
	}
	if !got.Settings.Guilds[0].Notifications.Muted {
		t.Fatalf("expected guild notifications to stay muted, got %#v", got.Settings.Guilds[0].Notifications)
	}
	if got.Settings.Guilds[0].Notifications.Notifications != model.NotificationsNone {
		t.Fatalf("expected guild notification level %d, got %d", model.NotificationsNone, got.Settings.Guilds[0].Notifications.Notifications)
	}
	if !got.Settings.Guilds[0].Notifications.SuppressUserMentions ||
		!got.Settings.Guilds[0].Notifications.SuppressRoleMentions ||
		got.Settings.Guilds[0].Notifications.SuppressEveryoneMentions ||
		!got.Settings.Guilds[0].Notifications.SuppressHereMentions {
		t.Fatalf("expected guild mention suppression flags to round-trip, got %#v", got.Settings.Guilds[0].Notifications)
	}

	if len(got.Settings.ChannelsSettings) != 1 {
		t.Fatalf("expected one channel settings entry, got %#v", got.Settings.ChannelsSettings)
	}
	if got.Settings.ChannelsSettings[0].ChannelId != 1234 {
		t.Fatalf("expected channel settings entry for channel 1234, got %#v", got.Settings.ChannelsSettings[0])
	}
	if got.Settings.ChannelsSettings[0].Notifications.Notifications != model.NotificationsMentions {
		t.Fatalf("expected channel notification level %d, got %d", model.NotificationsMentions, got.Settings.ChannelsSettings[0].Notifications.Notifications)
	}
	if got.Settings.ChannelsSettings[0].Notifications.SuppressUserMentions ||
		!got.Settings.ChannelsSettings[0].Notifications.SuppressRoleMentions ||
		!got.Settings.ChannelsSettings[0].Notifications.SuppressEveryoneMentions ||
		got.Settings.ChannelsSettings[0].Notifications.SuppressHereMentions {
		t.Fatalf("expected channel mention suppression flags to round-trip, got %#v", got.Settings.ChannelsSettings[0].Notifications)
	}

	if len(got.Settings.UsersSettings) != 1 {
		t.Fatalf("expected one user settings entry, got %#v", got.Settings.UsersSettings)
	}
	if got.Settings.UsersSettings[0].UserId != 77 {
		t.Fatalf("expected user notification entry for user 77, got %#v", got.Settings.UsersSettings[0])
	}
	if got.Settings.UsersSettings[0].Notifications.Notifications != model.NotificationsMentions {
		t.Fatalf("expected user notification level %d, got %d", model.NotificationsMentions, got.Settings.UsersSettings[0].Notifications.Notifications)
	}
	if len(got.Settings.DMChannels) != 1 {
		t.Fatalf("expected one DM channel settings entry, got %#v", got.Settings.DMChannels)
	}
	if got.Settings.DMChannels[0].UserId != 88 || got.Settings.DMChannels[0].ChannelId != 188 {
		t.Fatalf("expected DM channel snowflake IDs to round-trip, got %#v", got.Settings.DMChannels[0])
	}
	if !got.Settings.DMChannels[0].Hidden || got.Settings.DMChannels[0].HiddenAfter != 55 {
		t.Fatalf("expected DM channel settings to round-trip, got %#v", got.Settings.DMChannels[0])
	}
}

func TestGetUserSettingsReturnsEmptyNotificationCollectionsForUI(t *testing.T) {
	app := newSettingsTestApp(newSettingsTestEntity(newUserSettingsRepoMock()))

	req := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("expected GET request to complete, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected GET status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	var got UserSettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}

	if got.Settings == nil {
		t.Fatal("expected settings payload in GET response")
	}
	if got.Settings.Guilds == nil || len(got.Settings.Guilds) != 0 {
		t.Fatalf("expected empty guild settings slice for UI, got %#v", got.Settings.Guilds)
	}
	if got.Settings.UsersSettings == nil || len(got.Settings.UsersSettings) != 0 {
		t.Fatalf("expected empty user settings slice for UI, got %#v", got.Settings.UsersSettings)
	}
}

func TestGetUserSettingsGuildsIncludeSystemChannelID(t *testing.T) {
	const systemChannelID int64 = 777

	repo := newUserSettingsRepoMock()
	e := newSettingsTestEntity(repo)
	e.member = &memberRepoMock{guilds: []model.UserGuild{{GuildId: 10, UserId: 1}}}
	e.guild = &guildRepoMock{guilds: []model.Guild{{
		Id:             10,
		Name:           "guild",
		OwnerId:        99,
		Public:         true,
		Permissions:    123,
		SystemMessages: int64Ptr(systemChannelID),
	}}}
	app := newSettingsTestApp(e)

	req := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("expected GET request to complete, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected GET status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	var got UserSettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}

	if len(got.Guilds) != 1 {
		t.Fatalf("expected one guild metadata entry, got %#v", got.Guilds)
	}
	if got.Guilds[0].SystemChannelId == nil || *got.Guilds[0].SystemChannelId != systemChannelID {
		t.Fatalf("expected system channel id %d, got %#v", systemChannelID, got.Guilds[0].SystemChannelId)
	}
}

func TestGetUserSettingsReturnsNoContentWhenVersionIsCurrent(t *testing.T) {
	repo := newUserSettingsRepoMock()
	repo.settings[1] = model.UserSettings{
		UserId:   1,
		Settings: json.RawMessage(`{"status":{"status":"online"}}`),
		Version:  3,
	}

	app := newSettingsTestApp(&entity{uset: repo})

	req := httptest.NewRequest(http.MethodGet, "/user/me/settings?version=3", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("expected GET request to complete, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected GET status %d, got %d", fiber.StatusNoContent, resp.StatusCode)
	}
}

func TestUserSettingsResolvesDevicesPerKey(t *testing.T) {
	app := newSettingsTestApp(newSettingsTestEntity(newUserSettingsRepoMock()))

	desktopReq := httptest.NewRequest(http.MethodPost, "/user/me/settings", strings.NewReader(`{
		"status":{"status":"online"},
		"devices":{
			"audio_input_device":"desk-mic",
			"audio_output_device":"desk-speakers",
			"video_device":"desk-cam",
			"noise_suppression":true,
			"echo_cancellation":true,
			"audio_input_level":100,
			"audio_output_level":75,
			"input_mode":"push_to_talk",
			"push_to_talk_key":"KeyV"
		}
	}`))
	desktopReq.Header.Set("Content-Type", "application/json")
	desktopReq.Header.Set(userSettingsDeviceKeyHeader, "desktop-browser")

	desktopResp, err := app.Test(desktopReq, -1)
	if err != nil {
		t.Fatalf("expected desktop POST request to complete, got %v", err)
	}
	defer desktopResp.Body.Close()

	if desktopResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected desktop POST status %d, got %d", fiber.StatusOK, desktopResp.StatusCode)
	}

	phoneReq := httptest.NewRequest(http.MethodPost, "/user/me/settings", strings.NewReader(`{
		"status":{"status":"online"},
		"devices":{
			"audio_input_device":"phone-mic",
			"audio_output_device":"phone-speaker",
			"video_device":"phone-cam",
			"noise_suppression":false,
			"echo_cancellation":false,
			"audio_input_level":85,
			"audio_output_level":60
		}
	}`))
	phoneReq.Header.Set("Content-Type", "application/json")
	phoneReq.Header.Set(userSettingsDeviceKeyHeader, "phone-web")

	phoneResp, err := app.Test(phoneReq, -1)
	if err != nil {
		t.Fatalf("expected phone POST request to complete, got %v", err)
	}
	defer phoneResp.Body.Close()

	if phoneResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected phone POST status %d, got %d", fiber.StatusOK, phoneResp.StatusCode)
	}

	getDesktopReq := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	getDesktopReq.Header.Set(userSettingsDeviceKeyHeader, "desktop-browser")
	getDesktopResp, err := app.Test(getDesktopReq, -1)
	if err != nil {
		t.Fatalf("expected desktop GET request to complete, got %v", err)
	}
	defer getDesktopResp.Body.Close()

	var desktopSettings UserSettingsResponse
	if err := json.NewDecoder(getDesktopResp.Body).Decode(&desktopSettings); err != nil {
		t.Fatalf("expected valid desktop JSON response, got %v", err)
	}
	if desktopSettings.Settings == nil {
		t.Fatal("expected desktop settings payload")
	}
	if desktopSettings.Settings.Devices.AudioInputDevice != "desk-mic" {
		t.Fatalf("expected desktop device settings, got %#v", desktopSettings.Settings.Devices)
	}
	if desktopSettings.Settings.Devices.InputMode != "push_to_talk" || desktopSettings.Settings.Devices.PushToTalkKey != "KeyV" {
		t.Fatalf("expected desktop voice mode settings to round-trip, got %#v", desktopSettings.Settings.Devices)
	}
	if desktopSettings.Settings.DevicesByKey["phone-web"].AudioInputDevice != "phone-mic" {
		t.Fatalf("expected phone bucket to remain stored, got %#v", desktopSettings.Settings.DevicesByKey)
	}

	getPhoneReq := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	getPhoneReq.Header.Set(userSettingsDeviceKeyHeader, "phone-web")
	getPhoneResp, err := app.Test(getPhoneReq, -1)
	if err != nil {
		t.Fatalf("expected phone GET request to complete, got %v", err)
	}
	defer getPhoneResp.Body.Close()

	var phoneSettings UserSettingsResponse
	if err := json.NewDecoder(getPhoneResp.Body).Decode(&phoneSettings); err != nil {
		t.Fatalf("expected valid phone JSON response, got %v", err)
	}
	if phoneSettings.Settings == nil {
		t.Fatal("expected phone settings payload")
	}
	if phoneSettings.Settings.Devices.AudioInputDevice != "phone-mic" {
		t.Fatalf("expected phone device settings, got %#v", phoneSettings.Settings.Devices)
	}
}

func TestLegacySettingsUpdateKeepsStoredDeviceBuckets(t *testing.T) {
	repo := newUserSettingsRepoMock()
	repo.settings[1] = model.UserSettings{
		UserId: 1,
		Settings: json.RawMessage(`{
			"status":{"status":"online"},
			"devices":{"audio_input_device":"legacy-mic"},
			"devices_by_key":{
				"desktop-browser":{"audio_input_device":"desk-mic"},
				"phone-web":{"audio_input_device":"phone-mic"}
			}
		}`),
		Version: 1,
	}

	app := newSettingsTestApp(newSettingsTestEntity(repo))

	postReq := httptest.NewRequest(http.MethodPost, "/user/me/settings", strings.NewReader(`{
		"status":{"status":"idle"},
		"devices":{"audio_input_device":"legacy-mic-updated"}
	}`))
	postReq.Header.Set("Content-Type", "application/json")

	postResp, err := app.Test(postReq, -1)
	if err != nil {
		t.Fatalf("expected legacy POST request to complete, got %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected legacy POST status %d, got %d", fiber.StatusOK, postResp.StatusCode)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	getReq.Header.Set(userSettingsDeviceKeyHeader, "phone-web")
	getResp, err := app.Test(getReq, -1)
	if err != nil {
		t.Fatalf("expected GET request to complete, got %v", err)
	}
	defer getResp.Body.Close()

	var got UserSettingsResponse
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}
	if got.Settings == nil {
		t.Fatal("expected settings payload in GET response")
	}
	if got.Settings.Devices.AudioInputDevice != "phone-mic" {
		t.Fatalf("expected stored phone device bucket to survive legacy update, got %#v", got.Settings.Devices)
	}
}

func TestDeviceScopedSettingsUpdateWithoutDevicesKeepsCurrentDeviceBucket(t *testing.T) {
	repo := newUserSettingsRepoMock()
	repo.settings[1] = model.UserSettings{
		UserId: 1,
		Settings: json.RawMessage(`{
			"language":"en",
			"status":{"status":"online"},
			"devices":{"audio_input_device":"legacy-mic"},
			"devices_by_key":{
				"desktop-browser":{
					"audio_input_device":"desk-mic",
					"audio_output_device":"desk-speakers",
					"video_device":"desk-cam",
					"audio_input_level":80,
					"audio_output_level":90,
					"input_mode":"push_to_talk",
					"push_to_talk_key":"KeyV"
				}
			}
		}`),
		Version: 1,
	}

	app := newSettingsTestApp(newSettingsTestEntity(repo))

	postReq := httptest.NewRequest(http.MethodPost, "/user/me/settings", strings.NewReader(`{
		"language":"ru",
		"status":{"status":"idle"}
	}`))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set(userSettingsDeviceKeyHeader, "desktop-browser")

	postResp, err := app.Test(postReq, -1)
	if err != nil {
		t.Fatalf("expected POST request to complete, got %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected POST status %d, got %d", fiber.StatusOK, postResp.StatusCode)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/user/me/settings", nil)
	getReq.Header.Set(userSettingsDeviceKeyHeader, "desktop-browser")
	getResp, err := app.Test(getReq, -1)
	if err != nil {
		t.Fatalf("expected GET request to complete, got %v", err)
	}
	defer getResp.Body.Close()

	var got UserSettingsResponse
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}
	if got.Settings == nil {
		t.Fatal("expected settings payload in GET response")
	}
	if got.Settings.Language != "ru" {
		t.Fatalf("expected language update to persist, got %q", got.Settings.Language)
	}
	if got.Settings.Devices.AudioInputDevice != "desk-mic" ||
		got.Settings.Devices.AudioOutputDevice != "desk-speakers" ||
		got.Settings.Devices.VideoDevice != "desk-cam" {
		t.Fatalf("expected device-scoped bucket to survive unrelated update, got %#v", got.Settings.Devices)
	}
	if got.Settings.Devices.InputMode != "push_to_talk" || got.Settings.Devices.PushToTalkKey != "KeyV" {
		t.Fatalf("expected voice mode settings to survive unrelated update, got %#v", got.Settings.Devices)
	}
}

func TestUserSettingsEvictsLeastRecentlyUpdatedDeviceBuckets(t *testing.T) {
	repo := newUserSettingsRepoMock()
	app := newSettingsTestApp(newSettingsTestEntity(repo))

	postSettings := func(deviceKey, audioInput string) {
		t.Helper()

		req := httptest.NewRequest(http.MethodPost, "/user/me/settings", strings.NewReader(`{
			"status":{"status":"online"},
			"devices":{"audio_input_device":"`+audioInput+`"}
		}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(userSettingsDeviceKeyHeader, deviceKey)

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("expected POST request for %s to complete, got %v", deviceKey, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected POST status %d for %s, got %d", fiber.StatusOK, deviceKey, resp.StatusCode)
		}
	}

	for i := 1; i <= maxStoredDeviceSettingsBuckets; i++ {
		postSettings("device-"+strconv.Itoa(i), "mic-"+strconv.Itoa(i))
	}

	// Touch device-2 again so it stays newer than device-1 when the next bucket arrives.
	postSettings("device-2", "mic-2-refresh")
	postSettings("device-17", "mic-17")

	stored, err := model.UnmarshalStoredUserSettingsData(repo.settings[1].Settings)
	if err != nil {
		t.Fatalf("expected stored settings to unmarshal, got %v", err)
	}

	if len(stored.DevicesByKey) != maxStoredDeviceSettingsBuckets {
		t.Fatalf("expected %d stored device buckets, got %d", maxStoredDeviceSettingsBuckets, len(stored.DevicesByKey))
	}
	if _, ok := stored.DevicesByKey["device-1"]; ok {
		t.Fatalf("expected oldest device bucket to be evicted, got %#v", stored.DevicesByKey)
	}
	if _, ok := stored.DevicesByKey["device-2"]; !ok {
		t.Fatalf("expected refreshed device bucket to remain, got %#v", stored.DevicesByKey)
	}
	if stored.DevicesByKey["device-2"].AudioInputDevice != "mic-2-refresh" {
		t.Fatalf("expected refreshed device settings to persist, got %#v", stored.DevicesByKey["device-2"])
	}
	if _, ok := stored.DevicesByKey["device-17"]; !ok {
		t.Fatalf("expected newest device bucket to be stored, got %#v", stored.DevicesByKey)
	}

	order := stored.DeviceUsageOrder()
	if len(order) != maxStoredDeviceSettingsBuckets {
		t.Fatalf("expected usage order to track %d buckets, got %#v", maxStoredDeviceSettingsBuckets, order)
	}
	if order[0] != "device-3" {
		t.Fatalf("expected device-3 to become the oldest retained bucket, got %#v", order)
	}
	if order[len(order)-1] != "device-17" {
		t.Fatalf("expected newest bucket to be last in usage order, got %#v", order)
	}
}
