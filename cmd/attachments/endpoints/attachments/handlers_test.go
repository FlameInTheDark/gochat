package attachments

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

type fakeAttachmentRepo struct {
	placeholder model.Attachment
}

func (f *fakeAttachmentRepo) CreateAttachment(ctx context.Context, id, channelId, authorId, ttlSeconds, fileSize int64, name string) error {
	return nil
}
func (f *fakeAttachmentRepo) RemoveAttachment(ctx context.Context, id, channelId int64) error {
	return nil
}
func (f *fakeAttachmentRepo) GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error) {
	return f.placeholder, nil
}
func (f *fakeAttachmentRepo) DoneAttachment(ctx context.Context, id, channelId int64, contentType, url, previewURL *string, height, width, fileSize *int64, name *string, authorId *int64) error {
	return nil
}
func (f *fakeAttachmentRepo) SelectAttachmentsByChannel(ctx context.Context, channelId int64, ids []int64) ([]model.Attachment, error) {
	return nil, nil
}
func (f *fakeAttachmentRepo) UpdateFileSize(ctx context.Context, id, channelId int64, fileSize int64) error {
	return nil
}
func (f *fakeAttachmentRepo) ListDoneZeroSize(ctx context.Context) ([]model.Attachment, error) {
	return nil, nil
}
func (f *fakeAttachmentRepo) UpdateName(ctx context.Context, id, channelId int64, name string) error {
	return nil
}

type fakeStorage struct {
	uploads []string
}

func (f *fakeStorage) UploadObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	f.uploads = append(f.uploads, key)
	_, _ = io.ReadAll(body)
	return nil
}
func (f *fakeStorage) MakeDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return "signed://" + key, nil
}
func (f *fakeStorage) RemoveAttachment(ctx context.Context, key string) error { return nil }

type fakeProcessor struct{}

func (f *fakeProcessor) CreateWebPPreview(ctx context.Context, source string, maxDimension int) ([]byte, error) {
	return []byte("RIFFxxxxWEBPVP8Xabcdefghij"), nil
}
func (f *fakeProcessor) ConvertToWebP(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64) ([]byte, error) {
	return nil, nil
}
func (f *fakeProcessor) ProbeDimensions(ctx context.Context, source string) (int64, int64, error) {
	return 640, 480, nil
}

func TestUploadHandlerReturnsCreated(t *testing.T) {
	ownerID := int64(7)
	body := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'd', 'a', 't', 'a'}
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 66, ChannelId: 55, Name: "preview.webp", FileSize: int64(len(body)), AuthorId: &ownerID}}
	storage := &fakeStorage{}
	e := &entity{uploader: upload.NewAttachmentService(repo, storage, "", &fakeProcessor{})}

	app := fiber.New()
	app.Post("/:channel_id/:attachment_id", func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: ownerID}})
		return e.Upload(c)
	})

	req := httptest.NewRequest("POST", "/55/66", bytes.NewReader(body))
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if len(storage.uploads) != 2 || storage.uploads[0] != "media/55/66/original" || storage.uploads[1] != "media/55/66/preview.webp" {
		t.Fatalf("unexpected uploaded keys: %#v", storage.uploads)
	}
}
