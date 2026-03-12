package message

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
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

type fakeReplyMessageRepo struct {
	getMessage    model.Message
	getMessageErr error
	createCalls   int
	lastCreate    struct {
		id               int64
		channelID        int64
		userID           int64
		content          string
		attachments      []int64
		embedsJSON       string
		autoEmbedsJSON   string
		flags            int
		msgType          model.MessageType
		referenceChannel int64
		reference        int64
		thread           int64
		position         int64
	}
}

func (f *fakeReplyMessageRepo) CreateMessage(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, position int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) CreateMessageWithMeta(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, flags int, msgType model.MessageType, referenceChannel, reference, thread, position int64) error {
	f.createCalls++
	f.lastCreate.id = id
	f.lastCreate.channelID = channelID
	f.lastCreate.userID = userID
	f.lastCreate.content = content
	f.lastCreate.attachments = append([]int64(nil), attachments...)
	f.lastCreate.embedsJSON = embedsJSON
	f.lastCreate.autoEmbedsJSON = autoEmbedsJSON
	f.lastCreate.flags = flags
	f.lastCreate.msgType = msgType
	f.lastCreate.referenceChannel = referenceChannel
	f.lastCreate.reference = reference
	f.lastCreate.thread = thread
	f.lastCreate.position = position
	return nil
}
func (f *fakeReplyMessageRepo) CreateSystemMessage(ctx context.Context, id, channelId, userId int64, content string, msgType model.MessageType, position int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) CreateThreadCreatedMessageRef(ctx context.Context, threadID, channelID, messageID int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) ClaimThread(ctx context.Context, channelID, messageID, threadID int64) (bool, int64, error) {
	return false, 0, nil
}
func (f *fakeReplyMessageRepo) DeleteThreadCreatedMessageRef(ctx context.Context, threadID int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) ReleaseThreadClaim(ctx context.Context, channelID, messageID int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) SetThread(ctx context.Context, id, channelID, threadID int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) UpdateMessageContent(ctx context.Context, id, channelID int64, content string) error {
	return nil
}
func (f *fakeReplyMessageRepo) UpdateMessage(ctx context.Context, id, channelID int64, content, embedsJSON, autoEmbedsJSON string, flags int) error {
	return nil
}
func (f *fakeReplyMessageRepo) UpdateGeneratedEmbeds(ctx context.Context, id, channelID int64, autoEmbedsJSON string) error {
	return nil
}
func (f *fakeReplyMessageRepo) DeleteMessage(ctx context.Context, id, channelId int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) DeleteChannelMessages(ctx context.Context, channelID, lastId int64) error {
	return nil
}
func (f *fakeReplyMessageRepo) GetMessage(ctx context.Context, id, channelId int64) (model.Message, error) {
	return f.getMessage, f.getMessageErr
}
func (f *fakeReplyMessageRepo) GetMessagesBefore(ctx context.Context, channelId, msgId int64, limit int) ([]model.Message, []int64, error) {
	return nil, nil, nil
}
func (f *fakeReplyMessageRepo) GetMessagesAfter(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error) {
	return nil, nil, nil
}
func (f *fakeReplyMessageRepo) GetMessagesAround(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error) {
	return nil, nil, nil
}
func (f *fakeReplyMessageRepo) GetMessagesList(ctx context.Context, msgIds []int64) ([]model.Message, error) {
	return nil, nil
}
func (f *fakeReplyMessageRepo) GetChannelMessagesByIDs(ctx context.Context, channelId int64, ids []int64) ([]model.Message, error) {
	return nil, nil
}
func (f *fakeReplyMessageRepo) GetThreadCreatedMessageRef(ctx context.Context, threadID int64) (int64, int64, error) {
	return 0, 0, nil
}

type fakeReplyChannelRepo struct {
	lastSetChannelID int64
	lastSetMessageID int64
}

