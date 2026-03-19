package guild

import (
	"context"
	"errors"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
)

func TestValidatePatchChannelRequestForThread(t *testing.T) {
	e := &entity{}
	thread := &model.Channel{Type: model.ChannelTypeThread}

	if err := e.validatePatchChannelRequest(thread, &PatchGuildChannelRequest{Name: strPtr("thread title"), Closed: boolPtr(true)}); err != nil {
		t.Fatalf("expected thread patch to be allowed, got %v", err)
	}

	err := e.validatePatchChannelRequest(thread, &PatchGuildChannelRequest{Private: boolPtr(true)})
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != fiber.StatusBadRequest {
		t.Fatalf("expected bad request for private thread patch, got %v", err)
	}
}

func TestValidatePatchChannelRequestForGuildChannel(t *testing.T) {
	e := &entity{}
	channel := &model.Channel{Type: model.ChannelTypeGuild}

	if err := e.validatePatchChannelRequest(channel, &PatchGuildChannelRequest{Name: strPtr("general_chat")}); err != nil {
		t.Fatalf("expected guild channel patch to be allowed, got %v", err)
	}

	err := e.validatePatchChannelRequest(channel, &PatchGuildChannelRequest{Closed: boolPtr(true)})
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) || fiberErr.Code != fiber.StatusBadRequest {
		t.Fatalf("expected bad request for closed non-thread patch, got %v", err)
	}
}

func strPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func TestChannelModelToDTOWithThreadMemberIncludesMember(t *testing.T) {
	thread := &model.Channel{Id: 1, Type: model.ChannelTypeThread, Name: "thread", MessageCount: 12}
	member := &dto.ThreadMember{UserId: 42, Flags: 7}

	got := channelModelToDTOWithThreadMember(thread, nil, 0, nil, member, []int64{42, 77})
	if got.Member == nil {
		t.Fatal("expected thread member to be included in dto")
	}
	if got.Member.UserId != 42 || got.Member.Flags != 7 {
		t.Fatalf("unexpected thread member payload: %#v", got.Member)
	}
	if len(got.MemberIds) != 2 || got.MemberIds[0] != 42 || got.MemberIds[1] != 77 {
		t.Fatalf("unexpected thread member ids payload: %#v", got.MemberIds)
	}
	if got.MessageCount == nil || *got.MessageCount != 12 {
		t.Fatalf("expected thread message count to be included, got %#v", got.MessageCount)
	}
}

func TestDetachDeletedThreadMessagesDetachesSourceAndFollowup(t *testing.T) {
	const (
		guildID         int64 = 77
		parentChannelID int64 = 10
		threadID        int64 = 555
		sourceMessageID int64 = 100
		followupMsgID   int64 = 101
	)

	msgRepo := &fakeDetachMessageRepo{
		threadRefs: map[int64]fakeThreadCreatedRef{
			threadID: {channelID: parentChannelID, messageID: followupMsgID},
		},
		messages: map[fakeDetachMessageKey]model.Message{
			{channelID: parentChannelID, messageID: sourceMessageID}: {
				Id:        sourceMessageID,
				ChannelId: parentChannelID,
				UserId:    1,
				Content:   "source message",
				Type:      int(model.MessageTypeChat),
				Thread:    threadID,
			},
			{channelID: parentChannelID, messageID: followupMsgID}: {
				Id:               followupMsgID,
				ChannelId:        parentChannelID,
				UserId:           2,
				Content:          "Thread Name",
				Type:             int(model.MessageTypeThreadCreated),
				Reference:        sourceMessageID,
				ReferenceChannel: parentChannelID,
				Thread:           threadID,
			},
		},
	}
	transport := &fakeDetachTransport{}
	e := &entity{
		msg: msgRepo,
		mqt: transport,
	}

	if err := e.detachDeletedThreadMessages(context.Background(), guildID, threadID); err != nil {
		t.Fatalf("detachDeletedThreadMessages returned error: %v", err)
	}

	if len(msgRepo.setThreadCalls) != 2 {
		t.Fatalf("expected 2 SetThread calls, got %d", len(msgRepo.setThreadCalls))
	}
	if msgRepo.setThreadCalls[0] != (fakeSetThreadCall{id: sourceMessageID, channelID: parentChannelID, threadID: 0}) {
		t.Fatalf("unexpected source SetThread call: %#v", msgRepo.setThreadCalls[0])
	}
	if msgRepo.setThreadCalls[1] != (fakeSetThreadCall{id: followupMsgID, channelID: parentChannelID, threadID: 0}) {
		t.Fatalf("unexpected followup SetThread call: %#v", msgRepo.setThreadCalls[1])
	}
	if len(msgRepo.releaseCalls) != 1 || msgRepo.releaseCalls[0] != (fakeDetachMessageKey{channelID: parentChannelID, messageID: sourceMessageID}) {
		t.Fatalf("unexpected ReleaseThreadClaim calls: %#v", msgRepo.releaseCalls)
	}
	if len(msgRepo.deletedRefs) != 1 || msgRepo.deletedRefs[0] != threadID {
		t.Fatalf("unexpected DeleteThreadCreatedMessageRef calls: %#v", msgRepo.deletedRefs)
	}

	sourceMessage := msgRepo.messages[fakeDetachMessageKey{channelID: parentChannelID, messageID: sourceMessageID}]
	if sourceMessage.Thread != 0 {
		t.Fatalf("expected source message thread to be cleared, got %d", sourceMessage.Thread)
	}
	followupMessage := msgRepo.messages[fakeDetachMessageKey{channelID: parentChannelID, messageID: followupMsgID}]
	if followupMessage.Thread != 0 {
		t.Fatalf("expected followup message thread to be cleared, got %d", followupMessage.Thread)
	}

	if len(transport.updates) != 2 {
		t.Fatalf("expected 2 message update events, got %d", len(transport.updates))
	}
	updatesByMessage := make(map[int64]dto.Message, len(transport.updates))
	for _, update := range transport.updates {
		if update.channelID != parentChannelID {
			t.Fatalf("expected parent channel updates only, got channel %d", update.channelID)
		}
		updatesByMessage[update.message.Id] = update.message
	}

	sourceUpdate, ok := updatesByMessage[sourceMessageID]
	if !ok {
		t.Fatal("expected source message update event")
	}
	if sourceUpdate.ThreadId != nil || sourceUpdate.Thread != nil {
		t.Fatalf("expected source update to have detached thread metadata, got %#v", sourceUpdate)
	}

	followupUpdate, ok := updatesByMessage[followupMsgID]
	if !ok {
		t.Fatal("expected followup message update event")
	}
	if followupUpdate.ThreadId != nil || followupUpdate.Thread != nil {
		t.Fatalf("expected followup update to have detached thread metadata, got %#v", followupUpdate)
	}
	if followupUpdate.Reference == nil || *followupUpdate.Reference != sourceMessageID {
		t.Fatalf("expected followup reference to be preserved, got %#v", followupUpdate.Reference)
	}
}

