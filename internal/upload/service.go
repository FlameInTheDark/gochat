package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocql/gocql"

	attachmentrepo "github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	avatarrepo "github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	iconrepo "github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/observability"
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
	log            *slog.Logger
}

func NewAttachmentService(repo attachmentrepo.Attachment, storage Storage, publicBase string, processor MediaProcessor, logger *slog.Logger) *AttachmentService {
	return &AttachmentService{
		repo:           repo,
		storage:        storage,
		processor:      processor,
		publicBase:     publicBase,
		previewMaxSize: DefaultAttachmentPreviewSize,
		log:            logger,
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
	isWebP, isAnimatedWebP, widthPtr, heightPtr := sniffWEBPMetadata(prepared.Sniff)
	var capturedWebP bytes.Buffer

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

	uploadBody := prepared.Reader
	if kind == "image" && isWebP {
		uploadBody = io.TeeReader(prepared.Reader, &capturedWebP)
	}

	if err := uploadReader(ctx, s.storage, originalKey, uploadBody, prepared.ContentType); err != nil {
		return nil, err
	}
	uploadedKeys = append(uploadedKeys, originalKey)
	if kind == "image" && isWebP {
		fullWebP := capturedWebP.Bytes()
		if !isAnimatedWebP {
			isAnimatedWebP = isAnimatedWEBP(fullWebP)
		}
		if (widthPtr == nil || heightPtr == nil) && len(fullWebP) > 0 {
			if width, height, dimErr := DecodeImageDimensions(fullWebP); dimErr == nil {
				widthPtr = &width
				heightPtr = &height
			}
		}
	}

	var previewURL *string

	if kind == "image" || kind == "video" {
		if kind == "image" && isAnimatedWebP {
			previewBytes, renderedWidth, renderedHeight, previewErr := s.createAnimatedWEBPPreview(ctx, capturedWebP.Bytes())
			switch {
			case previewErr == nil && len(previewBytes) > 0:
				if widthPtr == nil || heightPtr == nil {
					widthPtr = &renderedWidth
					heightPtr = &renderedHeight
				}
				previewKey := AttachmentPreviewKey(channelID, attachmentID)
				if err := uploadBytes(ctx, s.storage, previewKey, previewBytes, "image/webp"); err != nil {
					return nil, err
				}
				uploadedKeys = append(uploadedKeys, previewKey)
				preview := PublicURL(s.publicBase, previewKey)
				previewURL = &preview
			default:
				preview := finalURL
				previewURL = &preview
				reason := "animated_preview_render_failed"
				extra := make([]any, 0, 2)
				if previewErr != nil {
					extra = append(extra, "error", previewErr.Error())
				}
				s.logAttachmentPreviewFallback(ctx, slog.LevelWarn, "animated webp preview fallback to original asset",
					channelID, attachmentID, placeholder.Name, prepared.ContentType, actualDimensions(widthPtr, heightPtr), reason, extra...)
			}
		} else {
			source, urlErr := s.storage.MakeDownloadURL(ctx, originalKey, defaultDownloadURLTTL)
			if urlErr != nil {
				return nil, fmt.Errorf("%w: %v", ErrStorage, urlErr)
			}

			if widthPtr == nil || heightPtr == nil {
				if width, height, probeErr := s.processor.ProbeDimensions(ctx, source); probeErr == nil && width > 0 && height > 0 {
					widthPtr = &width
					heightPtr = &height
				}
			}

			previewBytes, previewErr := s.processor.CreateWebPPreview(ctx, source, s.previewMaxSize)
			switch {
			case previewErr == nil && len(previewBytes) > 0:
				previewKey := AttachmentPreviewKey(channelID, attachmentID)
				if err := uploadBytes(ctx, s.storage, previewKey, previewBytes, "image/webp"); err != nil {
					return nil, err
				}
				uploadedKeys = append(uploadedKeys, previewKey)
				preview := PublicURL(s.publicBase, previewKey)
				previewURL = &preview
			case kind == "image" && isWebP:
				preview := finalURL
				previewURL = &preview
				reason := "empty_preview_output"
				extra := make([]any, 0, 2)
				if previewErr != nil {
					reason = "preview_generation_failed"
					extra = append(extra, "error", previewErr.Error())
				}
				s.logAttachmentPreviewFallback(ctx, slog.LevelWarn, "webp preview fallback to original asset",
					channelID, attachmentID, placeholder.Name, prepared.ContentType, actualDimensions(widthPtr, heightPtr), reason, extra...)
			case previewErr != nil:
				return nil, previewErr
			default:
				return nil, ErrMediaProcess
			}
		}
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

func (s *AttachmentService) createAnimatedWEBPPreview(ctx context.Context, payload []byte) ([]byte, int64, int64, error) {
	if len(payload) == 0 {
		return nil, 0, 0, fmt.Errorf("%w: empty animated webp payload", ErrMediaProcess)
	}

	rendered, width, height, err := RenderAnimatedWEBPFirstFramePNG(payload)
	if err != nil {
		return nil, 0, 0, err
	}

	preview, err := s.processor.CreateWebPPreviewFromReader(ctx, bytes.NewReader(rendered), s.previewMaxSize)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(preview) == 0 {
		return nil, 0, 0, ErrMediaProcess
	}
	return preview, width, height, nil
}

func sniffWEBPMetadata(sniff []byte) (bool, bool, *int64, *int64) {
	if !isWEBP(sniff) {
		return false, false, nil, nil
	}

	var widthPtr *int64
	var heightPtr *int64
	if width, height, err := DecodeImageDimensions(sniff); err == nil && width > 0 && height > 0 {
		widthPtr = &width
		heightPtr = &height
	}

	return true, isAnimatedWEBP(sniff), widthPtr, heightPtr
}

func actualDimensions(widthPtr, heightPtr *int64) string {
	if widthPtr == nil || heightPtr == nil {
		return ""
	}
	return fmt.Sprintf("%dx%d", *widthPtr, *heightPtr)
}

func (s *AttachmentService) logAttachmentPreviewFallback(ctx context.Context, level slog.Level, msg string, channelID, attachmentID int64, fileName, contentType, dimensions, reason string, extra ...any) {
	if s.log == nil {
		return
	}

	attrs := []any{
		"channel_id", channelID,
		"attachment_id", attachmentID,
		"file_name", fileName,
		"content_type", contentType,
		"preview_strategy", "original_asset",
	}
	if dimensions != "" {
		attrs = append(attrs, "dimensions", dimensions)
	}
	if reason != "" {
		attrs = append(attrs, "reason", reason)
	}
	if len(extra) > 0 {
		attrs = append(attrs, extra...)
	}
	s.log.Log(observability.BackgroundFromContext(ctx), level, msg, attrs...)
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
