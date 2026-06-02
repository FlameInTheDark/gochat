package handler

import (
	"context"
	"testing"

	dbmodel "github.com/FlameInTheDark/gochat/internal/database/model"
)

type fakeGatewayReadyIconRepo struct {
	icon dbmodel.Icon
	err  error
}

func (f fakeGatewayReadyIconRepo) CreateIcon(ctx context.Context, id, guildId, ttlSeconds, fileSize int64) error {
	return nil
}

func (f fakeGatewayReadyIconRepo) DoneIcon(ctx context.Context, id, guildId int64, contentType, url *string, height, width, fileSize *int64) error {
	return nil
}

func (f fakeGatewayReadyIconRepo) RemoveIcon(ctx context.Context, id, guildId int64) error {
	return nil
}

func (f fakeGatewayReadyIconRepo) GetIcon(ctx context.Context, id, guildId int64) (dbmodel.Icon, error) {
	return f.icon, f.err
}

func (f fakeGatewayReadyIconRepo) GetIconsByGuildId(ctx context.Context, guildId int64) ([]dbmodel.Icon, error) {
	return nil, nil
}

func TestGatewayReadyGuildModelToDTOIncludesResolvedIcon(t *testing.T) {
	t.Parallel()

	iconID := int64(123)
	guildID := int64(456)
	iconURL := "https://cdn.gochat.local/icons/456/123.png"
	width := int64(64)
	height := int64(64)

	h := &Handler{
		ico: fakeGatewayReadyIconRepo{
			icon: dbmodel.Icon{
				Id:       iconID,
				GuildId:  guildID,
				URL:      &iconURL,
				Width:    &width,
				Height:   &height,
				FileSize: 4096,
				Done:     true,
			},
		},
	}

	got := h.guildModelToDTO(context.Background(), dbmodel.Guild{
		Id:      guildID,
		Name:    "Guild",
		OwnerId: 1,
		Icon:    &iconID,
	})

	if got.Icon == nil {
		t.Fatal("expected gateway ready guild DTO to include icon")
	}
	if got.Icon.URL != iconURL {
		t.Fatalf("icon URL = %q, want %q", got.Icon.URL, iconURL)
	}
	if got.Icon.Width != width || got.Icon.Height != height {
		t.Fatalf("icon dimensions = %dx%d, want %dx%d", got.Icon.Width, got.Icon.Height, width, height)
	}
	if got.Icon.Filesize != 4096 {
		t.Fatalf("icon filesize = %d, want 4096", got.Icon.Filesize)
	}
}
