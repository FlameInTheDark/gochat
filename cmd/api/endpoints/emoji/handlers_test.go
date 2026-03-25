package emoji

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
)

type fakeEmojiRepo struct {
	lookup model.EmojiLookup
	err    error
}

func (f *fakeEmojiRepo) PruneExpired(ctx context.Context, guildID int64) error { return nil }

func (f *fakeEmojiRepo) CountActiveGuildEmojis(ctx context.Context, guildID int64) (int64, error) {
	return 0, nil
}

func (f *fakeEmojiRepo) CreatePlaceholder(ctx context.Context, emoji model.GuildEmoji) error {
	return nil
}

func (f *fakeEmojiRepo) ReusePendingPlaceholder(ctx context.Context, emoji model.GuildEmoji) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, emojirepo.ErrEmojiNotFound
}

func (f *fakeEmojiRepo) GetGuildEmoji(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, emojirepo.ErrEmojiNotFound
}

func (f *fakeEmojiRepo) GetEmojiLookup(ctx context.Context, emojiID int64) (model.EmojiLookup, error) {
	return f.lookup, f.err
}

func (f *fakeEmojiRepo) ListReadyGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

func (f *fakeEmojiRepo) ListReadyGuildEmojisByGuilds(ctx context.Context, guildIDs []int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

func (f *fakeEmojiRepo) MarkReady(ctx context.Context, guildID, emojiID int64, animated bool, actualFileSize int64, width, height int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) Rename(ctx context.Context, guildID, emojiID int64, name, normalized string) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) Delete(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) DeleteGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

type fakeGuildRepo struct {
	guilds map[int64]model.Guild
}

func (f *fakeGuildRepo) GetGuildById(ctx context.Context, id int64) (model.Guild, error) {
	return f.guilds[id], nil
}

func (f *fakeGuildRepo) CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error {
	return nil
}

func (f *fakeGuildRepo) DeleteGuild(ctx context.Context, id int64) error { return nil }

func (f *fakeGuildRepo) SetGuildIcon(ctx context.Context, id, icon int64) error { return nil }

func (f *fakeGuildRepo) SetGuildPublic(ctx context.Context, id int64, public bool) error { return nil }

func (f *fakeGuildRepo) ChangeGuildOwner(ctx context.Context, id, ownerId int64) error { return nil }

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

type fakeMemberRepo struct {
	members map[[2]int64]bool
	calls   int
}

func (f *fakeMemberRepo) AddMember(ctx context.Context, userID, guildID int64) error { return nil }

func (f *fakeMemberRepo) RemoveMember(ctx context.Context, userID, guildID int64) error { return nil }

func (f *fakeMemberRepo) RemoveMembersByGuild(ctx context.Context, guildID int64) error { return nil }

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
	f.calls++
	return f.members[[2]int64{guildId, userId}], nil
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

func newEmojiTestApp(t *testing.T, userID int64, e *entity) *fiber.App {
	t.Helper()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: userID}})
		return c.Next()
	})

	root := app.Group("/api/v1")
	e.Init(root.Group(e.Name()))
	return app
}

func TestGetInfoReturnsServerNameForPublicGuild(t *testing.T) {
	members := &fakeMemberRepo{members: map[[2]int64]bool{}}
	e := &entity{
		emoji: &fakeEmojiRepo{lookup: model.EmojiLookup{Id: 1, GuildId: 77, Name: "party-cat", Done: true}},
		guild: &fakeGuildRepo{guilds: map[int64]model.Guild{
			77: {Id: 77, Name: "Public Guild", Public: true},
		}},
		member: members,
	}
	app := newEmojiTestApp(t, 15, e)

	req := httptest.NewRequest("GET", "/api/v1/info/emoji/1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got dto.EmojiInfo
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if got.Name != "party-cat" {
		t.Fatalf("unexpected emoji name: %#v", got)
	}
	if got.ServerName == nil || *got.ServerName != "Public Guild" {
		t.Fatalf("expected server name to be returned, got %#v", got.ServerName)
	}
	if got.ServerPrivate {
		t.Fatalf("expected public server flag, got %#v", got)
	}
	if members.calls != 0 {
		t.Fatalf("did not expect membership check for public guild, got %d calls", members.calls)
	}
}

func TestGetInfoReturnsServerNameForPrivateGuildMember(t *testing.T) {
	e := &entity{
		emoji: &fakeEmojiRepo{lookup: model.EmojiLookup{Id: 1, GuildId: 77, Name: "party-cat", Done: true}},
		guild: &fakeGuildRepo{guilds: map[int64]model.Guild{
			77: {Id: 77, Name: "Private Guild", Public: false},
		}},
		member: &fakeMemberRepo{members: map[[2]int64]bool{{77, 15}: true}},
	}
	app := newEmojiTestApp(t, 15, e)

	req := httptest.NewRequest("GET", "/api/v1/info/emoji/1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var got dto.EmojiInfo
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if got.ServerName == nil || *got.ServerName != "Private Guild" {
		t.Fatalf("expected private guild member to receive server name, got %#v", got.ServerName)
	}
	if !got.ServerPrivate {
		t.Fatalf("expected private server flag, got %#v", got)
	}
}

func TestGetInfoHidesServerNameForPrivateGuildNonMember(t *testing.T) {
	e := &entity{
		emoji: &fakeEmojiRepo{lookup: model.EmojiLookup{Id: 1, GuildId: 77, Name: "party-cat", Done: true}},
		guild: &fakeGuildRepo{guilds: map[int64]model.Guild{
			77: {Id: 77, Name: "Private Guild", Public: false},
		}},
		member: &fakeMemberRepo{members: map[[2]int64]bool{}},
	}
	app := newEmojiTestApp(t, 15, e)

	req := httptest.NewRequest("GET", "/api/v1/info/emoji/1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if raw["name"] != "party-cat" {
		t.Fatalf("unexpected emoji name payload: %#v", raw)
	}
	if _, ok := raw["server_name"]; ok {
		t.Fatalf("expected server_name to be hidden, got %#v", raw["server_name"])
	}
	if raw["server_private"] != true {
		t.Fatalf("expected private server flag in payload, got %#v", raw)
	}
}

func TestGetInfoReturnsNotFoundForPendingEmoji(t *testing.T) {
	e := &entity{
		emoji:  &fakeEmojiRepo{lookup: model.EmojiLookup{Id: 1, GuildId: 77, Name: "party-cat", Done: false}},
		guild:  &fakeGuildRepo{guilds: map[int64]model.Guild{}},
		member: &fakeMemberRepo{members: map[[2]int64]bool{}},
	}
	app := newEmojiTestApp(t, 15, e)

	req := httptest.NewRequest("GET", "/api/v1/info/emoji/1", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
