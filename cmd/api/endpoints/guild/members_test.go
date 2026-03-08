package guild

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type testPermKey struct {
	guildID int64
	userID  int64
	perm    permissions.RolePermission
}

type fakePermissionChecker struct {
	results map[testPermKey]bool
	calls   []testPermKey
	err     error
}

func (f *fakePermissionChecker) ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error) {
	return nil, nil, nil, false, nil
}

func (f *fakePermissionChecker) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	required := permissions.RolePermission(0)
	if len(perm) > 0 {
		required = perm[0]
	}
	call := testPermKey{guildID: guildID, userID: userID, perm: required}
	f.calls = append(f.calls, call)
	if f.err != nil {
		return nil, false, f.err
	}
	return &model.Guild{Id: guildID}, f.results[call], nil
}

func (f *fakePermissionChecker) GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error) {
	return 0, nil
}

type testMemberKey struct {
	guildID int64
	userID  int64
}

type fakeMemberRepo struct {
	members     map[testMemberKey]bool
	removeCalls []testMemberKey
	addCalls    []testMemberKey
}

func (f *fakeMemberRepo) AddMember(ctx context.Context, userID, guildID int64) error {
	key := testMemberKey{guildID: guildID, userID: userID}
	f.addCalls = append(f.addCalls, key)
	f.members[key] = true
	return nil
}

func (f *fakeMemberRepo) RemoveMember(ctx context.Context, userID, guildID int64) error {
	key := testMemberKey{guildID: guildID, userID: userID}
	f.removeCalls = append(f.removeCalls, key)
	delete(f.members, key)
	return nil
}

func (f *fakeMemberRepo) RemoveMembersByGuild(ctx context.Context, guildID int64) error { return nil }

func (f *fakeMemberRepo) GetMember(ctx context.Context, userId, guildId int64) (model.Member, error) {
	if !f.members[testMemberKey{guildID: guildId, userID: userId}] {
		return model.Member{}, sql.ErrNoRows
	}
	return model.Member{UserId: userId, GuildId: guildId}, nil
}

func (f *fakeMemberRepo) GetMembersList(ctx context.Context, guildId int64, ids []int64) ([]model.Member, error) {
	out := make([]model.Member, 0, len(ids))
	for _, id := range ids {
		if f.members[testMemberKey{guildID: guildId, userID: id}] {
			out = append(out, model.Member{UserId: id, GuildId: guildId})
		}
	}
	return out, nil
}

func (f *fakeMemberRepo) GetGuildMembers(ctx context.Context, guildId int64) ([]model.Member, error) {
	return nil, nil
}

