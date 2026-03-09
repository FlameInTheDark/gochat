package upload

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type uploadCall struct {
	key         string
	contentType string
	data        []byte
}

type fakeStorage struct {
	uploads      []uploadCall
	removed      []string
	downloadKeys []string
	downloadURL  string
	uploadErr    error
	downloadErr  error
}

func (f *fakeStorage) UploadObject(ctx context.Context, key string, body io.Reader, contentType string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	f.uploads = append(f.uploads, uploadCall{key: key, contentType: contentType, data: data})
	return nil
}

func (f *fakeStorage) MakeDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	f.downloadKeys = append(f.downloadKeys, key)
	if f.downloadErr != nil {
		return "", f.downloadErr
	}
	if f.downloadURL != "" {
		return f.downloadURL, nil
	}
	return "signed://" + key, nil
}

func (f *fakeStorage) RemoveAttachment(ctx context.Context, key string) error {
	f.removed = append(f.removed, key)
	return nil
}

type fakeProcessor struct {
	previewBytes  []byte
	convertBytes  []byte
	probeWidth    int64
	probeHeight   int64
	previewErr    error
	convertErr    error
	probeErr      error
	previewSource string
	probeSource   string
	convertData   []byte
}

func (f *fakeProcessor) CreateWebPPreview(ctx context.Context, source string, maxDimension int) ([]byte, error) {
	f.previewSource = source
	if f.previewErr != nil {
		return nil, f.previewErr
	}
	return append([]byte(nil), f.previewBytes...), nil
}

func (f *fakeProcessor) ConvertToWebP(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64) ([]byte, error) {
	f.convertData, _ = io.ReadAll(source)
	if f.convertErr != nil {
		return nil, f.convertErr
	}
	return append([]byte(nil), f.convertBytes...), nil
}

func (f *fakeProcessor) ProbeDimensions(ctx context.Context, source string) (int64, int64, error) {
	f.probeSource = source
	if f.probeErr != nil {
		return 0, 0, f.probeErr
	}
	return f.probeWidth, f.probeHeight, nil
}

type attachmentDoneCall struct {
	contentType string
	url         string
	previewURL  *string
	width       *int64
	height      *int64
	size        int64
	name        string
	authorID    int64
}

type fakeAttachmentRepo struct {
	placeholder   model.Attachment
	getErr        error
	doneErr       error
	selected      []model.Attachment
	selectErr     error
	selectChannel int64
	selectIDs     []int64
	removeCalls   int
	doneCall      *attachmentDoneCall
}

