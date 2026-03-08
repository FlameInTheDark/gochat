package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocql/gocql"

	attachmentrepo "github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	avatarrepo "github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	iconrepo "github.com/FlameInTheDark/gochat/internal/database/entities/icon"
)

const (
	DefaultAttachmentPreviewSize = 350
	defaultDownloadURLTTL        = time.Minute
)

type Storage interface {
	UploadObject(ctx context.Context, key string, body io.Reader, contentType string) error
	MakeDownloadURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	RemoveAttachment(ctx context.Context, key string) error
}

type AttachmentResult struct {
	AlreadyDone bool
	Kind        string
	ContentType string
	URL         string
	PreviewURL  *string
	Width       *int64
	Height      *int64
	Size        int64
}

type AttachmentService struct {
	repo           attachmentrepo.Attachment
	storage        Storage
	processor      MediaProcessor
	publicBase     string
	previewMaxSize int
}

func NewAttachmentService(repo attachmentrepo.Attachment, storage Storage, publicBase string, processor MediaProcessor) *AttachmentService {
	return &AttachmentService{
		repo:           repo,
		storage:        storage,
		processor:      processor,
		publicBase:     publicBase,
		previewMaxSize: DefaultAttachmentPreviewSize,
	}
}

func (s *AttachmentService) Upload(ctx context.Context, actorID, channelID, attachmentID int64, body io.Reader) (_ *AttachmentResult, err error) {
	placeholder, err := s.repo.GetAttachment(ctx, attachmentID, channelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, ErrPlaceholderNotFound
		}
		return nil, err
	}
	if placeholder.AuthorId == nil || *placeholder.AuthorId != actorID || placeholder.ChannelId != channelID || placeholder.Id != attachmentID {
		return nil, ErrForbidden
	}
	if placeholder.Done {
		return &AttachmentResult{AlreadyDone: true}, nil
	}

	prepared, err := PrepareBody(body, placeholder.FileSize)
	if err != nil {
		return nil, err
	}

	kind := InferAttachmentKind(prepared.ContentType, placeholder.Name)
	if kind == "" {
		kind = "other"
	}
	originalKey := AttachmentOriginalKey(channelID, attachmentID)
	finalURL := PublicURL(s.publicBase, originalKey)

	uploadedKeys := make([]string, 0, 2)
	defer func() {
		if err == nil || len(uploadedKeys) == 0 {
			return
		}
		for _, key := range uploadedKeys {
			_ = s.storage.RemoveAttachment(ctx, key)
		}
		_ = s.repo.RemoveAttachment(ctx, attachmentID, channelID)
	}()

	if err := uploadReader(ctx, s.storage, originalKey, prepared.Reader, prepared.ContentType); err != nil {
		return nil, err
	}
	uploadedKeys = append(uploadedKeys, originalKey)

	var previewURL *string
	var widthPtr *int64
	var heightPtr *int64

	if kind == "image" || kind == "video" {
		source, urlErr := s.storage.MakeDownloadURL(ctx, originalKey, defaultDownloadURLTTL)
		if urlErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrStorage, urlErr)
		}

		if width, height, probeErr := s.processor.ProbeDimensions(ctx, source); probeErr == nil {
			widthPtr = &width
			heightPtr = &height
		}

		previewBytes, previewErr := s.processor.CreateWebPPreview(ctx, source, s.previewMaxSize)
		if previewErr != nil {
			return nil, previewErr
		}
		if len(previewBytes) == 0 {
			return nil, ErrMediaProcess
		}

		previewKey := AttachmentPreviewKey(channelID, attachmentID)
		if err := uploadBytes(ctx, s.storage, previewKey, previewBytes, "image/webp"); err != nil {
			return nil, err
		}
		uploadedKeys = append(uploadedKeys, previewKey)
		preview := PublicURL(s.publicBase, previewKey)
		previewURL = &preview
	}

	actualSize := prepared.Size
	if err := s.repo.DoneAttachment(ctx, attachmentID, channelID, &prepared.ContentType, &finalURL, previewURL, heightPtr, widthPtr, &actualSize, &placeholder.Name, placeholder.AuthorId); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFinalize, err)
	}

	return &AttachmentResult{
		Kind:        kind,
		ContentType: prepared.ContentType,
		URL:         finalURL,
		PreviewURL:  previewURL,
		Width:       widthPtr,
		Height:      heightPtr,
		Size:        actualSize,
	}, nil
}

type AvatarResult struct {
	AlreadyDone bool
	URL         string
	ContentType string
	Width       int64
	Height      int64
	Size        int64
}

type AvatarService struct {
	repo        avatarrepo.Avatar
	storage     Storage
	processor   MediaProcessor
	publicBase  string
	maxDim      int
	maxFileSize int64
}

func NewAvatarService(repo avatarrepo.Avatar, storage Storage, publicBase string, processor MediaProcessor, maxDim int, maxFileSize int64) *AvatarService {
	return &AvatarService{
		repo:        repo,
		storage:     storage,
		processor:   processor,
		publicBase:  publicBase,
		maxDim:      maxDim,
		maxFileSize: maxFileSize,
	}
}

