package search

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type fakeChannelRepo struct {
	channel model.Channel
	err     error
}

func (f *fakeChannelRepo) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	return f.channel, f.err
}
func (f *fakeChannelRepo) GetChannelsBulk(ctx context.Context, ids []int64) ([]model.Channel, error) {
	return nil, nil
}
func (f *fakeChannelRepo) GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error) {
	return nil, nil
}
func (f *fakeChannelRepo) CreateChannel(ctx context.Context, id int64, name string, channelType model.ChannelType, parent *int64, permissions *int64, private bool) error {
	return nil
}
func (f *fakeChannelRepo) DeleteChannel(ctx context.Context, id int64) error { return nil }
func (f *fakeChannelRepo) RenameChannel(ctx context.Context, id int64, newName string) error {
	return nil
}
func (f *fakeChannelRepo) SetChannelPermissions(ctx context.Context, id int64, permissions int) error {
	return nil
}
func (f *fakeChannelRepo) SetChannelPrivate(ctx context.Context, id int64, private bool) error {
	return nil
}
func (f *fakeChannelRepo) SetChannelTopic(ctx context.Context, id int64, topic *string) error {
	return nil
}
func (f *fakeChannelRepo) SetChannelParent(ctx context.Context, id int64, parent *int64) error {
	return nil
}
func (f *fakeChannelRepo) SetChannelParentBulk(ctx context.Context, id []int64, parent *int64) error {
	return nil
}
func (f *fakeChannelRepo) SetLastMessage(ctx context.Context, id, lastMessage int64) error {
	return nil
}
func (f *fakeChannelRepo) UpdateChannel(ctx context.Context, id int64, parent *int64, private *bool, name, topic *string) (model.Channel, error) {
	return model.Channel{}, nil
}
func (f *fakeChannelRepo) SetChannelVoiceRegion(ctx context.Context, id int64, region *string) error {
	return nil
}
func (f *fakeChannelRepo) GetChannelVoiceRegion(ctx context.Context, id int64) (*string, error) {
	return nil, nil
}

type fakeGuildChannelsRepo struct {
	guildByChannel model.GuildChannel
	err            error
}

func (f *fakeGuildChannelsRepo) AddChannel(ctx context.Context, guildID, channelID int64, channelName string, channelType model.ChannelType, parentID *int64, private bool, position int) error {
	return nil
}
func (f *fakeGuildChannelsRepo) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	return f.guildByChannel, f.err
}
func (f *fakeGuildChannelsRepo) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	return nil, nil
}
func (f *fakeGuildChannelsRepo) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	return f.guildByChannel, f.err
}
func (f *fakeGuildChannelsRepo) RemoveChannel(ctx context.Context, guildID, channelID int64) error {
	return nil
}
func (f *fakeGuildChannelsRepo) SetGuildChannelPosition(ctx context.Context, updates []model.GuildChannelUpdatePosition) error {
	return nil
}
func (f *fakeGuildChannelsRepo) ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error {
	return nil
}
func (f *fakeGuildChannelsRepo) GetGuildsChannelsIDsMany(ctx context.Context, guilds []int64) ([]int64, error) {
	return nil, nil
}

type fakeRoleCheck struct {
	channel      *model.Channel
	guildChannel *model.GuildChannel
	guild        *model.Guild
	canRead      bool
	err          error
	lastGuildID  int64
	lastChannel  int64
	lastUserID   int64
}

func (f *fakeRoleCheck) getUserRoleIDs(ctx context.Context, guildID, userID int64) ([]int64, error) {
	return nil, nil
}
func (f *fakeRoleCheck) ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error) {
	f.lastGuildID = guildID
	f.lastChannel = channelID
	f.lastUserID = userID
	return f.channel, f.guildChannel, f.guild, f.canRead, f.err
}
func (f *fakeRoleCheck) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	return nil, false, nil
}
func (f *fakeRoleCheck) GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error) {
	return 0, nil
}

func TestAuthorizeSearchScopeAllowsDMParticipants(t *testing.T) {
	perm := &fakeRoleCheck{canRead: true}
	e := &entity{
		ch:   &fakeChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeDM}},
		perm: perm,
	}

	guildID, err := e.authorizeSearchScope(context.Background(), 9, 42, nil)
	if err != nil {
		t.Fatalf("authorizeSearchScope returned error: %v", err)
	}
	if guildID != nil {
		t.Fatalf("expected nil guild id for DM search, got %v", *guildID)
	}
	if perm.lastGuildID != 0 {
		t.Fatalf("expected DM permission check without guild id, got %d", perm.lastGuildID)
	}
}

func TestAuthorizeSearchScopeRejectsUnauthorizedDMSearch(t *testing.T) {
	e := &entity{
		ch:   &fakeChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGroupDM}},
		perm: &fakeRoleCheck{canRead: false},
	}

	_, err := e.authorizeSearchScope(context.Background(), 9, 42, nil)
	assertFiberErrorCode(t, err, fiber.StatusForbidden)
}

func TestAuthorizeSearchScopeResolvesGuildForChannelScopedSearch(t *testing.T) {
	gc := &model.GuildChannel{GuildId: 77, ChannelId: 9}
	perm := &fakeRoleCheck{canRead: true, guildChannel: gc}
	e := &entity{
		ch:   &fakeChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGuild}},
		gc:   &fakeGuildChannelsRepo{guildByChannel: *gc},
		perm: perm,
	}

	guildID, err := e.authorizeSearchScope(context.Background(), 9, 42, nil)
	if err != nil {
		t.Fatalf("authorizeSearchScope returned error: %v", err)
	}
	if guildID == nil || *guildID != 77 {
		t.Fatalf("expected resolved guild id 77, got %v", guildID)
	}
	if perm.lastGuildID != 77 {
		t.Fatalf("expected permission check against resolved guild id, got %d", perm.lastGuildID)
	}
}

func TestAuthorizeSearchScopeRejectsDMSearchThroughGuildRoute(t *testing.T) {
	guildID := int64(77)
	e := &entity{
		ch:   &fakeChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeDM}},
		perm: &fakeRoleCheck{canRead: true},
	}

	_, err := e.authorizeSearchScope(context.Background(), 9, 42, &guildID)
	assertFiberErrorCode(t, err, fiber.StatusForbidden)
}

func TestAuthorizeSearchScopeRejectsUnsupportedChannelTypes(t *testing.T) {
	e := &entity{
		ch: &fakeChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGuildVoice}},
	}

	_, err := e.authorizeSearchScope(context.Background(), 9, 42, nil)
	assertFiberErrorCode(t, err, fiber.StatusBadRequest)
}

func assertFiberErrorCode(t *testing.T, err error, want int) {
	t.Helper()
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) {
		t.Fatalf("expected fiber error, got %v", err)
	}
	if fiberErr.Code != want {
		t.Fatalf("expected status %d, got %d", want, fiberErr.Code)
	}
}