func (f *fakeAttachmentRepo) CreateAttachment(ctx context.Context, id, channelId, authorId, ttlSeconds, fileSize int64, name string) error {
	return nil
}
func (f *fakeAttachmentRepo) RemoveAttachment(ctx context.Context, id, channelId int64) error {
	f.removeCalls++
	return nil
}
func (f *fakeAttachmentRepo) GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error) {
	return f.placeholder, f.getErr
}
func (f *fakeAttachmentRepo) DoneAttachment(ctx context.Context, id, channelId int64, contentType, url, previewURL *string, height, width, fileSize *int64, name *string, authorId *int64) error {
	if contentType == nil || url == nil || fileSize == nil || name == nil || authorId == nil {
		return errors.New("missing finalize arguments")
	}
	f.doneCall = &attachmentDoneCall{
		contentType: *contentType,
		url:         *url,
		previewURL:  previewURL,
		width:       width,
		height:      height,
		size:        *fileSize,
		name:        *name,
		authorID:    *authorId,
	}
	return f.doneErr
}
func (f *fakeAttachmentRepo) SelectAttachmentsByChannel(ctx context.Context, channelId int64, ids []int64) ([]model.Attachment, error) {
	f.selectChannel = channelId
	f.selectIDs = append([]int64(nil), ids...)
	return append([]model.Attachment(nil), f.selected...), f.selectErr
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

type avatarDoneCall struct {
	contentType string
	url         string
	width       int64
	height      int64
	size        int64
}

type fakeAvatarRepo struct {
	placeholder model.Avatar
	getErr      error
	doneErr     error
	removeCalls int
	doneCall    *avatarDoneCall
}

func (f *fakeAvatarRepo) CreateAvatar(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error {
	return nil
}
func (f *fakeAvatarRepo) GetAvatar(ctx context.Context, id, userId int64) (model.Avatar, error) {
	return f.placeholder, f.getErr
}
func (f *fakeAvatarRepo) DoneAvatar(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error {
	if contentType == nil || url == nil || height == nil || width == nil || fileSize == nil {
		return errors.New("missing finalize arguments")
	}
	f.doneCall = &avatarDoneCall{contentType: *contentType, url: *url, width: *width, height: *height, size: *fileSize}
	return f.doneErr
}
func (f *fakeAvatarRepo) RemoveAvatar(ctx context.Context, id, userId int64) error {
	f.removeCalls++
	return nil
}
func (f *fakeAvatarRepo) GetAvatarsByUserId(ctx context.Context, userId int64) ([]model.Avatar, error) {
	return nil, nil
}

type fakeIconRepo struct {
	placeholder model.Icon
	getErr      error
	doneErr     error
	removeCalls int
	doneCall    *avatarDoneCall
}

func (f *fakeIconRepo) CreateIcon(ctx context.Context, id, guildId, ttlSeconds, fileSize int64) error {
	return nil
}
func (f *fakeIconRepo) DoneIcon(ctx context.Context, id, guildId int64, contentType, url *string, height, width, fileSize *int64) error {
	if contentType == nil || url == nil || height == nil || width == nil || fileSize == nil {
		return errors.New("missing finalize arguments")
	}
	f.doneCall = &avatarDoneCall{contentType: *contentType, url: *url, width: *width, height: *height, size: *fileSize}
	return f.doneErr
}
func (f *fakeIconRepo) RemoveIcon(ctx context.Context, id, guildId int64) error {
	f.removeCalls++
	return nil
}
func (f *fakeIconRepo) GetIcon(ctx context.Context, id, guildId int64) (model.Icon, error) {
	return f.placeholder, f.getErr
}
func (f *fakeIconRepo) GetIconsByGuildId(ctx context.Context, guildId int64) ([]model.Icon, error) {
	return nil, nil
}

func TestAttachmentServiceUploadImageUsesDeterministicKeys(t *testing.T) {
	ownerID := int64(7)
	body := pngPayload()
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 66, ChannelId: 55, Name: "preview.webp", FileSize: int64(len(body)), AuthorId: &ownerID}}
	storage := &fakeStorage{downloadURL: "signed://media/55/66/original"}
	processor := &fakeProcessor{previewBytes: makeWebP(300, 200), probeWidth: 640, probeHeight: 480}
	service := NewAttachmentService(repo, storage, "", processor)

	result, err := service.Upload(context.Background(), ownerID, 55, 66, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.Kind != "image" {
		t.Fatalf("expected image kind, got %q", result.Kind)
	}
	if result.URL != "media/55/66/original" {
		t.Fatalf("expected deterministic original key URL, got %q", result.URL)
	}
	if result.PreviewURL == nil || *result.PreviewURL != "media/55/66/preview.webp" {
		t.Fatalf("unexpected preview URL: %#v", result.PreviewURL)
	}
	if len(storage.uploads) != 2 {
		t.Fatalf("expected 2 uploads, got %d", len(storage.uploads))
	}
	if storage.uploads[0].key != "media/55/66/original" {
		t.Fatalf("unexpected original key: %q", storage.uploads[0].key)
	}
	if storage.uploads[1].key != "media/55/66/preview.webp" {
		t.Fatalf("unexpected preview key: %q", storage.uploads[1].key)
	}
	if len(storage.downloadKeys) != 1 || storage.downloadKeys[0] != "media/55/66/original" {
		t.Fatalf("unexpected download URL requests: %#v", storage.downloadKeys)
	}
	if processor.previewSource != "signed://media/55/66/original" || processor.probeSource != "signed://media/55/66/original" {
		t.Fatalf("expected signed S3 download URL to drive preview/probe, got preview=%q probe=%q", processor.previewSource, processor.probeSource)
	}
	if repo.doneCall == nil {
		t.Fatal("expected DoneAttachment to be called")
	}
	if repo.doneCall.name != "preview.webp" {
		t.Fatalf("expected metadata name to stay original, got %q", repo.doneCall.name)
	}
	if repo.doneCall.size != int64(len(body)) {
		t.Fatalf("expected body size to be persisted, got %d", repo.doneCall.size)
	}
}

func TestAttachmentServiceUploadVideoUsesExtensionFallback(t *testing.T) {
	ownerID := int64(11)
	body := []byte("not-a-real-video")
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 4, ChannelId: 9, Name: "clip.mp4", FileSize: int64(len(body)), AuthorId: &ownerID}}
	storage := &fakeStorage{downloadURL: "signed://media/9/4/original"}
	processor := &fakeProcessor{previewBytes: makeWebP(100, 50), probeWidth: 1920, probeHeight: 1080}
	service := NewAttachmentService(repo, storage, "https://files.example", processor)

	result, err := service.Upload(context.Background(), ownerID, 9, 4, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.Kind != "video" {
		t.Fatalf("expected video kind, got %q", result.Kind)
	}
	if result.PreviewURL == nil || *result.PreviewURL != "https://files.example/media/9/4/preview.webp" {
		t.Fatalf("unexpected preview URL: %#v", result.PreviewURL)
	}
	if repo.doneCall == nil || repo.doneCall.width == nil || *repo.doneCall.width != 1920 || repo.doneCall.height == nil || *repo.doneCall.height != 1080 {
		t.Fatalf("expected probed dimensions to be persisted, got %#v", repo.doneCall)
	}
}

func TestAttachmentServiceUploadOtherStoresOnlyOriginal(t *testing.T) {
	ownerID := int64(5)
	body := []byte("plain text body")
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 3, ChannelId: 2, Name: "notes.txt", FileSize: int64(len(body)), AuthorId: &ownerID}}
	storage := &fakeStorage{}
	service := NewAttachmentService(repo, storage, "https://files.example", &fakeProcessor{})

	result, err := service.Upload(context.Background(), ownerID, 2, 3, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.Kind != "other" {
		t.Fatalf("expected other kind, got %q", result.Kind)
	}
	if result.PreviewURL != nil {
		t.Fatalf("expected no preview URL, got %#v", result.PreviewURL)
	}
	if len(storage.uploads) != 1 || storage.uploads[0].key != "media/2/3/original" {
		t.Fatalf("unexpected uploads: %#v", storage.uploads)
	}
	if len(storage.downloadKeys) != 0 {
		t.Fatalf("unexpected signed download URL requests: %#v", storage.downloadKeys)
	}
}

