package message

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type fakeAttachmentRepo struct {
	selected      []model.Attachment
	selectErr     error
	selectChannel int64
	selectIDs     []int64
	selectCalls   int
}

func (f *fakeAttachmentRepo) CreateAttachment(ctx context.Context, id, channelId, authorId, ttlSeconds, fileSize int64, name string) error {
	return nil
}
func (f *fakeAttachmentRepo) RemoveAttachment(ctx context.Context, id, channelId int64) error {
	return nil
}
func (f *fakeAttachmentRepo) GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error) {
	return model.Attachment{}, nil
}
func (f *fakeAttachmentRepo) DoneAttachment(ctx context.Context, id, channelId int64, contentType, url, previewURL *string, height, width, fileSize *int64, name *string, authorId *int64) error {
	return nil
}
func (f *fakeAttachmentRepo) SelectAttachmentsByChannel(ctx context.Context, channelId int64, ids []int64) ([]model.Attachment, error) {
	f.selectCalls++
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

func TestValidateMessageAttachmentsUsesChannelScopedLookup(t *testing.T) {
	userID := int64(42)
	repo := &fakeAttachmentRepo{selected: []model.Attachment{
		{Id: 2, ChannelId: 99, Done: true, AuthorId: &userID},
		{Id: 1, ChannelId: 99, Done: true, AuthorId: &userID},
	}}
	e := &entity{at: repo}

	attachments, err := e.validateMessageAttachments(context.Background(), 99, userID, []int64{1, 2})
	if err != nil {
		t.Fatalf("validateMessageAttachments returned error: %v", err)
	}
	if repo.selectCalls != 1 || repo.selectChannel != 99 {
		t.Fatalf("expected channel-scoped lookup, got calls=%d channel=%d", repo.selectCalls, repo.selectChannel)
	}
	if len(repo.selectIDs) != 2 || repo.selectIDs[0] != 1 || repo.selectIDs[1] != 2 {
		t.Fatalf("unexpected selected ids: %#v", repo.selectIDs)
	}
	if len(attachments) != 2 || attachments[0].Id != 1 || attachments[1].Id != 2 {
		t.Fatalf("expected request order to be preserved, got %#v", attachments)
	}
}

func TestValidateMessageAttachmentsRejectsInvalidSets(t *testing.T) {
	ownerID := int64(7)
	otherID := int64(8)
	tests := []struct {
		name        string
		requested   []int64
		selected    []model.Attachment
		wantNoQuery bool
	}{
		{
			name:        "duplicate ids",
			requested:   []int64{1, 1},
			selected:    []model.Attachment{{Id: 1, ChannelId: 5, Done: true, AuthorId: &ownerID}},
			wantNoQuery: true,
		},
		{
			name:      "missing row",
			requested: []int64{1, 2},
			selected:  []model.Attachment{{Id: 1, ChannelId: 5, Done: true, AuthorId: &ownerID}},
		},
		{
			name:      "pending attachment",
			requested: []int64{1},
			selected:  []model.Attachment{{Id: 1, ChannelId: 5, Done: false, AuthorId: &ownerID}},
		},
		{
			name:      "foreign attachment",
			requested: []int64{1},
			selected:  []model.Attachment{{Id: 1, ChannelId: 5, Done: true, AuthorId: &otherID}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAttachmentRepo{selected: tt.selected}
			e := &entity{at: repo}

			_, err := e.validateMessageAttachments(context.Background(), 5, ownerID, tt.requested)
			var fiberErr *fiber.Error
			if !errors.As(err, &fiberErr) || fiberErr.Code != fiber.StatusBadRequest {
				t.Fatalf("expected bad request fiber error, got %v", err)
			}
			if tt.wantNoQuery && repo.selectCalls != 0 {
				t.Fatalf("expected no repo lookup, got %d calls", repo.selectCalls)
			}
		})
	}
}
