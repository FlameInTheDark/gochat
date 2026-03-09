package guild

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type fakeRoleRepo struct {
	roles             map[int64]model.Role
	getGuildRolesCall int
	getRolesBulkCall  int
	setPositionCalls  [][]model.RoleUpdatePosition
}

func (f *fakeRoleRepo) GetRoleByID(ctx context.Context, id int64) (model.Role, error) {
	role, ok := f.roles[id]
	if !ok {
		return model.Role{}, errors.New("role not found")
	}
	return role, nil
}

func (f *fakeRoleRepo) GetGuildRoles(ctx context.Context, guildId int64) ([]model.Role, error) {
	f.getGuildRolesCall++
	return f.rolesForGuild(guildId), nil
}

func (f *fakeRoleRepo) GetRolesBulk(ctx context.Context, guildID int64, ids []int64) ([]model.Role, error) {
	f.getRolesBulkCall++
	allowed := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}
	roles := f.rolesForGuild(guildID)
	filtered := make([]model.Role, 0, len(ids))
	for _, role := range roles {
		if _, ok := allowed[role.Id]; ok {
			filtered = append(filtered, role)
		}
	}
	return filtered, nil
}

func (f *fakeRoleRepo) CreateRole(ctx context.Context, id, guildId int64, name string, color int, permissions int64) error {
	f.roles[id] = model.Role{
		Id:          id,
		GuildId:     guildId,
		Name:        name,
		Color:       color,
		Permissions: permissions,
		Position:    len(f.rolesForGuild(guildId)),
	}
	return nil
}

func (f *fakeRoleRepo) RemoveRole(ctx context.Context, id int64) error {
	delete(f.roles, id)
	return nil
}

func (f *fakeRoleRepo) SetRoleColor(ctx context.Context, id int64, color int) error {
	role := f.roles[id]
	role.Color = color
	f.roles[id] = role
	return nil
}

func (f *fakeRoleRepo) SetRoleName(ctx context.Context, id int64, name string) error {
	role := f.roles[id]
	role.Name = name
	f.roles[id] = role
	return nil
}

func (f *fakeRoleRepo) SetRolePermissions(ctx context.Context, id int64, permissions int64) error {
	role := f.roles[id]
	role.Permissions = permissions
	f.roles[id] = role
	return nil
}

func (f *fakeRoleRepo) SetRolePosition(ctx context.Context, updates []model.RoleUpdatePosition) error {
	copied := make([]model.RoleUpdatePosition, len(updates))
	copy(copied, updates)
	f.setPositionCalls = append(f.setPositionCalls, copied)

	for _, update := range updates {
		role := f.roles[update.RoleId]
		role.Position = update.Position
		f.roles[update.RoleId] = role
	}
	return nil
}

func (f *fakeRoleRepo) rolesForGuild(guildId int64) []model.Role {
	roles := make([]model.Role, 0, len(f.roles))
	for _, role := range f.roles {
		if role.GuildId == guildId {
			roles = append(roles, role)
		}
	}
	sort.Slice(roles, func(i, j int) bool {
		if roles[i].Position == roles[j].Position {
			return roles[i].Id < roles[j].Id
		}
		return roles[i].Position < roles[j].Position
	})
	return roles
}

type fakeCache struct {
	jsonValues map[string][]byte
	deleted    []string
	deleteCh   chan string
}

func (f *fakeCache) Set(ctx context.Context, key, val string) error { return nil }
func (f *fakeCache) Get(ctx context.Context, key string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeCache) SetTimed(ctx context.Context, key, val string, ttl int64) error { return nil }
func (f *fakeCache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	return nil
}
func (f *fakeCache) SetInt64(ctx context.Context, key string, val int64) error { return nil }
func (f *fakeCache) SetTTL(ctx context.Context, key string, ttl int64) error   { return nil }
func (f *fakeCache) Incr(ctx context.Context, key string) (int64, error)       { return 0, nil }
func (f *fakeCache) GetInt64(ctx context.Context, key string) (int64, error)   { return 0, nil }
func (f *fakeCache) HGet(ctx context.Context, key, field string) (string, error) {
	return "", nil
}
func (f *fakeCache) HSet(ctx context.Context, key, field, value string) error { return nil }
func (f *fakeCache) HDel(ctx context.Context, key, field string) error        { return nil }
func (f *fakeCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeCache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	return nil
}

func (f *fakeCache) Delete(ctx context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	if f.deleteCh != nil {
		select {
		case f.deleteCh <- key:
		default:
		}
	}
	return nil
}

func (f *fakeCache) SetJSON(ctx context.Context, key string, val interface{}) error {
	if f.jsonValues == nil {
		f.jsonValues = make(map[string][]byte)
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return err
	}
	f.jsonValues[key] = raw
	return nil
}

func (f *fakeCache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error {
	return f.SetJSON(ctx, key, val)
}

func (f *fakeCache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error) {
	if _, ok := f.jsonValues[key]; ok {
		return false, nil
	}
	return true, f.SetJSON(ctx, key, val)
}

func (f *fakeCache) GetJSON(ctx context.Context, key string, v interface{}) error {
	raw, ok := f.jsonValues[key]
	if !ok {
		return errors.New("cache miss")
	}
	return json.Unmarshal(raw, v)
}

type fakeRoleTransport struct {
	roleUpdates chan *mqmsg.UpdateGuildRole
	roleCreates chan *mqmsg.CreateGuildRole
}

func (f *fakeRoleTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeRoleTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	switch evt := message.(type) {
	case *mqmsg.UpdateGuildRole:
		select {
		case f.roleUpdates <- evt:
		default:
		}
	case *mqmsg.CreateGuildRole:
		select {
		case f.roleCreates <- evt:
		default:
		}
	}
	return nil
}

func (f *fakeRoleTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}

type roleOrderPermissionChecker struct {
	ownerID  int64
	admins   map[int64]bool
	managers map[int64]bool
}

func (f *roleOrderPermissionChecker) ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error) {
	return nil, nil, nil, false, nil
}