func (s *AvatarService) Upload(ctx context.Context, actorID, userID, avatarID int64, body io.Reader) (_ *AvatarResult, err error) {
	placeholder, err := s.repo.GetAvatar(ctx, avatarID, userID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, ErrPlaceholderNotFound
		}
		return nil, err
	}
	if placeholder.UserId != actorID || userID != actorID {
		return nil, ErrForbidden
	}
	if placeholder.Done {
		return &AvatarResult{AlreadyDone: true}, nil
	}

	buffered, err := ReadBodyToMemory(body, placeholder.FileSize)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(strings.ToLower(buffered.ContentType), "image/") {
		return nil, ErrUnsupportedMedia
	}

	webpBytes, err := s.processor.ConvertToWebP(ctx, bytes.NewReader(buffered.Data), s.maxDim, s.maxFileSize)
	if err != nil {
		return nil, err
	}
	if len(webpBytes) == 0 {
		return nil, ErrMediaProcess
	}
	if int64(len(webpBytes)) > s.maxFileSize {
		return nil, ErrTooLarge
	}

	width, height, err := DecodeImageDimensions(webpBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMediaProcess, err)
	}

	key := AvatarKey(userID, avatarID)
	defer func() {
		if err == nil {
			return
		}
		_ = s.storage.RemoveAttachment(ctx, key)
		_ = s.repo.RemoveAvatar(ctx, avatarID, userID)
	}()

	if err := uploadBytes(ctx, s.storage, key, webpBytes, "image/webp"); err != nil {
		return nil, err
	}

	publicURL := PublicURL(s.publicBase, key)
	contentType := "image/webp"
	size := int64(len(webpBytes))
	if err := s.repo.DoneAvatar(ctx, avatarID, userID, &contentType, &publicURL, &height, &width, &size); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFinalize, err)
	}

	return &AvatarResult{
		URL:         publicURL,
		ContentType: contentType,
		Width:       width,
		Height:      height,
		Size:        size,
	}, nil
}

type IconResult struct {
	AlreadyDone bool
	URL         string
	ContentType string
	Width       int64
	Height      int64
	Size        int64
}

type IconService struct {
	repo        iconrepo.Icon
	storage     Storage
	processor   MediaProcessor
	publicBase  string
	maxDim      int
	maxFileSize int64
}

func NewIconService(repo iconrepo.Icon, storage Storage, publicBase string, processor MediaProcessor, maxDim int, maxFileSize int64) *IconService {
	return &IconService{
		repo:        repo,
		storage:     storage,
		processor:   processor,
		publicBase:  publicBase,
		maxDim:      maxDim,
		maxFileSize: maxFileSize,
	}
}

func (s *IconService) Upload(ctx context.Context, guildID, iconID int64, body io.Reader) (_ *IconResult, err error) {
	placeholder, err := s.repo.GetIcon(ctx, iconID, guildID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, ErrPlaceholderNotFound
		}
		return nil, err
	}
	if placeholder.GuildId != guildID {
		return nil, ErrForbidden
	}
	if placeholder.Done {
		return &IconResult{AlreadyDone: true}, nil
	}

	buffered, err := ReadBodyToMemory(body, placeholder.FileSize)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(strings.ToLower(buffered.ContentType), "image/") {
		return nil, ErrUnsupportedMedia
	}

	webpBytes, err := s.processor.ConvertToWebP(ctx, bytes.NewReader(buffered.Data), s.maxDim, s.maxFileSize)
	if err != nil {
		return nil, err
	}
	if len(webpBytes) == 0 {
		return nil, ErrMediaProcess
	}
	if int64(len(webpBytes)) > s.maxFileSize {
		return nil, ErrTooLarge
	}

	width, height, err := DecodeImageDimensions(webpBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMediaProcess, err)
	}

	key := IconKey(guildID, iconID)
	defer func() {
		if err == nil {
			return
		}
		_ = s.storage.RemoveAttachment(ctx, key)
		_ = s.repo.RemoveIcon(ctx, iconID, guildID)
	}()

	if err := uploadBytes(ctx, s.storage, key, webpBytes, "image/webp"); err != nil {
		return nil, err
	}

	publicURL := PublicURL(s.publicBase, key)
	contentType := "image/webp"
	size := int64(len(webpBytes))
	if err := s.repo.DoneIcon(ctx, iconID, guildID, &contentType, &publicURL, &height, &width, &size); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFinalize, err)
	}

	return &IconResult{
		URL:         publicURL,
		ContentType: contentType,
		Width:       width,
		Height:      height,
		Size:        size,
	}, nil
}

func uploadReader(ctx context.Context, storage Storage, key string, body io.Reader, contentType string) error {
	if err := storage.UploadObject(ctx, key, body, contentType); err != nil {
		if errors.Is(err, ErrEmptyBody) || errors.Is(err, ErrSizeMismatch) || errors.Is(err, ErrTooLarge) {
			return err
		}
		return fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return nil
}

func uploadBytes(ctx context.Context, storage Storage, key string, payload []byte, contentType string) error {
	if err := storage.UploadObject(ctx, key, bytes.NewReader(payload), contentType); err != nil {
		return fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return nil
}

func InferAttachmentKind(contentType, name string) string {
	ct := strings.ToLower(contentType)
	if strings.HasPrefix(ct, "image/") {
		return "image"
	}
	if strings.HasPrefix(ct, "video/") {
		return "video"
	}

	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff":
		return "image"
	case ".mp4", ".m4v", ".mov", ".webm", ".mkv", ".avi", ".wmv", ".flv", ".ogv", ".3gp", ".3g2", ".ts", ".m2ts":
		return "video"
	default:
		return ""
	}
}
