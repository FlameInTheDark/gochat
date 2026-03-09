package avatars

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

type fakeAvatarRepo struct {
	placeholder model.Avatar
}

func (f *fakeAvatarRepo) CreateAvatar(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error {
	return nil
}
func (f *fakeAvatarRepo) GetAvatar(ctx context.Context, id, userId int64) (model.Avatar, error) {
	return f.placeholder, nil
}
func (f *fakeAvatarRepo) DoneAvatar(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error {
	return nil
}
func (f *fakeAvatarRepo) RemoveAvatar(ctx context.Context, id, userId int64) error { return nil }
func (f *fakeAvatarRepo) GetAvatarsByUserId(ctx context.Context, userId int64) ([]model.Avatar, error) {
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
	return makeWebP(128, 64), nil
}
func (f *fakeProcessor) ProbeDimensions(ctx context.Context, source string) (int64, int64, error) {
	return 0, 0, nil
}

type fakeUserRepo struct {
	setCh chan struct{}
}

func (f *fakeUserRepo) ModifyUser(ctx context.Context, userId int64, name *string, avatar *int64) error {
	return nil
}
func (f *fakeUserRepo) GetUserById(ctx context.Context, id int64) (model.User, error) {
	return model.User{Id: id, Name: "alice"}, nil
}
func (f *fakeUserRepo) GetUsersList(ctx context.Context, ids []int64) ([]model.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) CreateUser(ctx context.Context, id int64, name string) error { return nil }
func (f *fakeUserRepo) SetUserAvatar(ctx context.Context, id, attachmentId int64) error {
	select {
	case f.setCh <- struct{}{}:
	default:
	}
	return nil
}
func (f *fakeUserRepo) SetUsername(ctx context.Context, id, name string) error           { return nil }
func (f *fakeUserRepo) SetUserBlocked(ctx context.Context, id int64, blocked bool) error { return nil }
func (f *fakeUserRepo) SetUploadLimit(ctx context.Context, id int64, uploadLimit int64) error {
	return nil
}

type fakeTransport struct {
	userUpdateCh chan struct{}
}

func (f *fakeTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}
func (f *fakeTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	return nil
}
func (f *fakeTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	select {
	case f.userUpdateCh <- struct{}{}:
	default:
	}
	return nil
}

func TestUploadHandlerTriggersAvatarSideEffects(t *testing.T) {
	userID := int64(9)
	body := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'd', 'a', 't', 'a'}
	repo := &fakeAvatarRepo{placeholder: model.Avatar{Id: 5, UserId: userID, FileSize: int64(len(body))}}
	userRepo := &fakeUserRepo{setCh: make(chan struct{}, 1)}
	transport := &fakeTransport{userUpdateCh: make(chan struct{}, 1)}
	e := &entity{
		usr:      userRepo,
		mqt:      transport,
		uploader: upload.NewAvatarService(repo, &fakeStorage{}, "", &fakeProcessor{}, avatarMaxDim, avatarMaxSizeBytes),
	}

	app := fiber.New()
	app.Post("/:user_id/:avatar_id", func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: userID}})
		return e.Upload(c)
	})

	req := httptest.NewRequest("POST", "/9/5", bytes.NewReader(body))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	select {
	case <-userRepo.setCh:
	case <-time.After(time.Second):
		t.Fatal("expected avatar activation side effect")
	}
	select {
	case <-transport.userUpdateCh:
	case <-time.After(time.Second):
		t.Fatal("expected avatar update event")
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
