package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

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

	data, err := json.Marshal(settings)
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

type guildRepoMock struct{}

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
	return []model.Guild{}, nil
}
func (m *guildRepoMock) SetGuildPermissions(context.Context, int64, int64) error { return nil }
func (m *guildRepoMock) UpdateGuild(context.Context, int64, *string, *int64, *bool, *int64) error {
	return nil
}
func (m *guildRepoMock) SetSystemMessagesChannel(context.Context, int64, *int64) error { return nil }

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
		rs:           &readStatesRepoMock{},
		gclm:         &guildChannelMessagesRepoMock{},
		gc:           &guildChannelsRepoMock{},
		tm:           &threadMemberRepoMock{},
		ch:           &channelRepoMock{},
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
