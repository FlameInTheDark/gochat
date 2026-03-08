package message

import (
	"context"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
)

type fakeMessageBanRepo struct {
	bans map[[2]int64]bool
}

func (f *fakeMessageBanRepo) BanUser(ctx context.Context, guildID, userID int64, reason *string) error {
	return nil
}

func (f *fakeMessageBanRepo) UnbanUser(ctx context.Context, guildID, userID int64) error {
	return nil
}

func (f *fakeMessageBanRepo) IsBanned(ctx context.Context, guildID, userID int64) (bool, error) {
	return f.bans[[2]int64{guildID, userID}], nil
}

func (f *fakeMessageBanRepo) GetGuildBans(ctx context.Context, guildID int64) ([]model.GuildBan, error) {
	return nil, nil
}

func TestRedactBannedMessagesClearsContentAttachmentsAndEmbeds(t *testing.T) {
	embedsJSON, err := embed.MarshalEmbeds([]embed.Embed{{Description: "hidden"}})
	if err != nil {
		t.Fatalf("marshal embeds: %v", err)
	}

	raw := []model.Message{
		{Id: 1, UserId: 10, Content: "secret", EmbedsJSON: &embedsJSON},
		{Id: 2, UserId: 11, Content: "visible"},
	}
	messages := []dto.Message{
		{Id: 1, Author: dto.User{Id: 10}, Content: "secret", Attachments: []dto.Attachment{{Filename: "secret.png", URL: "https://example.com/secret.png"}}, Embeds: []embed.Embed{{Description: "hidden"}}, Flags: 0},
		{Id: 2, Author: dto.User{Id: 11}, Content: "visible", Attachments: []dto.Attachment{{Filename: "shown.png", URL: "https://example.com/shown.png"}}, Embeds: []embed.Embed{{Description: "shown"}}, Flags: model.MessageFlagSuppressEmbeds},
	}

	e := &entity{ban: &fakeMessageBanRepo{bans: map[[2]int64]bool{{1, 10}: true}}}
	if err := e.redactBannedMessages(context.Background(), 1, raw, messages); err != nil {
		t.Fatalf("redactBannedMessages returned error: %v", err)
	}

	if messages[0].Content != "" {
		t.Fatalf("expected banned content to be cleared, got %q", messages[0].Content)
	}
	if len(messages[0].Attachments) != 0 {
		t.Fatalf("expected banned attachments to be removed, got %#v", messages[0].Attachments)
	}
	if len(messages[0].Embeds) != 0 {
		t.Fatalf("expected banned embeds to be removed, got %#v", messages[0].Embeds)
	}
	if !model.HasMessageFlag(messages[0].Flags, model.MessageFlagBannedAuthor) {
		t.Fatalf("expected banned flag to be set, flags=%d", messages[0].Flags)
	}
	if messages[1].Content != "visible" {
		t.Fatalf("expected unbanned content to stay visible, got %q", messages[1].Content)
	}
	if len(messages[1].Attachments) != 1 {
		t.Fatalf("expected unbanned attachments to stay visible, got %#v", messages[1].Attachments)
	}
	if !model.HasMessageFlag(messages[1].Flags, model.MessageFlagSuppressEmbeds) {
		t.Fatalf("expected unrelated flags to stay intact, flags=%d", messages[1].Flags)
	}
}