func TestAttachmentServiceUploadFinalizeFailureCleansUp(t *testing.T) {
	ownerID := int64(8)
	body := pngPayload()
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 1, ChannelId: 2, Name: "photo.png", FileSize: int64(len(body)), AuthorId: &ownerID}, doneErr: errors.New("boom")}
	storage := &fakeStorage{}
	processor := &fakeProcessor{previewBytes: makeWebP(50, 50), probeWidth: 50, probeHeight: 50}
	service := NewAttachmentService(repo, storage, "https://files.example", processor)

	_, err := service.Upload(context.Background(), ownerID, 2, 1, bytes.NewReader(body))
	if !errors.Is(err, ErrFinalize) {
		t.Fatalf("expected finalize error, got %v", err)
	}
	if repo.removeCalls != 1 {
		t.Fatalf("expected placeholder cleanup, got %d calls", repo.removeCalls)
	}
	if len(storage.removed) != 2 {
		t.Fatalf("expected both uploaded objects to be removed, got %#v", storage.removed)
	}
}

func TestAvatarServiceUploadSuccess(t *testing.T) {
	body := pngPayload()
	repo := &fakeAvatarRepo{placeholder: model.Avatar{Id: 5, UserId: 9, FileSize: int64(len(body))}}
	storage := &fakeStorage{}
	processor := &fakeProcessor{convertBytes: makeWebP(128, 64)}
	service := NewAvatarService(repo, storage, "", processor, 128, 250*1024)

	result, err := service.Upload(context.Background(), 9, 9, 5, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.URL != "avatars/9/5.webp" {
		t.Fatalf("unexpected avatar URL: %q", result.URL)
	}
	if len(storage.uploads) != 1 || storage.uploads[0].key != "avatars/9/5.webp" {
		t.Fatalf("unexpected uploads: %#v", storage.uploads)
	}
	if repo.doneCall == nil || repo.doneCall.width != 128 || repo.doneCall.height != 64 {
		t.Fatalf("expected webp dimensions to be persisted, got %#v", repo.doneCall)
	}
	if !bytes.Equal(processor.convertData, body) {
		t.Fatalf("expected in-memory conversion input to match original body")
	}
}

func TestAvatarServiceUploadSizeMismatch(t *testing.T) {
	body := pngPayload()
	repo := &fakeAvatarRepo{placeholder: model.Avatar{Id: 5, UserId: 9, FileSize: int64(len(body) + 1)}}
	service := NewAvatarService(repo, &fakeStorage{}, "", &fakeProcessor{}, 128, 250*1024)

	_, err := service.Upload(context.Background(), 9, 9, 5, bytes.NewReader(body))
	if !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("expected size mismatch, got %v", err)
	}
}

func TestIconServiceUploadFinalizeFailureRemovesObjectAndPlaceholder(t *testing.T) {
	body := pngPayload()
	repo := &fakeIconRepo{placeholder: model.Icon{Id: 4, GuildId: 77, FileSize: int64(len(body))}, doneErr: errors.New("boom")}
	storage := &fakeStorage{}
	processor := &fakeProcessor{convertBytes: makeWebP(64, 64)}
	service := NewIconService(repo, storage, "https://files.example", processor, 128, 250*1024)

	_, err := service.Upload(context.Background(), 77, 4, bytes.NewReader(body))
	if !errors.Is(err, ErrFinalize) {
		t.Fatalf("expected finalize error, got %v", err)
	}
	if repo.removeCalls != 1 {
		t.Fatalf("expected placeholder cleanup, got %d calls", repo.removeCalls)
	}
	if len(storage.removed) != 1 || storage.removed[0] != "icons/77/4.webp" {
		t.Fatalf("expected uploaded icon cleanup, got %#v", storage.removed)
	}
}

func pngPayload() []byte {
	return []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'd', 'a', 't', 'a'}
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