func (f *fakeReplyChannelRepo) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	return model.Channel{}, nil
}
func (f *fakeReplyChannelRepo) GetChannelsBulk(ctx context.Context, ids []int64) ([]model.Channel, error) {
	return nil, nil
}
func (f *fakeReplyChannelRepo) GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error) {
	return nil, nil
}
func (f *fakeReplyChannelRepo) CreateChannel(ctx context.Context, id int64, name string, channelType model.ChannelType, parent *int64, permissions *int64, private bool) error {
	return nil
}
func (f *fakeReplyChannelRepo) DeleteChannel(ctx context.Context, id int64) error {
	return nil
}
func (f *fakeReplyChannelRepo) RenameChannel(ctx context.Context, id int64, newName string) error {
	return nil
}
func (f *fakeReplyChannelRepo) SetChannelPermissions(ctx context.Context, id int64, permissions int) error {
	return nil
}
func (f *fakeReplyChannelRepo) SetChannelPrivate(ctx context.Context, id int64, private bool) error {
	return nil
}
func (f *fakeReplyChannelRepo) SetChannelTopic(ctx context.Context, id int64, topic *string) error {
	return nil
}
func (f *fakeReplyChannelRepo) SetChannelParent(ctx context.Context, id int64, parent *int64) error {
	return nil
}
func (f *fakeReplyChannelRepo) SetChannelParentBulk(ctx context.Context, id []int64, parent *int64) error {
	return nil
}
func (f *fakeReplyChannelRepo) SetLastMessage(ctx context.Context, id, lastMessage int64) error {
	f.lastSetChannelID = id
	f.lastSetMessageID = lastMessage
	return nil
}
func (f *fakeReplyChannelRepo) AdjustMessageCount(ctx context.Context, id, delta int64) error {
	return nil
}
func (f *fakeReplyChannelRepo) UpdateChannel(ctx context.Context, id int64, parent *int64, private *bool, name, topic *string, closed *bool) (model.Channel, error) {
	return model.Channel{}, nil
}
func (f *fakeReplyChannelRepo) SetChannelVoiceRegion(ctx context.Context, id int64, region *string) error {
	return nil
}
func (f *fakeReplyChannelRepo) GetChannelVoiceRegion(ctx context.Context, id int64) (*string, error) {
	return nil, nil
}
func (f *fakeReplyChannelRepo) GetChannelMessagePosition(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}
func (f *fakeReplyChannelRepo) ReserveMessagePositions(ctx context.Context, id, count int64) (int64, error) {
	return count, nil
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

func TestDeriveThreadName(t *testing.T) {
	t.Run("uses explicit name when provided", func(t *testing.T) {
		name := deriveThreadName("  my thread  ", &model.Message{Id: 10, Content: "hello thread"}, "starter")
		if name != "my thread" {
			t.Fatalf("expected explicit thread name, got %q", name)
		}
	})

	t.Run("falls back to source message content", func(t *testing.T) {
		name := deriveThreadName("", &model.Message{Id: 10, Content: "  hello\nthread  "}, "starter")
		if name != "hello thread" {
			t.Fatalf("expected normalized source-message name, got %q", name)
		}
	})

	t.Run("falls back to starter content when source content is empty", func(t *testing.T) {
		name := deriveThreadName("", &model.Message{Id: 42, Content: " \n\t "}, "starter content")
		if name != "starter content" {
			t.Fatalf("expected starter-content fallback, got %q", name)
		}
	})

	t.Run("falls back to generated name when all content is empty", func(t *testing.T) {
		name := deriveThreadName("", &model.Message{Id: 42, Content: " \n\t "}, "  ")
		if name != "thread-42" {
			t.Fatalf("expected fallback thread name, got %q", name)
		}
	})

	t.Run("truncates to max thread name length", func(t *testing.T) {
		input := strings.Repeat("a", maxThreadNameLength+25)
		name := deriveThreadName("", &model.Message{Id: 1, Content: input}, "")
		if got := len([]rune(name)); got != maxThreadNameLength {
			t.Fatalf("expected %d runes, got %d", maxThreadNameLength, got)
		}
	})
}

func TestCloneChannelDTOPreservesThreadMember(t *testing.T) {
	joinedAt := time.Unix(123, 0).UTC()
	channel := &dto.Channel{
		Id:           1,
		Type:         model.ChannelTypeThread,
		Name:         "thread",
		Member:       &dto.ThreadMember{UserId: 55, JoinTimestamp: joinedAt, Flags: 3},
		MemberIds:    []int64{55, 99},
		MessageCount: int64Ptr(14),
	}

	cloned := cloneChannelDTO(channel)
	if cloned == nil || cloned.Member == nil {
		t.Fatal("expected cloned channel to keep thread member")
	}
	if cloned.Member.UserId != 55 || cloned.Member.Flags != 3 || !cloned.Member.JoinTimestamp.Equal(joinedAt) {
		t.Fatalf("unexpected cloned member payload: %#v", cloned.Member)
	}

	cloned.Member.Flags = 9
	if channel.Member.Flags != 3 {
		t.Fatalf("expected member clone to be independent, original=%d cloned=%d", channel.Member.Flags, cloned.Member.Flags)
	}

	cloned.MemberIds[0] = 88
	if channel.MemberIds[0] != 55 {
		t.Fatalf("expected member id slice clone to be independent, original=%v cloned=%v", channel.MemberIds, cloned.MemberIds)
	}

	*cloned.MessageCount = 20
	if *channel.MessageCount != 14 {
		t.Fatalf("expected message count clone to be independent, original=%d cloned=%d", *channel.MessageCount, *cloned.MessageCount)
	}
}

type fakeThreadMemberRepo struct {
	members []model.ThreadMember
}

func (f *fakeThreadMemberRepo) AddThreadMember(ctx context.Context, threadID, userID int64) (model.ThreadMember, error) {
	return model.ThreadMember{ThreadId: threadID, UserId: userID}, nil
}
func (f *fakeThreadMemberRepo) RemoveThreadMember(ctx context.Context, threadID, userID int64) error {
	return nil
}
func (f *fakeThreadMemberRepo) RemoveThreadMembers(ctx context.Context, threadID int64) error {
	return nil
}
func (f *fakeThreadMemberRepo) GetThreadMember(ctx context.Context, threadID, userID int64) (model.ThreadMember, error) {
	return model.ThreadMember{}, errors.New("not implemented")
}
func (f *fakeThreadMemberRepo) GetThreadMembers(ctx context.Context, threadID int64) ([]model.ThreadMember, error) {
	return append([]model.ThreadMember(nil), f.members...), nil
}
func (f *fakeThreadMemberRepo) GetThreadMembersBulk(ctx context.Context, threadIDs []int64) ([]model.ThreadMember, error) {
	return append([]model.ThreadMember(nil), f.members...), nil
}
func (f *fakeThreadMemberRepo) GetThreadMembersByUser(ctx context.Context, userID int64, threadIDs []int64) ([]model.ThreadMember, error) {
	return nil, nil
}
func (f *fakeThreadMemberRepo) GetUserThreadMembers(ctx context.Context, userID int64) ([]model.ThreadMember, error) {
	return nil, nil
}

type fakeMessageTransport struct {
	mu           sync.Mutex
	channelSends []int64
	guildSends   []int64
	userSends    []int64
	guildEvents  []mqmsg.EventDataMessage
}

func (f *fakeMessageTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.channelSends = append(f.channelSends, channelId)
	return nil
}

func (f *fakeMessageTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.guildSends = append(f.guildSends, guildId)
	f.guildEvents = append(f.guildEvents, message)
	return nil
}

func (f *fakeMessageTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.userSends = append(f.userSends, userId)
	return nil
}

type fakeMessageCache struct {
	deleted []string
}

var _ cache.Cache = (*fakeMessageCache)(nil)

func (f *fakeMessageCache) Set(ctx context.Context, key, val string) error { return nil }
func (f *fakeMessageCache) Get(ctx context.Context, key string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeMessageCache) Delete(ctx context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}
func (f *fakeMessageCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeMessageCache) SetTimed(ctx context.Context, key, val string, ttl int64) error {
	return nil
}
func (f *fakeMessageCache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	return nil
}
func (f *fakeMessageCache) SetInt64(ctx context.Context, key string, val int64) error { return nil }
func (f *fakeMessageCache) SetTTL(ctx context.Context, key string, ttl int64) error   { return nil }
func (f *fakeMessageCache) Incr(ctx context.Context, key string) (int64, error)       { return 0, nil }
func (f *fakeMessageCache) GetInt64(ctx context.Context, key string) (int64, error)   { return 0, nil }
func (f *fakeMessageCache) SetJSON(ctx context.Context, key string, val interface{}) error {
	return nil
}
func (f *fakeMessageCache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error {
	return nil
}
func (f *fakeMessageCache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error) {
	return true, nil
}
func (f *fakeMessageCache) GetJSON(ctx context.Context, key string, v interface{}) error {
	return errors.New("not implemented")
}
func (f *fakeMessageCache) HGet(ctx context.Context, key, field string) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeMessageCache) HSet(ctx context.Context, key, field, value string) error { return nil }
func (f *fakeMessageCache) HDel(ctx context.Context, key, field string) error        { return nil }
func (f *fakeMessageCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeMessageCache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	return nil
}

func TestSendMessageCreateEventForThreadTargetsJoinedUsersOnly(t *testing.T) {
	guildID := int64(77)
	transport := &fakeMessageTransport{}
	threadMembers := &fakeThreadMemberRepo{
		members: []model.ThreadMember{
			{ThreadId: 99, UserId: 1},
			{ThreadId: 99, UserId: 2},
			{ThreadId: 99, UserId: 3},
		},
	}
	e := &entity{
		mqt: transport,
		tm:  threadMembers,
	}

	e.sendMessageCreateEvent(&model.Channel{Id: 99, Type: model.ChannelTypeThread}, &guildID, dto.Message{
		Id:     500,
		Author: dto.User{Id: 1},
	})

	if len(transport.channelSends) != 1 || transport.channelSends[0] != 99 {
		t.Fatalf("expected one channel event for thread, got %#v", transport.channelSends)
	}
	if len(transport.guildSends) != 0 {
		t.Fatalf("expected no guild-wide activity event for thread messages, got %#v", transport.guildSends)
	}
	if len(transport.userSends) != 2 || transport.userSends[0] != 2 || transport.userSends[1] != 3 {
		t.Fatalf("expected thread activity to target joined non-author users, got %#v", transport.userSends)
	}
}

func TestSendThreadCreateEventsSendsGuildThreadLifecycleEvents(t *testing.T) {
	const (
		guildID         int64 = 77
		parentChannelID int64 = 10
		threadID        int64 = 99
	)

	transport := &fakeMessageTransport{}
	cacheStore := &fakeMessageCache{}
	parentID := parentChannelID
	now := time.Unix(1000, 0).UTC()
	e := &entity{
		mqt:   transport,
		tm:    &fakeThreadMemberRepo{},
		cache: cacheStore,
	}

	e.sendThreadCreateEvents(guildID, &model.Channel{Id: parentChannelID, Type: model.ChannelTypeGuild}, &threadCreateResult{
		Channel: &model.Channel{
			Id:           threadID,
			Type:         model.ChannelTypeThread,
			Name:         "release discussion",
			ParentID:     &parentID,
			LastMessage:  202,
			MessageCount: 2,
			CreatedAt:    now,
		},
		Position:  0,
		MemberIds: []int64{1},
		SourceMessage: dto.Message{
			Id:        200,
			ChannelId: parentChannelID,
			Author:    dto.User{Id: 1},
		},
		Initial: dto.Message{
			Id:        201,
			ChannelId: threadID,
			Author:    dto.User{Id: 1},
			Type:      int(model.MessageTypeThreadInitial),
		},
		Starter: dto.Message{
			Id:        202,
			ChannelId: threadID,
			Author:    dto.User{Id: 1},
			Type:      int(model.MessageTypeChat),
		},
		Followup: dto.Message{
			Id:        203,
			ChannelId: parentChannelID,
			Author:    dto.User{Id: 1},
			Type:      int(model.MessageTypeThreadCreated),
		},
	}, nil)

	var sawChannelCreate bool
	var sawThreadCreate bool
	for _, event := range transport.guildEvents {
		switch evt := event.(type) {
		case *mqmsg.CreateChannel:
			sawChannelCreate = evt.Channel.Id == threadID && evt.Channel.Type == model.ChannelTypeThread
		case *mqmsg.CreateThread:
			sawThreadCreate = evt.Thread.Id == threadID && evt.Thread.Type == model.ChannelTypeThread
		}
	}

	if !sawChannelCreate {
		t.Fatal("expected thread create flow to emit Channel Create for the thread")
	}
	if !sawThreadCreate {
		t.Fatal("expected thread create flow to emit Thread Create for the thread")
	}
	if len(cacheStore.deleted) != 1 || cacheStore.deleted[0] != "guild:77:channels" {
		t.Fatalf("expected guild channel cache invalidation, got %#v", cacheStore.deleted)
	}
}

func TestDeriveThreadCreationMessageContent(t *testing.T) {
	t.Run("trims surrounding whitespace", func(t *testing.T) {
		content := deriveThreadCreationMessageContent("  hello\nthread   content  ")
		if content != "hello\nthread   content" {
			t.Fatalf("expected trimmed content, got %q", content)
		}
	})

	t.Run("returns empty for empty name", func(t *testing.T) {
		content := deriveThreadCreationMessageContent(" \n\t ")
		if content != "" {
			t.Fatalf("expected empty content, got %q", content)
		}
	})
}

func int64Ptr(value int64) *int64 {
	return &value
}

func TestOptionalReferenceChannelID(t *testing.T) {
	t.Run("returns nil when there is no reference", func(t *testing.T) {
		if got := optionalReferenceChannelID(10, 20, 0); got != nil {
			t.Fatalf("expected nil, got %v", *got)
		}
	})

	t.Run("falls back to the message channel for legacy same-channel references", func(t *testing.T) {
		got := optionalReferenceChannelID(10, 0, 99)
		if got == nil || *got != 10 {
			t.Fatalf("expected fallback channel 10, got %v", got)
		}
	})

	t.Run("preserves explicit reference channel ids", func(t *testing.T) {
		got := optionalReferenceChannelID(10, 20, 99)
		if got == nil || *got != 20 {
			t.Fatalf("expected explicit channel 20, got %v", got)
		}
	})
}

func TestBuildMessageResponsePreservesNonce(t *testing.T) {
	var nonce helper.MessageNonce
	if err := json.Unmarshal([]byte(`"draft-1"`), &nonce); err != nil {
		t.Fatalf("failed to unmarshal nonce: %v", err)
	}

	e := &entity{}
	message, err := e.buildMessageResponse(&fiber.Ctx{}, 15, &model.Channel{Id: 9}, 33, &messageUserData{
		User:          &model.User{Id: 7, Name: "alice"},
		Discriminator: &model.Discriminator{Discriminator: "1234"},
	}, &SendMessageRequest{
		Content: "hello",
		Nonce:   &nonce,
	}, nil)
	if err != nil {
		t.Fatalf("buildMessageResponse returned error: %v", err)
	}
	if message.Nonce == nil || string(*message.Nonce) != `"draft-1"` {
		t.Fatalf("expected response nonce to be preserved, got %#v", message.Nonce)
	}
	if message.Position == nil || *message.Position != 33 {
		t.Fatalf("expected response position 33, got %#v", message.Position)
	}

	(*message.Nonce)[0] = 'x'
	if string(nonce) != `"draft-1"` {
		t.Fatalf("expected source nonce to stay unchanged, got %q", string(nonce))
	}
}

func TestBuildMessageResponseSetsReplyMetadata(t *testing.T) {
	reference := helper.StringInt64(42)
	e := &entity{}
	message, err := e.buildMessageResponse(&fiber.Ctx{}, 15, &model.Channel{Id: 9}, 18, &messageUserData{
		User:          &model.User{Id: 7, Name: "alice"},
		Discriminator: &model.Discriminator{Discriminator: "1234"},
	}, &SendMessageRequest{
		Content:   "reply",
		Reference: &reference,
	}, nil)
	if err != nil {
		t.Fatalf("buildMessageResponse returned error: %v", err)
	}
	if message.Type != int(model.MessageTypeReply) {
		t.Fatalf("expected reply type, got %d", message.Type)
	}
	if message.Reference == nil || *message.Reference != 42 {
		t.Fatalf("expected reference 42, got %#v", message.Reference)
	}
	if message.ReferenceChannelId == nil || *message.ReferenceChannelId != 9 {
		t.Fatalf("expected reference channel 9, got %#v", message.ReferenceChannelId)
	}
	if message.Position == nil || *message.Position != 18 {
		t.Fatalf("expected position 18, got %#v", message.Position)
	}
}

func TestValidateReplyReferenceRequiresSameChannel(t *testing.T) {
	reference := helper.StringInt64(77)
	repo := &fakeReplyMessageRepo{getMessageErr: gocql.ErrNotFound}
	e := &entity{msg: repo}

	err := e.validateReplyReference(context.Background(), 99, &SendMessageRequest{Reference: &reference})
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) {
		t.Fatalf("expected fiber error, got %v", err)
	}
	if fiberErr.Code != fiber.StatusBadRequest || fiberErr.Message != ErrReplyMustBeInSameChannel {
		t.Fatalf("unexpected fiber error: %#v", fiberErr)
	}
}

