package guild

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type fakeCreateGuildChannelsRepo struct {
	guildChannels    map[int64]model.GuildChannel
	addCalls         []fakeCreateGuildChannelCall
	setPositionCalls [][]model.GuildChannelUpdatePosition
}

type fakeCreateGuildChannelCall struct {
	guildID   int64
	channelID int64
	name      string
	typ       model.ChannelType
	parentID  *int64
	private   bool
	position  int
}

func (f *fakeCreateGuildChannelsRepo) AddChannel(ctx context.Context, guildID, channelID int64, channelName string, channelType model.ChannelType, parentID *int64, private bool, position int, topic *string, creatorID *int64, closed bool) error {
	var parentCopy *int64
	if parentID != nil {
		parentValue := *parentID
		parentCopy = &parentValue
	}
	f.addCalls = append(f.addCalls, fakeCreateGuildChannelCall{
		guildID:   guildID,
		channelID: channelID,
		name:      channelName,
		typ:       channelType,
		parentID:  parentCopy,
		private:   private,
		position:  position,
	})
	if f.guildChannels == nil {
		f.guildChannels = make(map[int64]model.GuildChannel)
	}
	f.guildChannels[channelID] = model.GuildChannel{
		GuildId:   guildID,
		ChannelId: channelID,
		Position:  position,
	}
	return nil
}

func (f *fakeCreateGuildChannelsRepo) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	guildChannel, ok := f.guildChannels[channelID]
	if !ok || guildChannel.GuildId != guildID {
		return model.GuildChannel{}, sql.ErrNoRows
	}
	return guildChannel, nil
}

func (f *fakeCreateGuildChannelsRepo) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	guildChannels := make([]model.GuildChannel, 0, len(f.guildChannels))
	for _, guildChannel := range f.guildChannels {
		if guildChannel.GuildId == guildID {
			guildChannels = append(guildChannels, guildChannel)
		}
	}
	sort.Slice(guildChannels, func(i, j int) bool {
		if guildChannels[i].Position == guildChannels[j].Position {
			return guildChannels[i].ChannelId < guildChannels[j].ChannelId
		}
		return guildChannels[i].Position < guildChannels[j].Position
	})
	return guildChannels, nil
}

func (f *fakeCreateGuildChannelsRepo) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	guildChannel, ok := f.guildChannels[channelID]
	if !ok {
		return model.GuildChannel{}, sql.ErrNoRows
	}
	return guildChannel, nil
}

func (f *fakeCreateGuildChannelsRepo) GetGuildChannelsByChannelIDs(ctx context.Context, channelIDs []int64) ([]model.GuildChannel, error) {
	result := make([]model.GuildChannel, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		if guildChannel, ok := f.guildChannels[channelID]; ok {
			result = append(result, guildChannel)
		}
	}
	return result, nil
}

func (f *fakeCreateGuildChannelsRepo) RemoveChannel(ctx context.Context, guildID, channelID int64) error {
	delete(f.guildChannels, channelID)
	return nil
}

func (f *fakeCreateGuildChannelsRepo) SetGuildChannelPosition(ctx context.Context, updates []model.GuildChannelUpdatePosition) error {
	copied := make([]model.GuildChannelUpdatePosition, len(updates))
	copy(copied, updates)
	f.setPositionCalls = append(f.setPositionCalls, copied)

	for _, update := range updates {
		guildChannel := f.guildChannels[update.ChannelId]
		guildChannel.Position = update.Position
		f.guildChannels[update.ChannelId] = guildChannel
	}
	return nil
}

func (f *fakeCreateGuildChannelsRepo) ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error {
	return nil
}

func (f *fakeCreateGuildChannelsRepo) GetGuildsChannelsIDsMany(ctx context.Context, guilds []int64) ([]int64, error) {
	return nil, nil
}

type fakeCreateChannelRepo struct {
	channels map[int64]model.Channel
}

func (f *fakeCreateChannelRepo) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	channel, ok := f.channels[id]
	if !ok {
		return model.Channel{}, sql.ErrNoRows
	}
	return channel, nil
}