func (f *fakeMemberRepo) IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error) {
	return f.members[testMemberKey{guildID: guildId, userID: userId}], nil
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

type fakeGuildRepo struct {
	guild model.Guild
}

func (f *fakeGuildRepo) GetGuildById(ctx context.Context, id int64) (model.Guild, error) {
	return f.guild, nil
}
func (f *fakeGuildRepo) CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error {
	return nil
}
func (f *fakeGuildRepo) DeleteGuild(ctx context.Context, id int64) error                 { return nil }
func (f *fakeGuildRepo) SetGuildIcon(ctx context.Context, id, icon int64) error          { return nil }
func (f *fakeGuildRepo) SetGuildPublic(ctx context.Context, id int64, public bool) error { return nil }
func (f *fakeGuildRepo) ChangeGuildOwner(ctx context.Context, id, ownerId int64) error   { return nil }
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

type fakeBanRepo struct {
	bans       map[testMemberKey]*string
	banCalls   []model.GuildBan
	unbanCalls []testMemberKey
}

func (f *fakeBanRepo) BanUser(ctx context.Context, guildID, userID int64, reason *string) error {
	if f.bans == nil {
		f.bans = make(map[testMemberKey]*string)
	}
	key := testMemberKey{guildID: guildID, userID: userID}
	f.bans[key] = reason
	f.banCalls = append(f.banCalls, model.GuildBan{GuildId: guildID, UserId: userID, Reason: reason})
	return nil
}

func (f *fakeBanRepo) UnbanUser(ctx context.Context, guildID, userID int64) error {
	key := testMemberKey{guildID: guildID, userID: userID}
	delete(f.bans, key)
	f.unbanCalls = append(f.unbanCalls, key)
	return nil
}

func (f *fakeBanRepo) IsBanned(ctx context.Context, guildID, userID int64) (bool, error) {
	_, ok := f.bans[testMemberKey{guildID: guildID, userID: userID}]
	return ok, nil
}

func (f *fakeBanRepo) GetGuildBans(ctx context.Context, guildID int64) ([]model.GuildBan, error) {
	out := make([]model.GuildBan, 0, len(f.bans))
	for key, reason := range f.bans {
		if key.guildID != guildID {
			continue
		}
		out = append(out, model.GuildBan{GuildId: guildID, UserId: key.userID, Reason: reason})
	}
	return out, nil
}

type fakeTransport struct {
	removed    chan *mqmsg.RemoveGuildMember
	moderation chan *mqmsg.GuildMemberModeration
}

func (f *fakeTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	switch evt := message.(type) {
	case *mqmsg.RemoveGuildMember:
		select {
		case f.removed <- evt:
		default:
		}
	case *mqmsg.GuildMemberModeration:
		select {
		case f.moderation <- evt:
		default:
		}
	}
	return nil
}

func (f *fakeTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}

type fakeUserRepo struct {
	users map[int64]model.User
}

func (f *fakeUserRepo) ModifyUser(ctx context.Context, userId int64, name *string, avatar *int64) error {
	return nil
}
func (f *fakeUserRepo) GetUserById(ctx context.Context, id int64) (model.User, error) {
	return f.users[id], nil
}
func (f *fakeUserRepo) GetUsersList(ctx context.Context, ids []int64) ([]model.User, error) {
	out := make([]model.User, 0, len(ids))
	for _, id := range ids {
		if user, ok := f.users[id]; ok {
			out = append(out, user)
		}
	}
	return out, nil
}
func (f *fakeUserRepo) CreateUser(ctx context.Context, id int64, name string) error      { return nil }
func (f *fakeUserRepo) SetUserAvatar(ctx context.Context, id, attachmentId int64) error  { return nil }
func (f *fakeUserRepo) SetUsername(ctx context.Context, id, name string) error           { return nil }
func (f *fakeUserRepo) SetUserBlocked(ctx context.Context, id int64, blocked bool) error { return nil }
func (f *fakeUserRepo) SetUploadLimit(ctx context.Context, id int64, uploadLimit int64) error {
	return nil
}

type fakeDiscriminatorRepo struct {
	discriminators map[int64]string
}

func (f *fakeDiscriminatorRepo) CreateDiscriminator(ctx context.Context, userId int64, discriminator string) error {
	return nil
}
func (f *fakeDiscriminatorRepo) GetDiscriminatorByUserId(ctx context.Context, userId int64) (model.Discriminator, error) {
	return model.Discriminator{UserId: userId, Discriminator: f.discriminators[userId]}, nil
}
func (f *fakeDiscriminatorRepo) GetUserIdByDiscriminator(ctx context.Context, discriminator string) (model.Discriminator, error) {
	return model.Discriminator{}, nil
}
func (f *fakeDiscriminatorRepo) GetDiscriminatorsByUserIDs(ctx context.Context, userIDs []int64) ([]model.Discriminator, error) {
	out := make([]model.Discriminator, 0, len(userIDs))
	for _, id := range userIDs {
		out = append(out, model.Discriminator{UserId: id, Discriminator: f.discriminators[id]})
	}
	return out, nil
}

type fakeInviteRepo struct {
	invite model.GuildInvite
}

func (f *fakeInviteRepo) CreateInvite(ctx context.Context, code string, inviteID, guildID, authorID int64, expiresAt int64) (model.GuildInvite, error) {
	return model.GuildInvite{}, nil
}
func (f *fakeInviteRepo) GetGuildInvites(ctx context.Context, guildID int64) ([]model.GuildInvite, error) {
	return nil, nil
}
func (f *fakeInviteRepo) DeleteInviteByCode(ctx context.Context, guildID int64, code string) error {
	return nil
}
func (f *fakeInviteRepo) DeleteInviteByID(ctx context.Context, guildID, inviteID int64) error {
	return nil
}
func (f *fakeInviteRepo) FetchInvite(ctx context.Context, code string) (model.GuildInvite, error) {
	return f.invite, nil
}

func newGuildTestApp(t *testing.T, userID int64, path string, handler fiber.Handler) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.All(path, func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: userID}})
		return handler(c)
	})
	return app
}

