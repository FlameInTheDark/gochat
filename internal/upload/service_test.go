package upload

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"image/png"
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
	previewBytes   []byte
	previewReader  []byte
	convertBytes   []byte
	probeWidth     int64
	probeHeight    int64
	previewErr     error
	previewReadErr error
	convertErr     error
	probeErr       error
	previewSource  string
	probeSource    string
	convertData    []byte
	previewData    []byte
	cropArea       *CropArea
	cropAnimated   bool
}

func (f *fakeProcessor) CreateWebPPreview(ctx context.Context, source string, maxDimension int) ([]byte, error) {
	f.previewSource = source
	if f.previewErr != nil {
		return nil, f.previewErr
	}
	return append([]byte(nil), f.previewBytes...), nil
}

func (f *fakeProcessor) CreateWebPPreviewFromReader(ctx context.Context, source io.Reader, maxDimension int) ([]byte, error) {
	f.previewData, _ = io.ReadAll(source)
	if f.previewReadErr != nil {
		return nil, f.previewReadErr
	}
	return append([]byte(nil), f.previewReader...), nil
}

func (f *fakeProcessor) ConvertToWebP(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64) ([]byte, error) {
	f.convertData, _ = io.ReadAll(source)
	if f.convertErr != nil {
		return nil, f.convertErr
	}
	return append([]byte(nil), f.convertBytes...), nil
}

