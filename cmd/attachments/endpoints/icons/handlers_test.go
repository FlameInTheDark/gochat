package icons

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

type fakeIconRepo struct {
	placeholder model.Icon
}

func (f *fakeIconRepo) CreateIcon(ctx context.Context, id, guildId, ttlSeconds, fileSize int64) error {
	return nil
}
func (f *fakeIconRepo) DoneIcon(ctx context.Context, id, guildId int64, contentType, url *string, height, width, fileSize *int64) error {
	return nil
}
func (f *fakeIconRepo) RemoveIcon(ctx context.Context, id, guildId int64) error { return nil }
func (f *fakeIconRepo) GetIcon(ctx context.Context, id, guildId int64) (model.Icon, error) {
	return f.placeholder, nil
}
func (f *fakeIconRepo) GetIconsByGuildId(ctx context.Context, guildId int64) ([]model.Icon, error) {
	return nil, nil
}

type fakeStorage struct{}

func (f *fakeStorage) UploadObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, _ = io.ReadAll(body)
	return nil
}
func (f *fakeStorage) MakeDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return "", nil
}
func (f *fakeStorage) RemoveAttachment(ctx context.Context, key string) error { return nil }

type fakeProcessor struct{}

func (f *fakeProcessor) CreateWebPPreview(ctx context.Context, source string, maxDimension int) ([]byte, error) {
	return nil, nil
}
func (f *fakeProcessor) ConvertToWebP(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64) ([]byte, error) {
	_, _ = io.ReadAll(source)
	return makeWebP(64, 64), nil
}
func (f *fakeProcessor) ProbeDimensions(ctx context.Context, source string) (int64, int64, error) {
	return 0, 0, nil
}

type fakeGuildRepo struct {
	setCh chan struct{}
}

func (f *fakeGuildRepo) GetGuildById(ctx context.Context, id int64) (model.Guild, error) {
	return model.Guild{Id: id, Name: "guild", OwnerId: 15}, nil
}
func (f *fakeGuildRepo) CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error {
	return nil
}
func (f *fakeGuildRepo) DeleteGuild(ctx context.Context, id int64) error { return nil }
func (f *fakeGuildRepo) SetGuildIcon(ctx context.Context, id, icon int64) error {
	select {
	case f.setCh <- struct{}{}:
	default:
	}
	return nil
}
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

type fakeTransport struct {
	guildUpdateCh chan struct{}
}

func (f *fakeTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}
func (f *fakeTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	select {
	case f.guildUpdateCh <- struct{}{}:
	default:
	}
	return nil
}
func (f *fakeTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func TestUploadHandlerTriggersGuildIconSideEffects(t *testing.T) {
	ownerID := int64(15)
	body := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'd', 'a', 't', 'a'}
	repo := &fakeIconRepo{placeholder: model.Icon{Id: 4, GuildId: 77, FileSize: int64(len(body))}}
	guildRepo := &fakeGuildRepo{setCh: make(chan struct{}, 1)}
	transport := &fakeTransport{guildUpdateCh: make(chan struct{}, 1)}
	e := &entity{
		gld:      guildRepo,
		mqt:      transport,
		uploader: upload.NewIconService(repo, &fakeStorage{}, "", &fakeProcessor{}, iconMaxDim, iconMaxSizeBytes),
	}

	app := fiber.New()
	app.Post("/:guild_id/:icon_id", func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: ownerID}})
		return e.Upload(c)
	})

	req := httptest.NewRequest("POST", "/77/4", bytes.NewReader(body))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	select {
	case <-guildRepo.setCh:
	case <-time.After(time.Second):
		t.Fatal("expected guild icon activation side effect")
	}
	select {
	case <-transport.guildUpdateCh:
	case <-time.After(time.Second):
		t.Fatal("expected guild update event")
	}
}

func makeWebP(width, height int) []byte {
	data := make([]byte, 30)
	copy(data[0:], []byte("RIFF"))
	binary.LittleEndian.PutUint32(data[4:], uint32(22))
	copy(data[8:], []byte("WEBP"))
	copy(data[12:], []byte("VP8X"))
	binary.LittleEndian.PutUint32(data[16:], uint32(10))
	w := width - 1
	h := height - 1
	data[24] = byte(w)
	data[25] = byte(w >> 8)
	data[26] = byte(w >> 16)
	data[27] = byte(h)
	data[28] = byte(h >> 8)
	data[29] = byte(h >> 16)
	return data
}