func TestKickMemberRemovesMemberAndSendsEvents(t *testing.T) {
	transport := &fakeTransport{removed: make(chan *mqmsg.RemoveGuildMember, 1), moderation: make(chan *mqmsg.GuildMemberModeration, 1)}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true, {guildID: 1, userID: 11}: true}}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermMembershipKickMembers}: true,
		{guildID: 1, userID: 11, perm: permissions.PermAdministrator}:         false,
	}}
	e := &entity{g: &fakeGuildRepo{guild: model.Guild{Id: 1, OwnerId: 99}}, memb: members, perm: perms, mqt: transport}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/member/:user_id/kick", e.KickMember)

	req := httptest.NewRequest("POST", "/guild/1/member/11/kick", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if len(members.removeCalls) != 1 || members.removeCalls[0] != (testMemberKey{guildID: 1, userID: 11}) {
		t.Fatalf("unexpected remove calls: %#v", members.removeCalls)
	}

	select {
	case evt := <-transport.removed:
		if evt.GuildId != 1 || evt.UserId != 11 {
			t.Fatalf("unexpected guild removal event: %#v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected guild removal event")
	}

	select {
	case evt := <-transport.moderation:
		if evt.GuildId != 1 || evt.UserId != 11 || evt.ActorId != 10 || evt.Action != mqmsg.GuildMemberModerationKick {
			t.Fatalf("unexpected moderation event: %#v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected guild moderation event")
	}
}

func TestKickMemberRejectsGuildOwner(t *testing.T) {
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true}}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{{guildID: 1, userID: 10, perm: permissions.PermMembershipKickMembers}: true}}
	e := &entity{g: &fakeGuildRepo{guild: model.Guild{Id: 1, OwnerId: 11}}, memb: members, perm: perms}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/member/:user_id/kick", e.KickMember)

	req := httptest.NewRequest("POST", "/guild/1/member/11/kick", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotAcceptable {
		t.Fatalf("expected 406, got %d", resp.StatusCode)
	}
	if len(members.removeCalls) != 0 {
		t.Fatalf("expected no removals, got %#v", members.removeCalls)
	}
}

