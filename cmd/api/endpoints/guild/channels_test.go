package guild

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

type fakeChannelListGuildChannelsRepo struct {
	getGuildChannelsCalls int
}

func (f *fakeChannelListGuildChannelsRepo) AddChannel(ctx context.Context, guildID, channelID int64, channelName string, channelType model.ChannelType, parentID *int64, private bool, position int, topic *string, creatorID *int64, closed bool) error {
	return nil
}

func (f *fakeChannelListGuildChannelsRepo) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	return model.GuildChannel{}, nil
}

func (f *fakeChannelListGuildChannelsRepo) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	f.getGuildChannelsCalls++
	return nil, errors.New("unexpected guild channel lookup")
}

func (f *fakeChannelListGuildChannelsRepo) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	return model.GuildChannel{}, nil
}

func (f *fakeChannelListGuildChannelsRepo) GetGuildChannelsByChannelIDs(ctx context.Context, channelIDs []int64) ([]model.GuildChannel, error) {
	return nil, nil
}

func (f *fakeChannelListGuildChannelsRepo) RemoveChannel(ctx context.Context, guildID, channelID int64) error {
	return nil
}

func (f *fakeChannelListGuildChannelsRepo) SetGuildChannelPosition(ctx context.Context, updates []model.GuildChannelUpdatePosition) error {
	return nil
}

func (f *fakeChannelListGuildChannelsRepo) ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error {
	return nil
}

func (f *fakeChannelListGuildChannelsRepo) GetGuildsChannelsIDsMany(ctx context.Context, guilds []int64) ([]int64, error) {
	return nil, nil
}

func TestGetChannelsUsesCache(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	guildID := int64(1)
	cachedChannels := []dto.Channel{
		{Id: 15, Type: model.ChannelTypeGuild, Name: "cached-general", Position: 2},
	}
	if err := cache.SetJSON(context.Background(), "guild:1:channels", cachedChannels); err != nil {
		t.Fatalf("unable to seed cache: %v", err)
	}

	guildChannelsRepo := &fakeChannelListGuildChannelsRepo{}
	members := &fakeMemberRepo{members: map[testMemberKey]bool{{guildID: guildID, userID: 10}: true}}
	e := &entity{
		cache: cache,
		gc:    guildChannelsRepo,
		memb:  members,
		g:     &fakeGuildRepo{guild: model.Guild{Id: guildID, Name: "cached-guild"}},
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

	var got []dto.Channel
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if len(got) != 1 || got[0].Id != 15 || got[0].Name != "cached-general" || got[0].Position != 2 {
		t.Fatalf("unexpected cached response: %#v", got)
	}
	if guildChannelsRepo.getGuildChannelsCalls != 0 {
		t.Fatalf("expected cache hit without repo call, got %d repo calls", guildChannelsRepo.getGuildChannelsCalls)
	}
}