func (f *roleOrderPermissionChecker) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	if userID == f.ownerID {
		return &model.Guild{Id: guildID, OwnerId: f.ownerID}, true, nil
	}
	if f.admins[userID] || f.managers[userID] {
		return &model.Guild{Id: guildID, OwnerId: f.ownerID}, true, nil
	}
	return nil, false, nil
}

func (f *roleOrderPermissionChecker) GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error) {
	return 0, nil
}

func TestGetGuildRolesUsesCache(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	cachedRoles := []dto.Role{{Id: 15, GuildId: 1, Name: "cached", Position: 3}}
	if err := cache.SetJSON(context.Background(), "guild:1:roles", cachedRoles); err != nil {
		t.Fatalf("unable to seed cache: %v", err)
	}

	roleRepo := &fakeRoleRepo{
		roles: map[int64]model.Role{
			10: {Id: 10, GuildId: 1, Name: "repo", Position: 0},
		},
	}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: 1, userID: 10}: true}}
	e := &entity{role: roleRepo, cache: cache, memb: members}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/roles", e.GetGuildRoles)

	req := httptest.NewRequest("GET", "/guild/1/roles", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got []dto.Role
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if len(got) != 1 || got[0].Name != "cached" || got[0].Position != 3 {
		t.Fatalf("unexpected cached response: %#v", got)
	}
	if roleRepo.getGuildRolesCall != 0 {
		t.Fatalf("expected cache hit without repo call, got %d repo calls", roleRepo.getGuildRolesCall)
	}
}

func TestPatchRoleOrderUpdatesRolesAndInvalidatesCache(t *testing.T) {
	roleRepo := &fakeRoleRepo{
		roles: map[int64]model.Role{
			101: {Id: 101, GuildId: 1, Name: "alpha", Position: 0},
			102: {Id: 102, GuildId: 1, Name: "beta", Position: 1},
			201: {Id: 201, GuildId: 2, Name: "other", Position: 0},
		},
	}
	cache := &fakeCache{deleteCh: make(chan string, 1)}
	transport := &fakeRoleTransport{roleUpdates: make(chan *mqmsg.UpdateGuildRole, 2)}
	perms := &fakePermissionChecker{results: map[testPermKey]bool{
		{guildID: 1, userID: 10, perm: permissions.PermServerManageRoles}: true,
	}}
	e := &entity{role: roleRepo, cache: cache, mqt: transport, perm: perms}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/roles/order", e.PatchRoleOrder)

	body := strings.NewReader(`{"roles":[{"id":"101","position":2},{"id":"102","position":0},{"id":"201","position":5}]}`)
	req := httptest.NewRequest("PATCH", "/guild/1/roles/order", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if len(roleRepo.setPositionCalls) != 1 {
		t.Fatalf("expected one position update call, got %d", len(roleRepo.setPositionCalls))
	}
	gotUpdates := roleRepo.setPositionCalls[0]
	if len(gotUpdates) != 2 {
		t.Fatalf("expected 2 in-guild role updates, got %#v", gotUpdates)
	}
	if roleRepo.roles[101].Position != 2 || roleRepo.roles[102].Position != 0 || roleRepo.roles[201].Position != 0 {
		t.Fatalf("unexpected stored positions: %#v", roleRepo.roles)
	}

	select {
	case key := <-cache.deleteCh:
		if key != "guild:1:roles" {
			t.Fatalf("unexpected deleted cache key: %s", key)
		}
	case <-time.After(time.Second):
		t.Fatal("expected role cache invalidation")
	}

	received := make(map[int64]int)
	for i := 0; i < 2; i++ {
		select {
		case evt := <-transport.roleUpdates:
			received[evt.Role.Id] = evt.Role.Position
		case <-time.After(time.Second):
			t.Fatal("expected role update event")
		}
	}
	if received[101] != 2 || received[102] != 0 {
		t.Fatalf("unexpected role update events: %#v", received)
	}
}

func TestPatchRoleOrderPermissionSources(t *testing.T) {
	roleRepo := &fakeRoleRepo{roles: map[int64]model.Role{}}
	cache := &fakeCache{}
	perms := &roleOrderPermissionChecker{
		ownerID:  99,
		admins:   map[int64]bool{20: true},
		managers: map[int64]bool{30: true},
	}

	tests := []struct {
		name       string
		userID     int64
		wantStatus int
	}{
		{name: "owner allowed", userID: 99, wantStatus: fiber.StatusOK},
		{name: "administrator allowed", userID: 20, wantStatus: fiber.StatusOK},
		{name: "manage roles allowed", userID: 30, wantStatus: fiber.StatusOK},
		{name: "other member denied", userID: 40, wantStatus: fiber.StatusNotAcceptable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &entity{role: roleRepo, cache: cache, perm: perms}
			app := newGuildTestApp(t, tt.userID, "/guild/:guild_id/roles/order", e.PatchRoleOrder)

			req := httptest.NewRequest("PATCH", "/guild/1/roles/order", strings.NewReader(`{"roles":[{"id":"101","position":1}]}`))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}