func TestSendUpdateChannelEventForThreadSendsThreadUpdate(t *testing.T) {
	guildID := int64(77)
	transport := &fakeGuildLifecycleTransport{}
	e := &entity{mqt: transport}
	parentID := int64(10)

	err := e.sendUpdateChannelEvent(guildID, dto.Channel{
		Id:       55,
		Type:     model.ChannelTypeThread,
		GuildId:  &guildID,
		Name:     "release discussion",
		ParentId: &parentID,
	})
	if err != nil {
		t.Fatalf("sendUpdateChannelEvent returned error: %v", err)
	}

	if len(transport.guildEvents) != 2 {
		t.Fatalf("expected 2 guild events, got %d", len(transport.guildEvents))
	}

	if _, ok := transport.guildEvents[0].(*mqmsg.UpdateChannel); !ok {
		t.Fatalf("expected first event to be UpdateChannel, got %T", transport.guildEvents[0])
	}
	threadUpdate, ok := transport.guildEvents[1].(*mqmsg.UpdateThread)
	if !ok {
		t.Fatalf("expected second event to be UpdateThread, got %T", transport.guildEvents[1])
	}
	if threadUpdate.Thread.Id != 55 || threadUpdate.Thread.Type != model.ChannelTypeThread {
		t.Fatalf("unexpected thread update payload: %#v", threadUpdate.Thread)
	}
}

func TestSendDeleteChannelEventForThreadSendsThreadDelete(t *testing.T) {
	const guildID int64 = 77
	transport := &fakeGuildLifecycleTransport{}
	e := &entity{mqt: transport}

	err := e.sendDeleteChannelEvent(guildID, &model.Channel{Id: 55, Type: model.ChannelTypeThread})
	if err != nil {
		t.Fatalf("sendDeleteChannelEvent returned error: %v", err)
	}

	if len(transport.guildEvents) != 2 {
		t.Fatalf("expected 2 guild events, got %d", len(transport.guildEvents))
	}

	if _, ok := transport.guildEvents[0].(*mqmsg.DeleteChannel); !ok {
		t.Fatalf("expected first event to be DeleteChannel, got %T", transport.guildEvents[0])
	}
	threadDelete, ok := transport.guildEvents[1].(*mqmsg.DeleteThread)
	if !ok {
		t.Fatalf("expected second event to be DeleteThread, got %T", transport.guildEvents[1])
	}
	if threadDelete.ThreadId != 55 {
		t.Fatalf("unexpected thread delete payload: %#v", threadDelete)
	}
}

type fakeDetachMessageKey struct {
	channelID int64
	messageID int64
}

type fakeThreadCreatedRef struct {
	channelID int64
	messageID int64
}

type fakeSetThreadCall struct {
	id        int64
	channelID int64
	threadID  int64
}

