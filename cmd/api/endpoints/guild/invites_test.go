package guild

import (
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gofiber/fiber/v2"
)

func TestReceiveInviteReturnsNotFoundWhenInviteIsMissing(t *testing.T) {
	e := &entity{
		inv: &fakeInviteRepo{err: sql.ErrNoRows},
	}
	app := newGuildTestApp(t, 10, "/guild/invites/receive/:invite_code", e.ReceiveInvite)

	req := httptest.NewRequest("GET", "/guild/invites/receive/ABCDEFGH", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAcceptInviteReturnsNotFoundWhenInviteIsMissing(t *testing.T) {
	e := &entity{
		inv: &fakeInviteRepo{err: sql.ErrNoRows},
	}
	app := newGuildTestApp(t, 10, "/guild/invites/accept/:invite_code", e.AcceptInvite)

	req := httptest.NewRequest("POST", "/guild/invites/accept/ABCDEFGH", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAcceptInviteSkipsSystemMessageWhenSystemChannelIsMissing(t *testing.T) {
	const (
		guildID         int64 = 1
		userID          int64 = 10
		systemChannelID int64 = 77
	)

	channelRepo := &fakeCreateChannelRepo{
		channels:     map[int64]model.Channel{},
		getChannelCh: make(chan int64, 1),
	}
	transport := &fakeGuildLifecycleTransport{guildEventCh: make(chan struct{}, 1)}
	messageRepo := &fakeDetachMessageRepo{}
	e := &entity{
		inv:  &fakeInviteRepo{invite: model.GuildInvite{GuildId: guildID}},
		memb: &fakeMemberRepo{members: map[testMemberKey]bool{}},
		user: &fakeUserRepo{users: map[int64]model.User{userID: {Id: userID, Name: "member-user"}}},
		disc: &fakeDiscriminatorRepo{discriminators: map[int64]string{userID: "1234"}},
		g: &fakeGuildRepo{guild: model.Guild{
			Id:             guildID,
			Name:           "guild",
			OwnerId:        99,
			SystemMessages: int64Ptr(systemChannelID),
		}},
		ch:  channelRepo,
		msg: messageRepo,
		mqt: transport,
	}
	app := newGuildTestApp(t, userID, "/guild/invites/accept/:invite_code", e.AcceptInvite)

	req := httptest.NewRequest("POST", "/guild/invites/accept/ABCDEFGH", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	select {
	case <-transport.guildEventCh:
	case <-time.After(time.Second):
		t.Fatal("expected guild member add event")
	}

	select {
	case gotChannelID := <-channelRepo.getChannelCh:
		if gotChannelID != systemChannelID {
			t.Fatalf("expected system channel existence check for %d, got %d", systemChannelID, gotChannelID)
		}
	case <-time.After(time.Second):
		t.Fatal("expected system channel existence check")
	}

	if messageRepo.createSystemCalls != 0 {
		t.Fatalf("expected no system messages to be created, got %d", messageRepo.createSystemCalls)
	}
}
