package guild

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
)

type fakeEmojiRepo struct {
	reuseResults []reusePendingResult
	count        int64
	countErr     error
	createErr    error
	createCalls  int
	countCalls   int
	pruneCalls   int
	created      *model.GuildEmoji
}

type reusePendingResult struct {
	emoji model.GuildEmoji
	err   error
}

func (f *fakeEmojiRepo) PruneExpired(ctx context.Context, guildID int64) error {
	f.pruneCalls++
	return nil
}

func (f *fakeEmojiRepo) CountActiveGuildEmojis(ctx context.Context, guildID int64) (int64, error) {
	f.countCalls++
	return f.count, f.countErr
}

func (f *fakeEmojiRepo) CreatePlaceholder(ctx context.Context, emoji model.GuildEmoji) error {
	f.createCalls++
	copyEmoji := emoji
	f.created = &copyEmoji
	return f.createErr
}

func (f *fakeEmojiRepo) ReusePendingPlaceholder(ctx context.Context, emoji model.GuildEmoji) (model.GuildEmoji, error) {
	if len(f.reuseResults) == 0 {
		return model.GuildEmoji{}, emojirepo.ErrEmojiNotFound
	}
	res := f.reuseResults[0]
	f.reuseResults = f.reuseResults[1:]
	return res.emoji, res.err
}

func (f *fakeEmojiRepo) GetGuildEmoji(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
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
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) Rename(ctx context.Context, guildID, emojiID int64, name, normalized string) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) Delete(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	return model.GuildEmoji{}, nil
}

func (f *fakeEmojiRepo) DeleteGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	return nil, nil
}

func TestReserveEmojiUploadReusesPendingPlaceholder(t *testing.T) {
	t.Parallel()

	repo := &fakeEmojiRepo{
		reuseResults: []reusePendingResult{{emoji: model.GuildEmoji{GuildId: 1, Id: 42, Name: "spin"}}},
	}
	e := &entity{emoji: repo, attachTTL: 60}

	result, err := e.reserveEmojiUpload(context.Background(), 1, 7, CreateEmojiRequest{Name: "spin", FileSize: 123})
	if err != nil {
		t.Fatalf("reserveEmojiUpload returned error: %v", err)
	}
	if result.Id != 42 || result.GuildId != 1 || result.Name != "spin" {
		t.Fatalf("unexpected upload metadata: %#v", result)
	}
	if repo.countCalls != 0 {
		t.Fatalf("expected no active-count query when reusing pending placeholder, got %d", repo.countCalls)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected no create call when reusing pending placeholder, got %d", repo.createCalls)
	}
}

func TestReserveEmojiUploadRejectsForeignPendingPlaceholder(t *testing.T) {
	t.Parallel()

	repo := &fakeEmojiRepo{
		reuseResults: []reusePendingResult{{err: emojirepo.ErrEmojiNameTaken}},
	}
	e := &entity{emoji: repo, attachTTL: 60}

	_, err := e.reserveEmojiUpload(context.Background(), 1, 7, CreateEmojiRequest{Name: "spin", FileSize: 123})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var ferr *fiber.Error
	if !errors.As(err, &ferr) {
		t.Fatalf("expected fiber error, got %T", err)
	}
	if ferr.Code != fiber.StatusConflict || ferr.Message != ErrEmojiNameTaken {
		t.Fatalf("unexpected fiber error: code=%d message=%v", ferr.Code, ferr.Message)
	}
	if repo.countCalls != 0 || repo.createCalls != 0 {
		t.Fatalf("expected no further work after conflict, countCalls=%d createCalls=%d", repo.countCalls, repo.createCalls)
	}
}

func TestReserveEmojiUploadFallsBackToPendingAfterCreateConflict(t *testing.T) {
	t.Parallel()

	repo := &fakeEmojiRepo{
		reuseResults: []reusePendingResult{
			{err: emojirepo.ErrEmojiNotFound},
			{emoji: model.GuildEmoji{GuildId: 9, Id: 77, Name: "retry"}},
		},
		createErr: emojirepo.ErrEmojiNameTaken,
	}
	e := &entity{emoji: repo, attachTTL: 120}

	result, err := e.reserveEmojiUpload(context.Background(), 9, 4, CreateEmojiRequest{Name: "retry", FileSize: 512})
	if err != nil {
		t.Fatalf("reserveEmojiUpload returned error: %v", err)
	}
	if result.Id != 77 || result.GuildId != 9 || result.Name != "retry" {
		t.Fatalf("unexpected upload metadata: %#v", result)
	}
	if repo.countCalls != 1 {
		t.Fatalf("expected one active-count query, got %d", repo.countCalls)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected one create attempt, got %d", repo.createCalls)
	}
	if repo.created == nil {
		t.Fatal("expected create placeholder payload to be captured")
	}
	if repo.created.NameNormalized != "retry" {
		t.Fatalf("expected normalized name to be stored, got %q", repo.created.NameNormalized)
	}
	if repo.created.UploadExpiresAt.Before(time.Now().UTC().Add(100 * time.Second)) {
		t.Fatalf("expected refreshed expiry in the future, got %v", repo.created.UploadExpiresAt)
	}
}
