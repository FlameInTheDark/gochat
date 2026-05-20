package user

import (
	"strings"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

func TestModifyUserRequestValidateAllowsBioAndProfileColors(t *testing.T) {
	bio := "Building gochat one endpoint at a time"
	bannerColor := 0
	panelColor := maxProfileColorValue

	req := ModifyUserRequest{
		Bio:         &bio,
		BannerColor: &bannerColor,
		PanelColor:  &panelColor,
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("expected request to be valid, got %v", err)
	}
}

func TestModifyUserRequestValidateRejectsTooLongBio(t *testing.T) {
	bio := strings.Repeat("a", 191)

	req := ModifyUserRequest{Bio: &bio}

	err := req.Validate()
	if err == nil {
		t.Fatal("expected bio length validation error")
	}
	if !strings.Contains(err.Error(), ErrUserBioTooLong) {
		t.Fatalf("expected %q, got %q", ErrUserBioTooLong, err.Error())
	}
}

func TestModifyUserRequestValidateRejectsInvalidProfileColor(t *testing.T) {
	bannerColor := maxProfileColorValue + 1

	req := ModifyUserRequest{BannerColor: &bannerColor}

	err := req.Validate()
	if err == nil {
		t.Fatal("expected profile color validation error")
	}
	if !strings.Contains(err.Error(), ErrProfileColorInvalid) {
		t.Fatalf("expected %q, got %q", ErrProfileColorInvalid, err.Error())
	}
}

func TestModelToUserIncludesProfileFields(t *testing.T) {
	bio := "Public profile bio"
	bannerColor := 3447003
	panelColor := 15158332

	got := modelToUser(model.User{
		Id:          42,
		Name:        "FancyUserName",
		Bio:         &bio,
		BannerColor: &bannerColor,
		PanelColor:  &panelColor,
	})

	if got.Bio == nil || *got.Bio != bio {
		t.Fatalf("expected bio to be copied, got %#v", got.Bio)
	}
	if got.BannerColor == nil || *got.BannerColor != bannerColor {
		t.Fatalf("expected banner color to be copied, got %#v", got.BannerColor)
	}
	if got.PanelColor == nil || *got.PanelColor != panelColor {
		t.Fatalf("expected panel color to be copied, got %#v", got.PanelColor)
	}
}

func TestFilterGuildLastMessagesExcludesThreadsAndDeletedChannels(t *testing.T) {
	glms := map[int64]map[int64]int64{
		1: {
			10: 100,
			11: 101,
			12: 102,
			13: 0,
		},
	}

	channels := []model.Channel{
		{Id: 10, Type: model.ChannelTypeGuild},
		{Id: 11, Type: model.ChannelTypeThread},
		{Id: 13, Type: model.ChannelTypeGuild},
	}

	filtered := filterGuildLastMessages(glms, channels)
	guildMessages := filtered[1]
	if len(guildMessages) != 1 {
		t.Fatalf("expected exactly one surviving channel, got %#v", guildMessages)
	}
	if guildMessages[10] != 100 {
		t.Fatalf("expected regular channel entry to survive, got %#v", guildMessages)
	}
	if _, ok := guildMessages[11]; ok {
		t.Fatalf("expected thread entry to be removed, got %#v", guildMessages)
	}
	if _, ok := guildMessages[12]; ok {
		t.Fatalf("expected deleted channel entry to be removed, got %#v", guildMessages)
	}
	if _, ok := guildMessages[13]; ok {
		t.Fatalf("expected empty channel entry to be removed, got %#v", guildMessages)
	}
}

func TestFilterThreadLastMessagesKeepsOnlyJoinedLiveThreads(t *testing.T) {
	glms := map[int64]map[int64]int64{
		1: {
			11: 101,
			12: 102,
			13: 0,
		},
		2: {
			21: 201,
		},
	}

	joined := map[int64]struct{}{
		11: {},
		13: {},
		21: {},
		99: {},
	}

	channels := []model.Channel{
		{Id: 11, Type: model.ChannelTypeThread},
		{Id: 12, Type: model.ChannelTypeThread},
		{Id: 13, Type: model.ChannelTypeThread},
		{Id: 21, Type: model.ChannelTypeThread},
		{Id: 22, Type: model.ChannelTypeGuild},
	}

	filtered := filterThreadLastMessages(joined, channels, glms)
	if len(filtered) != 2 {
		t.Fatalf("expected exactly two surviving threads, got %#v", filtered)
	}
	if filtered[11] != 101 || filtered[21] != 201 {
		t.Fatalf("unexpected filtered thread messages: %#v", filtered)
	}
	if _, ok := filtered[12]; ok {
		t.Fatalf("expected unjoined thread to be removed, got %#v", filtered)
	}
	if _, ok := filtered[99]; ok {
		t.Fatalf("expected deleted thread to be removed, got %#v", filtered)
	}
	if _, ok := filtered[13]; ok {
		t.Fatalf("expected empty thread to be removed, got %#v", filtered)
	}
}

func TestBuildJoinedThreadsGroupsByGuildAndParentAndSortsIDs(t *testing.T) {
	joined := map[int64]struct{}{
		31: {},
		29: {},
		44: {},
		99: {},
	}

	channels := []model.Channel{
		{Id: 31, Type: model.ChannelTypeThread, ParentID: int64Ptr(10)},
		{Id: 29, Type: model.ChannelTypeThread, ParentID: int64Ptr(10)},
		{Id: 44, Type: model.ChannelTypeThread, ParentID: int64Ptr(20)},
		{Id: 55, Type: model.ChannelTypeThread, ParentID: int64Ptr(20)},
		{Id: 66, Type: model.ChannelTypeGuild},
	}

	guildChannels := []model.GuildChannel{
		{GuildId: 1, ChannelId: 31},
		{GuildId: 1, ChannelId: 29},
		{GuildId: 2, ChannelId: 44},
		{GuildId: 2, ChannelId: 55},
	}

	got := buildJoinedThreads(joined, channels, guildChannels)

	if len(got) != 2 {
		t.Fatalf("expected two guild buckets, got %#v", got)
	}
	if want := []int64{29, 31}; len(got[1][10]) != len(want) || got[1][10][0] != want[0] || got[1][10][1] != want[1] {
		t.Fatalf("unexpected guild 1 / parent 10 threads: %#v", got[1][10])
	}
	if want := []int64{44}; len(got[2][20]) != len(want) || got[2][20][0] != want[0] {
		t.Fatalf("unexpected guild 2 / parent 20 threads: %#v", got[2][20])
	}
	if _, ok := got[2][55]; ok {
		t.Fatalf("expected thread ids to be grouped under parent ids, got %#v", got[2])
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