func TestCreateMessageWithCleanupStoresReplyMetadata(t *testing.T) {
	reference := helper.StringInt64(42)
	msgRepo := &fakeReplyMessageRepo{}
	chRepo := &fakeReplyChannelRepo{}
	e := &entity{
		msg: msgRepo,
		ch:  chRepo,
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	err := e.createMessageWithCleanup(c, 15, 9, 7, 41, &SendMessageRequest{
		Content:   "reply",
		Reference: &reference,
	})
	if err != nil {
		t.Fatalf("createMessageWithCleanup returned error: %v", err)
	}
	if msgRepo.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", msgRepo.createCalls)
	}
	if msgRepo.lastCreate.msgType != model.MessageTypeReply {
		t.Fatalf("expected reply message type, got %v", msgRepo.lastCreate.msgType)
	}
	if msgRepo.lastCreate.reference != 42 || msgRepo.lastCreate.referenceChannel != 9 {
		t.Fatalf("unexpected reply metadata: reference=%d reference_channel=%d", msgRepo.lastCreate.reference, msgRepo.lastCreate.referenceChannel)
	}
	if msgRepo.lastCreate.position != 41 {
		t.Fatalf("expected stored position 41, got %d", msgRepo.lastCreate.position)
	}
	if chRepo.lastSetChannelID != 9 || chRepo.lastSetMessageID != 15 {
		t.Fatalf("expected channel last message update, got channel=%d message=%d", chRepo.lastSetChannelID, chRepo.lastSetMessageID)
	}
}