type fakeDetachMessageRepo struct {
	threadRefs     map[int64]fakeThreadCreatedRef
	messages       map[fakeDetachMessageKey]model.Message
	setThreadCalls []fakeSetThreadCall
	releaseCalls   []fakeDetachMessageKey
	deletedRefs    []int64
}

func (f *fakeDetachMessageRepo) CreateMessage(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, position int64) error {
	return nil
}

func (f *fakeDetachMessageRepo) CreateMessageWithMeta(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, flags int, msgType model.MessageType, referenceChannel, reference, thread, position int64) error {
	return nil
}

func (f *fakeDetachMessageRepo) CreateSystemMessage(ctx context.Context, id, channelId, userId int64, content string, msgType model.MessageType, position int64) error {
	return nil
}

func (f *fakeDetachMessageRepo) CreateThreadCreatedMessageRef(ctx context.Context, threadID, channelID, messageID int64) error {
	return nil
}

func (f *fakeDetachMessageRepo) ClaimThread(ctx context.Context, channelID, messageID, threadID int64) (bool, int64, error) {
	return false, 0, nil
}

func (f *fakeDetachMessageRepo) DeleteThreadCreatedMessageRef(ctx context.Context, threadID int64) error {
	f.deletedRefs = append(f.deletedRefs, threadID)
	delete(f.threadRefs, threadID)
	return nil
}

func (f *fakeDetachMessageRepo) ReleaseThreadClaim(ctx context.Context, channelID, messageID int64) error {
	f.releaseCalls = append(f.releaseCalls, fakeDetachMessageKey{channelID: channelID, messageID: messageID})
	return nil
}

func (f *fakeDetachMessageRepo) SetThread(ctx context.Context, id, channelID, threadID int64) error {
	f.setThreadCalls = append(f.setThreadCalls, fakeSetThreadCall{id: id, channelID: channelID, threadID: threadID})
	key := fakeDetachMessageKey{channelID: channelID, messageID: id}
	msg, ok := f.messages[key]
	if !ok {
		return gocql.ErrNotFound
	}
	msg.Thread = threadID
	f.messages[key] = msg
	return nil
}

func (f *fakeDetachMessageRepo) UpdateMessageContent(ctx context.Context, id, channelID int64, content string) error {
	return nil
}

func (f *fakeDetachMessageRepo) UpdateMessage(ctx context.Context, id, channelID int64, content, embedsJSON, autoEmbedsJSON string, flags int) error {
	return nil
}

func (f *fakeDetachMessageRepo) UpdateGeneratedEmbeds(ctx context.Context, id, channelID int64, autoEmbedsJSON string) error {
	return nil
}

func (f *fakeDetachMessageRepo) DeleteMessage(ctx context.Context, id, channelId int64) error {
	return nil
}

func (f *fakeDetachMessageRepo) DeleteChannelMessages(ctx context.Context, channelID, lastId int64) error {
	return nil
}

func (f *fakeDetachMessageRepo) GetMessage(ctx context.Context, id, channelId int64) (model.Message, error) {
	msg, ok := f.messages[fakeDetachMessageKey{channelID: channelId, messageID: id}]
	if !ok {
		return model.Message{}, gocql.ErrNotFound
	}
	return msg, nil
}

func (f *fakeDetachMessageRepo) GetMessagesBefore(ctx context.Context, channelId, msgId int64, limit int) ([]model.Message, []int64, error) {
	return nil, nil, nil
}

func (f *fakeDetachMessageRepo) GetMessagesAfter(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error) {
	return nil, nil, nil
}

func (f *fakeDetachMessageRepo) GetMessagesAround(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error) {
	return nil, nil, nil
}

func (f *fakeDetachMessageRepo) GetMessagesList(ctx context.Context, msgIds []int64) ([]model.Message, error) {
	return nil, nil
}

func (f *fakeDetachMessageRepo) GetChannelMessagesByIDs(ctx context.Context, channelId int64, ids []int64) ([]model.Message, error) {
	return nil, nil
}

func (f *fakeDetachMessageRepo) GetThreadCreatedMessageRef(ctx context.Context, threadID int64) (int64, int64, error) {
	ref, ok := f.threadRefs[threadID]
	if !ok {
		return 0, 0, gocql.ErrNotFound
	}
	return ref.channelID, ref.messageID, nil
}

type fakeDetachTransport struct {
	updates []fakeDetachTransportUpdate
}

type fakeDetachTransportUpdate struct {
	channelID int64
	message   dto.Message
}

func (f *fakeDetachTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	update, ok := message.(*mqmsg.UpdateMessage)
	if ok {
		f.updates = append(f.updates, fakeDetachTransportUpdate{
			channelID: channelId,
			message:   update.Message,
		})
	}
	return nil
}

func (f *fakeDetachTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeDetachTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}

type fakeGuildLifecycleTransport struct {
	guildEvents []mqmsg.EventDataMessage
}

func (f *fakeGuildLifecycleTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeGuildLifecycleTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	f.guildEvents = append(f.guildEvents, message)
	return nil
}

func (f *fakeGuildLifecycleTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}
