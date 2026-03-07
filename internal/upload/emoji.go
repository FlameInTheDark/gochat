package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
)

type EmojiResult struct {
	AlreadyDone bool
	Name        string
	Animated    bool
}

type EmojiService struct {
	repo       emojirepo.Emoji
	storage    Storage
	processor  *FFmpegProcessor
	publicBase string
}

func NewEmojiService(repo emojirepo.Emoji, storage Storage, publicBase string, processor *FFmpegProcessor) *EmojiService {
	return &EmojiService{
		repo:       repo,
		storage:    storage,
		processor:  processor,
		publicBase: publicBase,
	}
}

func (s *EmojiService) Upload(ctx context.Context, guildID, emojiID int64, body io.Reader) (_ *EmojiResult, err error) {
	placeholder, err := s.repo.GetGuildEmoji(ctx, guildID, emojiID)
	if err != nil {
		if errors.Is(err, emojirepo.ErrEmojiNotFound) {
			return nil, ErrPlaceholderNotFound
		}
		return nil, err
	}
	if placeholder.Done {
		return &EmojiResult{AlreadyDone: true, Name: placeholder.Name, Animated: placeholder.Animated}, nil
	}
	if time.Now().UTC().After(placeholder.UploadExpiresAt) {
		_, _ = s.repo.Delete(ctx, guildID, emojiID)
		return nil, ErrUploadExpired
	}

	buffered, err := ReadBodyToMemory(body, placeholder.DeclaredFileSize)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(strings.ToLower(buffered.ContentType), "image/") {
		return nil, ErrUnsupportedMedia
	}

	sourceWidth, sourceHeight, err := DecodeImageDimensions(buffered.Data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMediaProcess, err)
	}
	if sourceWidth > emojiutil.MaxDimension || sourceHeight > emojiutil.MaxDimension {
		return nil, ErrInvalidDimensions
	}

	processed, err := s.processor.ProcessEmoji(ctx, buffered.Data, sourceWidth, sourceHeight, emojiutil.MaxUploadSizeBytes)
	if err != nil {
		return nil, err
	}

	uploadedKeys := make([]string, 0, 3)
	defer func() {
		if err == nil {
			return
		}
		for _, key := range uploadedKeys {
			_ = s.storage.RemoveAttachment(ctx, key)
		}
		_, _ = s.repo.Delete(ctx, guildID, emojiID)
	}()

	masterKey := EmojiMasterKey(emojiID)
	if err = uploadBytes(ctx, s.storage, masterKey, processed.Master, "image/webp"); err != nil {
		return nil, err
	}
	uploadedKeys = append(uploadedKeys, masterKey)

	key96 := EmojiSizedKey(emojiID, 96)
	if err = uploadBytes(ctx, s.storage, key96, processed.Size96, "image/webp"); err != nil {
		return nil, err
	}
	uploadedKeys = append(uploadedKeys, key96)

	key44 := EmojiSizedKey(emojiID, 44)
	if err = uploadBytes(ctx, s.storage, key44, processed.Size44, "image/webp"); err != nil {
		return nil, err
	}
	uploadedKeys = append(uploadedKeys, key44)

	updated, err := s.repo.MarkReady(ctx, guildID, emojiID, processed.Animated, int64(len(processed.Master)), processed.Width, processed.Height)
	if err != nil {
		switch {
		case errors.Is(err, emojirepo.ErrEmojiQuotaExceeded):
			return nil, ErrQuotaExceeded
		case errors.Is(err, emojirepo.ErrEmojiUploadExpired):
			return nil, ErrUploadExpired
		default:
			return nil, fmt.Errorf("%w: %v", ErrFinalize, err)
		}
	}

	return &EmojiResult{Name: updated.Name, Animated: updated.Animated}, nil
}

func (p *FFmpegProcessor) ProcessEmoji(ctx context.Context, source []byte, sourceWidth, sourceHeight int64, sizeLimit int64) (*ProcessedEmoji, error) {
	animated, err := DetectAnimated(source)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMediaProcess, err)
	}

	master, err := p.convertEmojiVariant(ctx, bytes.NewReader(source), 0, sizeLimit, animated)
	if err != nil {
		return nil, err
	}
	if len(master) == 0 {
		return nil, ErrMediaProcess
	}
	masterWidth, masterHeight, err := DecodeImageDimensions(master)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMediaProcess, err)
	}

	maxSource := sourceWidth
	if sourceHeight > maxSource {
		maxSource = sourceHeight
	}

	size96 := master
	if maxSource > 96 {
		size96, err = p.convertEmojiVariant(ctx, bytes.NewReader(source), 96, sizeLimit, animated)
		if err != nil {
			return nil, err
		}
	}

	size44 := master
	if maxSource > 44 {
		size44, err = p.convertEmojiVariant(ctx, bytes.NewReader(source), 44, sizeLimit, animated)
		if err != nil {
			return nil, err
		}
	}

	return &ProcessedEmoji{
		Master:   master,
		Size96:   size96,
		Size44:   size44,
		Animated: animated,
		Width:    masterWidth,
		Height:   masterHeight,
	}, nil
}

type ProcessedEmoji struct {
	Master   []byte
	Size96   []byte
	Size44   []byte
	Animated bool
	Width    int64
	Height   int64
}
