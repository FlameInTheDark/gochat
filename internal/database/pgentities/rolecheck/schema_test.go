package rolecheck

import (
	"context"
	"testing"
	"time"

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
func (f *fakeChannelRepo) GetChannelMessagePosition(ctx context.Context, id int64) (int64, error) {
	return 0, nil
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
func (f *fakeChannelRepo) AdjustMessageCount(ctx context.Context, id, delta int64) error {
	return nil
}
func (f *fakeChannelRepo) ReserveMessagePositions(ctx context.Context, id, count int64) (int64, error) {
	return count, nil
}
func (f *fakeChannelRepo) UpdateChannel(ctx context.Context, id int64, parent *int64, private *bool, name, topic *string, closed *bool) (model.Channel, error) {
	return model.Channel{}, nil
}
func (f *fakeChannelRepo) SetChannelVoiceRegion(ctx context.Context, id int64, region *string) error {
	return nil
}
func (f *fakeChannelRepo) GetChannelVoiceRegion(ctx context.Context, id int64) (*string, error) {
	return nil, nil
}

type fakeGuildRepo struct {
	guild model.Guild
	err   error
}

func (f *fakeGuildRepo) GetGuildById(ctx context.Context, id int64) (model.Guild, error) {
	return f.guild, f.err
}
func (f *fakeGuildRepo) CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error {
	return nil
}
func (f *fakeGuildRepo) DeleteGuild(ctx context.Context, id int64) error { return nil }
func (f *fakeGuildRepo) SetGuildIcon(ctx context.Context, id, icon int64) error {
	return nil
}
func (f *fakeGuildRepo) SetGuildPublic(ctx context.Context, id int64, public bool) error {
	return nil
}
func (f *fakeGuildRepo) ChangeGuildOwner(ctx context.Context, id, ownerId int64) error {
	return nil
}
func (f *fakeGuildRepo) GetGuildsList(ctx context.Context, ids []int64) ([]model.Guild, error) {
	return nil, nil
}
func (f *fakeGuildRepo) SetGuildPermissions(ctx context.Context, id int64, permissions int64) error {
	return nil
}
func (f *fakeGuildRepo) UpdateGuild(ctx context.Context, id int64, name *string, icon *int64, public *bool, permissions *int64) error {
	return nil
}
func (f *fakeGuildRepo) SetSystemMessagesChannel(ctx context.Context, id int64, channelId *int64) error {
	return nil
}

type fakeGuildChannelsRepo struct {
	guildChannel model.GuildChannel
	err          error
}

func (f *fakeGuildChannelsRepo) AddChannel(ctx context.Context, guildID, channelID int64, channelName string, channelType model.ChannelType, parentID *int64, private bool, position int, topic *string, creatorID *int64, closed bool) error {
	return nil
}
func (f *fakeGuildChannelsRepo) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	return f.guildChannel, f.err
}
func (f *fakeGuildChannelsRepo) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	return nil, nil
}
func (f *fakeGuildChannelsRepo) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	return f.guildChannel, f.err
}
func (f *fakeGuildChannelsRepo) GetGuildChannelsByChannelIDs(ctx context.Context, channelIDs []int64) ([]model.GuildChannel, error) {
	return nil, nil
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

type fakeMemberRepo struct {
	isMember bool
	err      error
}

func (f *fakeMemberRepo) AddMember(ctx context.Context, userID, guildID int64) error {
	return nil
}
func (f *fakeMemberRepo) RemoveMember(ctx context.Context, userID, guildID int64) error {
	return nil
}
func (f *fakeMemberRepo) RemoveMembersByGuild(ctx context.Context, guildID int64) error {
	return nil
}
func (f *fakeMemberRepo) GetMember(ctx context.Context, userId, guildId int64) (model.Member, error) {
	return model.Member{}, nil
}
func (f *fakeMemberRepo) GetMembersList(ctx context.Context, guildId int64, ids []int64) ([]model.Member, error) {
	return nil, nil
}
func (f *fakeMemberRepo) GetGuildMembers(ctx context.Context, guildId int64) ([]model.Member, error) {
	return nil, nil
}
func (f *fakeMemberRepo) IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error) {
	return f.isMember, f.err
}
func (f *fakeMemberRepo) GetUserGuilds(ctx context.Context, userId int64) ([]model.UserGuild, error) {
	return nil, nil
}
func (f *fakeMemberRepo) SetTimeout(ctx context.Context, userId, guildId int64, timeout *time.Time) error {
	return nil
}
func (f *fakeMemberRepo) CountGuildMembers(ctx context.Context, guildId int64) (int64, error) {
	return 0, nil
}

func newGuildPermissionEntity(isMember bool) *Entity {
	return &Entity{
		ch: &fakeChannelRepo{
			channel: model.Channel{Id: 9, Type: model.ChannelTypeGuild},
		},
		g: &fakeGuildRepo{
			guild: model.Guild{
				Id:          77,
				OwnerId:     1,
				Permissions: permissions.CreatePermissions(permissions.PermServerViewChannels),
			},
		},
		gc: &fakeGuildChannelsRepo{
			guildChannel: model.GuildChannel{GuildId: 77, ChannelId: 9},
		},
		m: &fakeMemberRepo{isMember: isMember},
	}
}

func TestChannelPermRejectsNonMemberGuildUser(t *testing.T) {
	e := newGuildPermissionEntity(false)

	channel, guildChannel, guild, ok, err := e.ChannelPerm(context.Background(), 77, 9, 42, permissions.PermServerViewChannels)
	if err != nil {
		t.Fatalf("ChannelPerm returned error: %v", err)
	}
	if ok {
		t.Fatal("expected non-member access to be denied")
	}
	if channel != nil || guildChannel != nil || guild != nil {
		t.Fatalf("expected no guild metadata for denied access, got channel=%v guildChannel=%v guild=%v", channel, guildChannel, guild)
	}
}

func TestGetChannelPermissionsRejectsNonMemberGuildUser(t *testing.T) {
	e := newGuildPermissionEntity(false)

	perms, err := e.GetChannelPermissions(context.Background(), 77, 9, 42)
	if err != nil {
		t.Fatalf("GetChannelPermissions returned error: %v", err)
	}
	if perms != 0 {
		t.Fatalf("expected non-member to have zero effective permissions, got %d", perms)
	}
}

func TestGuildPermRejectsNonMemberGuildUser(t *testing.T) {
	e := newGuildPermissionEntity(false)

	guild, ok, err := e.GuildPerm(context.Background(), 77, 42, permissions.PermMembershipCreateInvite)
	if err != nil {
		t.Fatalf("GuildPerm returned error: %v", err)
	}
	if ok {
		t.Fatal("expected non-member guild permission check to fail")
	}
	if guild != nil {
		t.Fatalf("expected nil guild for denied access, got %+v", guild)
	}
}