func (f *fakeCreateChannelRepo) GetChannelsBulk(ctx context.Context, ids []int64) ([]model.Channel, error) {
	channels := make([]model.Channel, 0, len(ids))
	for _, id := range ids {
		if channel, ok := f.channels[id]; ok {
			channels = append(channels, channel)
		}
	}
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Id < channels[j].Id
	})
	return channels, nil
}

func (f *fakeCreateChannelRepo) GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error) {
	return nil, nil
}

func (f *fakeCreateChannelRepo) GetChannelMessagePosition(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (f *fakeCreateChannelRepo) CreateChannel(ctx context.Context, id int64, name string, channelType model.ChannelType, parent *int64, permissions *int64, private bool) error {
	return nil
}

func (f *fakeCreateChannelRepo) DeleteChannel(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeCreateChannelRepo) RenameChannel(ctx context.Context, id int64, newName string) error {
	return nil
}

func (f *fakeCreateChannelRepo) SetChannelPermissions(ctx context.Context, id int64, permissions int) error {
	return nil
}

func (f *fakeCreateChannelRepo) SetChannelPrivate(ctx context.Context, id int64, private bool) error {
	return nil
}

func (f *fakeCreateChannelRepo) SetChannelTopic(ctx context.Context, id int64, topic *string) error {
	return nil
}

func (f *fakeCreateChannelRepo) SetChannelParent(ctx context.Context, id int64, parent *int64) error {
	return nil
}

func (f *fakeCreateChannelRepo) SetChannelParentBulk(ctx context.Context, id []int64, parent *int64) error {
	return nil
}

func (f *fakeCreateChannelRepo) SetLastMessage(ctx context.Context, id, lastMessage int64) error {
	return nil
}

func (f *fakeCreateChannelRepo) AdjustMessageCount(ctx context.Context, id, delta int64) error {
	return nil
}

func (f *fakeCreateChannelRepo) ReserveMessagePositions(ctx context.Context, id, count int64) (int64, error) {
	return 0, nil
}

func (f *fakeCreateChannelRepo) UpdateChannel(ctx context.Context, id int64, parent *int64, private *bool, name, topic *string, closed *bool) (model.Channel, error) {
	return model.Channel{}, nil
}

func (f *fakeCreateChannelRepo) SetChannelVoiceRegion(ctx context.Context, id int64, region *string) error {
	return nil
}

func (f *fakeCreateChannelRepo) GetChannelVoiceRegion(ctx context.Context, id int64) (*string, error) {
	return nil, nil
}

type fakeChannelRolePermRepo struct{}

func (f *fakeChannelRolePermRepo) GetChannelRolePermission(ctx context.Context, channelId, roleId int64) (model.ChannelRolesPermission, error) {
	return model.ChannelRolesPermission{}, nil
}

func (f *fakeChannelRolePermRepo) GetChannelRolePermissions(ctx context.Context, channelId int64) ([]model.ChannelRolesPermission, error) {
	return nil, nil
}

func (f *fakeChannelRolePermRepo) SetChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error {
	return nil
}

func (f *fakeChannelRolePermRepo) UpdateChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error {
	return nil
}

func (f *fakeChannelRolePermRepo) RemoveChannelRolePermission(ctx context.Context, channelId, roleId int64) error {
	return nil
}

func (f *fakeChannelRolePermRepo) GetChannelRolesBulk(ctx context.Context, channelIDs []int64) ([]model.ChannelRoles, error) {
	result := make([]model.ChannelRoles, len(channelIDs))
	for i, channelID := range channelIDs {
		result[i] = model.ChannelRoles{ChannelId: channelID}
	}
	return result, nil
}

type fakeCreateTransport struct {
	createEvents chan *mqmsg.CreateChannel
}

func (f *fakeCreateTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeCreateTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	createEvent, ok := message.(*mqmsg.CreateChannel)
	if !ok {
		return nil
	}
	select {
	case f.createEvents <- createEvent:
	default:
	}
	return nil
}

func (f *fakeCreateTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func TestCreateGuildChannelRequestAcceptsQuotedParentID(t *testing.T) {
	var req CreateGuildChannelRequest
	if err := json.Unmarshal([]byte(`{"name":"test","type":0,"parent_id":"2297450204871262208"}`), &req); err != nil {
		t.Fatalf("expected quoted parent_id to parse, got %v", err)
	}
	if req.ParentId == nil {
		t.Fatal("expected parent_id to be parsed")
	}
	if *req.ParentId != 2297450204871262208 {
		t.Fatalf("unexpected parent_id: %d", *req.ParentId)
	}
}

func TestCreateChannelAtTopShiftsExistingChannels(t *testing.T) {
	guildChannelsRepo := &fakeCreateGuildChannelsRepo{
		guildChannels: map[int64]model.GuildChannel{
			101: {GuildId: 1, ChannelId: 101, Position: 0},
			102: {GuildId: 1, ChannelId: 102, Position: 1},
		},
	}
	cache := &fakeCache{deleteCh: make(chan string, 1)}
	transport := &fakeCreateTransport{createEvents: make(chan *mqmsg.CreateChannel, 1)}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermServerManageChannels}: true,
	}}
	e := &entity{
		gc:    guildChannelsRepo,
		perm:  perms,
		cache: cache,
		mqt:   transport,
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/channel", e.CreateChannel)

	req := httptest.NewRequest("POST", "/guild/1/channel", strings.NewReader(`{"name":"fresh","type":0}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	if len(guildChannelsRepo.addCalls) != 1 {
		t.Fatalf("expected one add call, got %d", len(guildChannelsRepo.addCalls))
	}
	if guildChannelsRepo.addCalls[0].position != 0 {
		t.Fatalf("expected new channel at position 0, got %d", guildChannelsRepo.addCalls[0].position)
	}
	if guildChannelsRepo.guildChannels[101].Position != 1 || guildChannelsRepo.guildChannels[102].Position != 2 {
		t.Fatalf("expected existing channels to shift down, got %#v", guildChannelsRepo.guildChannels)
	}
}

func TestCreateChannelInsideCategoryAcceptsQuotedParentID(t *testing.T) {
	const parentID int64 = 200

	guildChannelsRepo := &fakeCreateGuildChannelsRepo{
		guildChannels: map[int64]model.GuildChannel{
			parentID: {GuildId: 1, ChannelId: parentID, Position: 4},
			201:      {GuildId: 1, ChannelId: 201, Position: 5},
			300:      {GuildId: 1, ChannelId: 300, Position: 8},
		},
	}
	channelRepo := &fakeCreateChannelRepo{
		channels: map[int64]model.Channel{
			parentID: {Id: parentID, Type: model.ChannelTypeGuildCategory, Name: "text"},
			201:      {Id: 201, Type: model.ChannelTypeGuild, Name: "general", ParentID: int64Ptr(parentID)},
			300:      {Id: 300, Type: model.ChannelTypeGuild, Name: "offtopic"},
		},
	}
	cache := &fakeCache{deleteCh: make(chan string, 1)}
	transport := &fakeCreateTransport{createEvents: make(chan *mqmsg.CreateChannel, 1)}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermServerManageChannels}: true,
	}}
	e := &entity{
		gc:    guildChannelsRepo,
		ch:    channelRepo,
		perm:  perms,
		cache: cache,
		mqt:   transport,
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/channel", e.CreateChannel)

	req := httptest.NewRequest("POST", "/guild/1/channel", strings.NewReader(`{"name":"test","type":0,"parent_id":"200"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	if len(guildChannelsRepo.addCalls) != 1 {
		t.Fatalf("expected one add call, got %d", len(guildChannelsRepo.addCalls))
	}
	addCall := guildChannelsRepo.addCalls[0]
	if addCall.parentID == nil || *addCall.parentID != parentID {
		t.Fatalf("expected parent_id %d to be preserved, got %#v", parentID, addCall.parentID)
	}
	if addCall.position != 5 {
		t.Fatalf("expected new category child at position 5, got %d", addCall.position)
	}
	if guildChannelsRepo.guildChannels[201].Position != 6 || guildChannelsRepo.guildChannels[300].Position != 9 {
		t.Fatalf("expected existing channels after category to shift, got %#v", guildChannelsRepo.guildChannels)
	}

	select {
	case event := <-transport.createEvents:
		if event.Channel.ParentId == nil || *event.Channel.ParentId != parentID {
			t.Fatalf("expected create event parent_id %d, got %#v", parentID, event.Channel.ParentId)
		}
		if event.Channel.Position != 5 {
			t.Fatalf("expected create event position 5, got %d", event.Channel.Position)
		}
	case <-time.After(time.Second):
		t.Fatal("expected create event")
	}

	select {
	case deletedKey := <-cache.deleteCh:
		if deletedKey != "guild:1:channels" {
			t.Fatalf("unexpected deleted cache key: %s", deletedKey)
		}
	case <-time.After(time.Second):
		t.Fatal("expected channels cache invalidation")
	}
}

func TestGetChannelsPreservesStoredParentAndSortsByPosition(t *testing.T) {
	const parentID int64 = 200

	cache := &fakeCache{jsonValues: map[string][]byte{}}
	guildChannelsRepo := &fakeCreateGuildChannelsRepo{
		guildChannels: map[int64]model.GuildChannel{
			100:      {GuildId: 1, ChannelId: 100, Position: 0},
			parentID: {GuildId: 1, ChannelId: parentID, Position: 4},
			201:      {GuildId: 1, ChannelId: 201, Position: 5},
		},
	}
	channelRepo := &fakeCreateChannelRepo{
		channels: map[int64]model.Channel{
			100:      {Id: 100, Type: model.ChannelTypeGuild, Name: "welcome"},
			parentID: {Id: parentID, Type: model.ChannelTypeGuildCategory, Name: "text"},
			201:      {Id: 201, Type: model.ChannelTypeGuild, Name: "general", ParentID: int64Ptr(parentID)},
		},
	}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true}}
	e := &entity{
		cache: cache,
		gc:    guildChannelsRepo,
		ch:    channelRepo,
		rperm: &fakeChannelRolePermRepo{},
		memb:  members,
		g:     &fakeGuildRepo{guild: model.Guild{Id: 1, Name: "guild"}},
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/channel", e.GetChannels)

	req := httptest.NewRequest("GET", "/guild/1/channel", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got []modelChannelForTest
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 channels, got %#v", got)
	}
	if got[0].ID != 100 || got[1].ID != parentID || got[2].ID != 201 {
		t.Fatalf("unexpected channel order: %#v", got)
	}
	if got[2].ParentID == nil || *got[2].ParentID != parentID {
		t.Fatalf("expected child to keep stored parent_id %d, got %#v", parentID, got[2].ParentID)
	}
}

type modelChannelForTest struct {
	ID       int64  `json:"id"`
	ParentID *int64 `json:"parent_id"`
	Position int    `json:"position"`
}

func int64Ptr(value int64) *int64 {
	return &value
}

func TestValidateParentCategoryRejectsNonCategoryParent(t *testing.T) {
	e := &entity{
		gc: &fakeCreateGuildChannelsRepo{
			guildChannels: map[int64]model.GuildChannel{
				101: {GuildId: 1, ChannelId: 101, Position: 2},
			},
		},
		ch: &fakeCreateChannelRepo{
			channels: map[int64]model.Channel{
				101: {Id: 101, Type: model.ChannelTypeGuild, Name: "not-category"},
			},
		},
	}

	parentID := int64(101)
	_, err := e.validateParentCategory(context.Background(), 1, &parentID)
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) {
		t.Fatalf("expected fiber error, got %v", err)
	}
	if fiberErr.Code != fiber.StatusBadRequest || fiberErr.Message != ErrParentCategoryInvalid {
		t.Fatalf("unexpected parent validation error: %#v", fiberErr)
	}
}