func TestBanMemberSendsModerationEventWithReason(t *testing.T) {
	reason := "spam links"
	transport := &fakeTransport{removed: make(chan *mqmsg.RemoveGuildMember, 1), moderation: make(chan *mqmsg.GuildMemberModeration, 1)}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true, {guildID: 1, userID: 11}: true}}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermMembershipBanMembers}: true,
		{guildID: 1, userID: 11, perm: permissions.PermAdministrator}:        false,
	}}
	bans := &fakeBanRepo{bans: map[testMemberKey]*string{}}
	e := &entity{g: &fakeGuildRepo{guild: model.Guild{Id: 1, OwnerId: 99}}, memb: members, perm: perms, ban: bans, mqt: transport}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/member/:user_id/ban", e.BanMember)

	req := httptest.NewRequest("POST", "/guild/1/member/11/ban", strings.NewReader(`{"reason":"spam links"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	select {
	case evt := <-transport.removed:
		if evt.GuildId != 1 || evt.UserId != 11 {
			t.Fatalf("unexpected guild removal event: %#v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected guild removal event")
	}

	select {
	case evt := <-transport.moderation:
		if evt.GuildId != 1 || evt.UserId != 11 || evt.ActorId != 10 || evt.Action != mqmsg.GuildMemberModerationBan || evt.Reason == nil || *evt.Reason != reason {
			t.Fatalf("unexpected moderation event: %#v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected guild moderation event")
	}
}

func TestBanMemberRejectsAdministratorForNonOwner(t *testing.T) {
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true, {guildID: 1, userID: 11}: true}}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermMembershipBanMembers}: true,
		{guildID: 1, userID: 11, perm: permissions.PermAdministrator}:        true,
	}}
	bans := &fakeBanRepo{bans: map[testMemberKey]*string{}}
	e := &entity{g: &fakeGuildRepo{guild: model.Guild{Id: 1, OwnerId: 99}}, memb: members, perm: perms, ban: bans}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/member/:user_id/ban", e.BanMember)

	req := httptest.NewRequest("POST", "/guild/1/member/11/ban", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotAcceptable {
		t.Fatalf("expected 406, got %d", resp.StatusCode)
	}
	if len(bans.banCalls) != 0 {
		t.Fatalf("expected no bans, got %#v", bans.banCalls)
	}
	if len(members.removeCalls) != 0 {
		t.Fatalf("expected no removals, got %#v", members.removeCalls)
	}
}

func TestUnbanMemberClearsBanAndSendsEvent(t *testing.T) {
	transport := &fakeTransport{removed: make(chan *mqmsg.RemoveGuildMember, 1), moderation: make(chan *mqmsg.GuildMemberModeration, 1)}
	bans := &fakeBanRepo{bans: map[testMemberKey]*string{{guildID: 1, userID: 11}: nil}}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true}}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermMembershipBanMembers}: true,
		{guildID: 1, userID: 11, perm: permissions.PermAdministrator}:        false,
	}}
	e := &entity{g: &fakeGuildRepo{guild: model.Guild{Id: 1, OwnerId: 99}}, memb: members, perm: perms, ban: bans, mqt: transport}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/member/:user_id/ban", e.UnbanMember)

	req := httptest.NewRequest("DELETE", "/guild/1/member/11/ban", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if len(bans.unbanCalls) != 1 || bans.unbanCalls[0] != (testMemberKey{guildID: 1, userID: 11}) {
		t.Fatalf("unexpected unban calls: %#v", bans.unbanCalls)
	}

	select {
	case evt := <-transport.moderation:
		if evt.GuildId != 1 || evt.UserId != 11 || evt.ActorId != 10 || evt.Action != mqmsg.GuildMemberModerationUnban || evt.Reason != nil {
			t.Fatalf("unexpected moderation event: %#v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected guild moderation event")
	}

	select {
	case evt := <-transport.removed:
		t.Fatalf("did not expect removal event on unban: %#v", evt)
	default:
	}
}

func TestGetBansReturnsUsersAndReasons(t *testing.T) {
	reason := "spam"
	bans := &fakeBanRepo{bans: map[testMemberKey]*string{{guildID: 1, userID: 11}: &reason}}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true}}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{{guildID: 1, userID: 10, perm: permissions.PermMembershipBanMembers}: true}}
	e := &entity{
		g:    &fakeGuildRepo{guild: model.Guild{Id: 1, OwnerId: 99}},
		memb: members,
		perm: perms,
		ban:  bans,
		user: &fakeUserRepo{users: map[int64]model.User{11: {Id: 11, Name: "banned-user"}}},
		disc: &fakeDiscriminatorRepo{discriminators: map[int64]string{11: "1234"}},
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/bans", e.GetBans)

	req := httptest.NewRequest("GET", "/guild/1/bans", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got []dto.GuildBan
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 ban, got %d", len(got))
	}
	if got[0].User.Id != 11 || got[0].User.Name != "banned-user" || got[0].Reason == nil || *got[0].Reason != reason {
		t.Fatalf("unexpected ban payload: %#v", got[0])
	}
}

func TestAcceptInviteRejectsBannedUser(t *testing.T) {
	bans := &fakeBanRepo{bans: map[testMemberKey]*string{{guildID: 1, userID: 10}: nil}}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{}}
	e := &entity{
		ban:  bans,
		memb: members,
		inv:  &fakeInviteRepo{invite: model.GuildInvite{GuildId: 1}},
	}
	app := newGuildTestApp(t, 10, "/guild/invites/accept/:invite_code", e.AcceptInvite)

	req := httptest.NewRequest("POST", "/guild/invites/accept/ABCDEFGH", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if len(members.addCalls) != 0 {
		t.Fatalf("expected no member additions, got %#v", members.addCalls)
	}
}
