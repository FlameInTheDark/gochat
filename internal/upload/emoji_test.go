package upload

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
)

type fakeEmojiRepo struct {
	placeholder    model.GuildEmoji
	getErr         error
	markReadyErr   error
	deleteCalls    int
	markReadyCalls int
}

func (f *fakeEmojiRepo) PruneExpired(ctx context.Context, guildID int64) error {
	return nil
}

func (f *fakeEmojiRepo) CountActiveGuildEmojis(ctx context.Context, guildID int64) (int64, error) {
	return 0, nil
}

func (f *fakeEmojiRepo) CreatePlaceholder(ctx context.Context, emoji model.GuildEmoji) error {
	return nil
}

func (f *fakeEmojiRepo) ReusePendingPlaceholder(ctx context.Context, emoji model.GuildEmoji) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, emojirepo.ErrEmojiNotFound
}

func (f *fakeEmojiRepo) GetGuildEmoji(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	return f.placeholder, f.getErr
}

func (f *fakeEmojiRepo) GetEmojiLookup(ctx context.Context, emojiID int64) (model.EmojiLookup, error) {
	return model.EmojiLookup{}, nil
}

func (f *fakeEmojiRepo) ListReadyGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

func (f *fakeEmojiRepo) ListReadyGuildEmojisByGuilds(ctx context.Context, guildIDs []int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

func (f *fakeEmojiRepo) MarkReady(ctx context.Context, guildID, emojiID int64, animated bool, actualFileSize int64, width, height int64) (model.GuildEmoji, error) {
	f.markReadyCalls++
	if f.markReadyErr != nil {
		return model.GuildEmoji{}, f.markReadyErr
	}
	return model.GuildEmoji{Name: f.placeholder.Name, Animated: animated}, nil
}

func (f *fakeEmojiRepo) Rename(ctx context.Context, guildID, emojiID int64, name, normalized string) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) Delete(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	f.deleteCalls++
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) DeleteGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

func TestProcessEmojiAnimatedWEBPBypassesFFmpeg(t *testing.T) {
	t.Parallel()

	source := makeAnimatedWEBP(64, 64)
	processor := &FFmpegProcessor{ffmpegPath: "missing-ffmpeg-binary"}

	processed, err := processor.ProcessEmoji(context.Background(), source, 64, 64, 256*1024)
	if err != nil {
		t.Fatalf("ProcessEmoji returned error: %v", err)
	}
	if !processed.Animated {
		t.Fatal("expected animated result")
	}
	if processed.Width != 64 || processed.Height != 64 {
		t.Fatalf("unexpected dimensions: %dx%d", processed.Width, processed.Height)
	}
	if !bytes.Equal(processed.Master, source) {
		t.Fatal("expected master payload to preserve animated webp source")
	}
	if !bytes.Equal(processed.Size96, source) {
		t.Fatal("expected 96px variant to reuse animated webp source")
	}
	if !bytes.Equal(processed.Size44, source) {
		t.Fatal("expected 44px variant to reuse animated webp source")
	}
}

func TestEmojiUploadRetainsPlaceholderOnFinalizeFailure(t *testing.T) {
	t.Parallel()

	source := makeAnimatedWEBP(96, 96)
	repo := &fakeEmojiRepo{placeholder: model.GuildEmoji{
		GuildId:          7,
		Id:               11,
		Name:             "spin",
		DeclaredFileSize: int64(len(source)),
		UploadExpiresAt:  time.Now().UTC().Add(time.Minute),
	}, markReadyErr: errors.New("db down")}
	storage := &fakeStorage{}
	service := NewEmojiService(repo, storage, "", &FFmpegProcessor{ffmpegPath: "missing-ffmpeg-binary"})

	_, err := service.Upload(context.Background(), 7, 11, bytes.NewReader(source))
	if !errors.Is(err, ErrFinalize) {
		t.Fatalf("expected ErrFinalize, got %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("expected placeholder to remain for retry, deleteCalls=%d", repo.deleteCalls)
	}
	wantRemoved := []string{EmojiMasterKey(11), EmojiSizedKey(11, 96), EmojiSizedKey(11, 44)}
	if !reflect.DeepEqual(storage.removed, wantRemoved) {
		t.Fatalf("unexpected removed keys: got %v want %v", storage.removed, wantRemoved)
	}
	if repo.markReadyCalls != 1 {
		t.Fatalf("expected one finalize attempt, got %d", repo.markReadyCalls)
	}
}

func makeAnimatedWEBP(width, height int) []byte {
	vp8x := make([]byte, 10)
	vp8x[0] = 0x02
	w := width - 1
	h := height - 1
	vp8x[4] = byte(w)
	vp8x[5] = byte(w >> 8)
	vp8x[6] = byte(w >> 16)
	vp8x[7] = byte(h)
	vp8x[8] = byte(h >> 8)
	vp8x[9] = byte(h >> 16)

	anim := make([]byte, 6)
	total := 12 + 8 + len(vp8x) + 8 + len(anim)
	data := make([]byte, total)
	copy(data[0:], []byte("RIFF"))
	binary.LittleEndian.PutUint32(data[4:], uint32(total-8))
	copy(data[8:], []byte("WEBP"))

	offset := 12
	copy(data[offset:], []byte("VP8X"))
	binary.LittleEndian.PutUint32(data[offset+4:], uint32(len(vp8x)))
	copy(data[offset+8:], vp8x)
	offset += 8 + len(vp8x)

	copy(data[offset:], []byte("ANIM"))
	binary.LittleEndian.PutUint32(data[offset+4:], uint32(len(anim)))
	copy(data[offset+8:], anim)

	return data
}