func (f *fakeProcessor) ConvertToWebPWithCrop(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64, crop CropArea, animated bool) ([]byte, error) {
	f.convertData, _ = io.ReadAll(source)
	f.cropArea = &crop
	f.cropAnimated = animated
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

type fakeBannerRepo struct {
	placeholder model.Banner
	getErr      error
	doneErr     error
	removeCalls int
	doneCall    *avatarDoneCall
}

func (f *fakeBannerRepo) CreateBanner(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error {
	return nil
}
func (f *fakeBannerRepo) GetBanner(ctx context.Context, id, userId int64) (model.Banner, error) {
	return f.placeholder, f.getErr
}
func (f *fakeBannerRepo) DoneBanner(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error {
	if contentType == nil || url == nil || height == nil || width == nil || fileSize == nil {
		return errors.New("missing finalize arguments")
	}
	f.doneCall = &avatarDoneCall{contentType: *contentType, url: *url, width: *width, height: *height, size: *fileSize}
	return f.doneErr
}
func (f *fakeBannerRepo) RemoveBanner(ctx context.Context, id, userId int64) error {
	f.removeCalls++
	return nil
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
	service := NewAttachmentService(repo, storage, "", processor, nil)

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
	service := NewAttachmentService(repo, storage, "https://files.example", processor, nil)

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
	service := NewAttachmentService(repo, storage, "https://files.example", &fakeProcessor{}, nil)

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

func TestRenderAnimatedWEBPFirstFramePNG(t *testing.T) {
	body := animatedWEBPFixture()

	rendered, width, height, err := RenderAnimatedWEBPFirstFramePNG(body)
	if err != nil {
		t.Fatalf("RenderAnimatedWEBPFirstFramePNG returned error: %v", err)
	}
	if width != 4 || height != 2 {
		t.Fatalf("unexpected rendered dimensions: %dx%d", width, height)
	}

	img, err := png.Decode(bytes.NewReader(rendered))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 2 {
		t.Fatalf("unexpected rendered png bounds: %v", img.Bounds())
	}
	r, g, b, a := img.At(0, 0).RGBA()
	if r>>8 < 200 || g>>8 > 30 || b>>8 > 30 || a>>8 < 200 {
		t.Fatalf("unexpected rendered pixel rgba=%d,%d,%d,%d", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestAttachmentServiceUploadAnimatedWEBPUsesExtractedPreview(t *testing.T) {
	ownerID := int64(17)
	body := animatedWEBPFixture()
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 9, ChannelId: 4, Name: "dance.webp", FileSize: int64(len(body)), AuthorId: &ownerID}}
	storage := &fakeStorage{}
	processor := &fakeProcessor{previewReader: makeWebP(64, 48)}
	service := NewAttachmentService(repo, storage, "https://files.example", processor, nil)

	result, err := service.Upload(context.Background(), ownerID, 4, 9, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.Kind != "image" {
		t.Fatalf("expected image kind, got %q", result.Kind)
	}
	if result.PreviewURL == nil || *result.PreviewURL != "https://files.example/media/4/9/preview.webp" {
		t.Fatalf("expected generated preview URL, got %#v", result.PreviewURL)
	}
	if result.Width == nil || *result.Width != 4 || result.Height == nil || *result.Height != 2 {
		t.Fatalf("expected rendered dimensions, got width=%v height=%v", result.Width, result.Height)
	}
	if len(storage.uploads) != 2 || storage.uploads[0].key != "media/4/9/original" || storage.uploads[1].key != "media/4/9/preview.webp" {
		t.Fatalf("unexpected uploads: %#v", storage.uploads)
	}
	if len(storage.downloadKeys) != 0 {
		t.Fatalf("expected animated webp preview path to skip signed download, got %#v", storage.downloadKeys)
	}
	if len(processor.previewData) == 0 {
		t.Fatal("expected rendered frame to be passed to the preview processor")
	}
	if repo.doneCall == nil || repo.doneCall.previewURL == nil || *repo.doneCall.previewURL != "https://files.example/media/4/9/preview.webp" {
		t.Fatalf("expected finalize preview fallback, got %#v", repo.doneCall)
	}
}

func TestAttachmentServiceUploadWEBPPreviewFailureFallsBackToOriginal(t *testing.T) {
	ownerID := int64(19)
	body := makeWebP(48, 24)
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 8, ChannelId: 6, Name: "still.webp", FileSize: int64(len(body)), AuthorId: &ownerID}}
	storage := &fakeStorage{downloadURL: "signed://media/6/8/original"}
	processor := &fakeProcessor{previewErr: errors.New("ffmpeg failed"), probeErr: errors.New("ffprobe failed")}
	service := NewAttachmentService(repo, storage, "https://files.example", processor, nil)

	result, err := service.Upload(context.Background(), ownerID, 6, 8, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.PreviewURL == nil || *result.PreviewURL != "https://files.example/media/6/8/original" {
		t.Fatalf("expected original URL preview fallback, got %#v", result.PreviewURL)
	}
	if result.Width == nil || *result.Width != 48 || result.Height == nil || *result.Height != 24 {
		t.Fatalf("expected sniffed dimensions, got width=%v height=%v", result.Width, result.Height)
	}
	if len(storage.uploads) != 1 || storage.uploads[0].key != "media/6/8/original" {
		t.Fatalf("unexpected uploads: %#v", storage.uploads)
	}
	if len(storage.downloadKeys) != 1 || storage.downloadKeys[0] != "media/6/8/original" {
		t.Fatalf("expected preview attempt to use signed download URL, got %#v", storage.downloadKeys)
	}
	if repo.doneCall == nil || repo.doneCall.previewURL == nil || *repo.doneCall.previewURL != "https://files.example/media/6/8/original" {
		t.Fatalf("expected finalize preview fallback, got %#v", repo.doneCall)
	}
}

func TestAttachmentServiceUploadFinalizeFailureCleansUp(t *testing.T) {
	ownerID := int64(8)
	body := pngPayload()
	repo := &fakeAttachmentRepo{placeholder: model.Attachment{Id: 1, ChannelId: 2, Name: "photo.png", FileSize: int64(len(body)), AuthorId: &ownerID}, doneErr: errors.New("boom")}
	storage := &fakeStorage{}
	processor := &fakeProcessor{previewBytes: makeWebP(50, 50), probeWidth: 50, probeHeight: 50}
	service := NewAttachmentService(repo, storage, "https://files.example", processor, nil)

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

func TestBannerServiceUploadSuccess(t *testing.T) {
	body := makeWebP(1200, 420)
	repo := &fakeBannerRepo{placeholder: model.Banner{Id: 7, UserId: 9, FileSize: int64(len(body))}}
	storage := &fakeStorage{}
	processor := &fakeProcessor{convertBytes: makeWebP(1200, 420)}
	service := NewBannerService(repo, storage, "", processor, 1920, 10*1024*1024, 680, 240)

	result, err := service.Upload(context.Background(), 9, 9, 7, bytes.NewReader(body), nil)
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.URL != "banners/9/7.webp" {
		t.Fatalf("unexpected banner URL: %q", result.URL)
	}
	if len(storage.uploads) != 1 || storage.uploads[0].key != "banners/9/7.webp" {
		t.Fatalf("unexpected uploads: %#v", storage.uploads)
	}
	if repo.doneCall == nil || repo.doneCall.width != 1200 || repo.doneCall.height != 420 {
		t.Fatalf("expected webp dimensions to be persisted, got %#v", repo.doneCall)
	}
}

func TestBannerServiceRejectsSmallDimensions(t *testing.T) {
	body := makeWebP(640, 240)
	repo := &fakeBannerRepo{placeholder: model.Banner{Id: 7, UserId: 9, FileSize: int64(len(body))}}
	service := NewBannerService(repo, &fakeStorage{}, "", &fakeProcessor{}, 1920, 10*1024*1024, 680, 240)

	_, err := service.Upload(context.Background(), 9, 9, 7, bytes.NewReader(body), nil)
	if !errors.Is(err, ErrInvalidDimensions) {
		t.Fatalf("expected invalid dimensions, got %v", err)
	}
}

func TestBannerServiceUploadUsesCropArea(t *testing.T) {
	body := makeWebP(1200, 420)
	repo := &fakeBannerRepo{placeholder: model.Banner{Id: 7, UserId: 9, FileSize: int64(len(body))}}
	storage := &fakeStorage{}
	crop := &CropArea{X: 120, Y: 60, Width: 680, Height: 240}
	processor := &fakeProcessor{convertBytes: makeWebP(680, 240)}
	service := NewBannerService(repo, storage, "", processor, 1920, 10*1024*1024, 680, 240)

	result, err := service.Upload(context.Background(), 9, 9, 7, bytes.NewReader(body), crop)
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if result.Width != 680 || result.Height != 240 {
		t.Fatalf("expected cropped dimensions, got %dx%d", result.Width, result.Height)
	}
	if processor.cropArea == nil || *processor.cropArea != *crop {
		t.Fatalf("expected crop to be sent to processor, got %#v", processor.cropArea)
	}
	if processor.cropAnimated {
		t.Fatal("expected still image crop conversion")
	}
	if len(storage.uploads) != 1 || !bytes.Equal(storage.uploads[0].data, processor.convertBytes) {
		t.Fatalf("unexpected upload payload: %#v", storage.uploads)
	}
}

func TestBannerServiceRejectsInvalidCropArea(t *testing.T) {
	body := makeWebP(1200, 420)
	repo := &fakeBannerRepo{placeholder: model.Banner{Id: 7, UserId: 9, FileSize: int64(len(body))}}
	processor := &fakeProcessor{convertBytes: makeWebP(680, 240)}
	service := NewBannerService(repo, &fakeStorage{}, "", processor, 1920, 10*1024*1024, 680, 240)

	_, err := service.Upload(context.Background(), 9, 9, 7, bytes.NewReader(body), &CropArea{X: 900, Y: 0, Width: 680, Height: 240})
	if !errors.Is(err, ErrInvalidDimensions) {
		t.Fatalf("expected invalid dimensions, got %v", err)
	}
	if processor.cropArea != nil {
		t.Fatalf("processor should not be called for invalid crop, got %#v", processor.cropArea)
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

func animatedWEBPFixture() []byte {
	still, err := base64.StdEncoding.DecodeString("UklGRjwAAABXRUJQVlA4IDAAAADQAQCdASoEAAIAAgA0JaACdLoB+AADsAD+8Oj3/yC5YXXI1/8gP+QH/ID/+PIAAAA=")
	if err != nil {
		panic(err)
	}
	payload := append([]byte(nil), still[12:]...)
	anmfHeader := make([]byte, 16)
	anmfHeader[6] = 3
	anmfHeader[9] = 1
	anmfPayload := append(anmfHeader, payload...)

	body := make([]byte, 0, 128)
	body = append(body, makeChunk("VP8X", []byte{0x02, 0, 0, 0, 3, 0, 0, 1, 0, 0})...)
	body = append(body, makeChunk("ANIM", []byte{0, 0, 0, 0, 0, 0})...)
	body = append(body, makeChunk("ANMF", anmfPayload)...)

	data := make([]byte, 12, 12+len(body))
	copy(data[:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(data[4:8], uint32(4+len(body)))
	copy(data[8:12], []byte("WEBP"))
	return append(data, body...)
}

func makeChunk(tag string, payload []byte) []byte {
	out := make([]byte, 8, 8+len(payload)+1)
	copy(out[:4], []byte(tag))
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(payload)))
	out = append(out, payload...)
	if len(payload)%2 == 1 {
		out = append(out, 0)
	}
	return out
}
