package message

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	icache "github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/cache/messagecache"
	"github.com/FlameInTheDark/gochat/internal/cache/testutil"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
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
	getCalls      int
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
	f.getCalls++
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
	channel          model.Channel
	getErr           error
	lastSetChannelID int64
	lastSetMessageID int64
}

func (f *fakeReplyChannelRepo) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	return f.channel, f.getErr
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

type fakeAccessMemberRepo struct {
	isMember bool
	err      error
}

func (f *fakeAccessMemberRepo) AddMember(ctx context.Context, userID, guildID int64) error {
	return nil
}
func (f *fakeAccessMemberRepo) RemoveMember(ctx context.Context, userID, guildID int64) error {
	return nil
}
func (f *fakeAccessMemberRepo) RemoveMembersByGuild(ctx context.Context, guildID int64) error {
	return nil
}
func (f *fakeAccessMemberRepo) GetMember(ctx context.Context, userId, guildId int64) (model.Member, error) {
	return model.Member{}, nil
}
func (f *fakeAccessMemberRepo) GetMembersList(ctx context.Context, guildId int64, ids []int64) ([]model.Member, error) {
	return nil, nil
}
func (f *fakeAccessMemberRepo) GetGuildMembers(ctx context.Context, guildId int64) ([]model.Member, error) {
	return nil, nil
}
func (f *fakeAccessMemberRepo) IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error) {
	return f.isMember, f.err
}
func (f *fakeAccessMemberRepo) GetUserGuilds(ctx context.Context, userId int64) ([]model.UserGuild, error) {
	return nil, nil
}
func (f *fakeAccessMemberRepo) SetTimeout(ctx context.Context, userId, guildId int64, timeout *time.Time) error {
	return nil
}
func (f *fakeAccessMemberRepo) CountGuildMembers(ctx context.Context, guildId int64) (int64, error) {
	return 0, nil
}

type fakeAccessGuildChannelsRepo struct {
	guildByChannel model.GuildChannel
	err            error
}

func (f *fakeAccessGuildChannelsRepo) AddChannel(ctx context.Context, guildID, channelID int64, channelName string, channelType model.ChannelType, parentID *int64, private bool, position int, topic *string, creatorID *int64, closed bool) error {
	return nil
}
func (f *fakeAccessGuildChannelsRepo) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	return f.guildByChannel, f.err
}
func (f *fakeAccessGuildChannelsRepo) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	return nil, nil
}
func (f *fakeAccessGuildChannelsRepo) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	return f.guildByChannel, f.err
}
func (f *fakeAccessGuildChannelsRepo) GetGuildChannelsByChannelIDs(ctx context.Context, channelIDs []int64) ([]model.GuildChannel, error) {
	return nil, nil
}
func (f *fakeAccessGuildChannelsRepo) RemoveChannel(ctx context.Context, guildID, channelID int64) error {
	return nil
}
func (f *fakeAccessGuildChannelsRepo) SetGuildChannelPosition(ctx context.Context, updates []model.GuildChannelUpdatePosition) error {
	return nil
}
func (f *fakeAccessGuildChannelsRepo) ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error {
	return nil
}
func (f *fakeAccessGuildChannelsRepo) GetGuildsChannelsIDsMany(ctx context.Context, guilds []int64) ([]int64, error) {
	return nil, nil
}

type fakeAccessDMRepo struct {
	isParticipant bool
	err           error
}

func (f *fakeAccessDMRepo) GetDmChannel(ctx context.Context, userId, participantId int64) (model.DMChannel, error) {
	return model.DMChannel{}, nil
}
func (f *fakeAccessDMRepo) CreateDmChannel(ctx context.Context, userId, participantId, channelId int64) error {
	return nil
}
func (f *fakeAccessDMRepo) IsDmChannelParticipant(ctx context.Context, channelId, userId int64) (bool, error) {
	return f.isParticipant, f.err
}
func (f *fakeAccessDMRepo) GetUserDmChannels(ctx context.Context, userId int64) ([]model.DMChannel, error) {
	return nil, nil
}
func (f *fakeAccessDMRepo) GetDmChannelByChannelId(ctx context.Context, channelId int64) ([]model.DMChannel, error) {
	return nil, nil
}

type fakeAccessGroupDMRepo struct {
	isParticipant bool
	err           error
}

func (f *fakeAccessGroupDMRepo) JoinGroupDmChannelMany(ctx context.Context, channelId int64, users []int64) error {
	return nil
}
func (f *fakeAccessGroupDMRepo) JoinGroupDmChannel(ctx context.Context, channelId, userId int64) error {
	return nil
}
func (f *fakeAccessGroupDMRepo) GetGroupDmChannel(ctx context.Context, channelId, userId int64) (model.GroupDMChannel, error) {
	return model.GroupDMChannel{}, nil
}
func (f *fakeAccessGroupDMRepo) LeaveGroupDmChannel(ctx context.Context, channelId, userId int64) error {
	return nil
}
func (f *fakeAccessGroupDMRepo) GetGroupDmParticipants(ctx context.Context, channelId int64) ([]model.GroupDMChannel, error) {
	return nil, nil
}
func (f *fakeAccessGroupDMRepo) IsGroupDmParticipant(ctx context.Context, channelId int64, userId int64) (bool, error) {
	return f.isParticipant, f.err
}
func (f *fakeAccessGroupDMRepo) GetUserGroupDmChannels(ctx context.Context, userId int64) ([]model.GroupDMChannel, error) {
	return nil, nil
}

type fakeDeleteRoleCheck struct {
	allowed bool
	calls   int
	gotPerm []permissions.RolePermission
}

func (f *fakeDeleteRoleCheck) getUserRoleIDs(ctx context.Context, guildID, userID int64) ([]int64, error) {
	return nil, nil
}
func (f *fakeDeleteRoleCheck) ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error) {
	f.calls++
	f.gotPerm = append([]permissions.RolePermission(nil), perm...)
	return &model.Channel{Id: channelID, Type: model.ChannelTypeGuild},
		&model.GuildChannel{GuildId: guildID, ChannelId: channelID},
		&model.Guild{Id: guildID},
		f.allowed,
		nil
}
func (f *fakeDeleteRoleCheck) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	return &model.Guild{Id: guildID}, f.allowed, nil
}
func (f *fakeDeleteRoleCheck) GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error) {
	return 0, nil
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

type fakeContextualUserTransport struct {
	ctxCh chan context.Context
}

var _ mq.ContextSendTransporter = (*fakeContextualUserTransport)(nil)

func (f *fakeContextualUserTransport) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeContextualUserTransport) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeContextualUserTransport) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeContextualUserTransport) SendChannelMessageContext(ctx context.Context, channelId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeContextualUserTransport) SendGuildUpdateContext(ctx context.Context, guildId int64, message mqmsg.EventDataMessage) error {
	return nil
}

func (f *fakeContextualUserTransport) SendUserUpdateContext(ctx context.Context, userId int64, message mqmsg.EventDataMessage) error {
	f.ctxCh <- ctx
	return nil
}

type fakeMessageCache struct {
	testutil.Noop
	deleted []string
}

func (f *fakeMessageCache) Delete(ctx context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}

type fakeWindowCache struct {
	testutil.Noop
	members   []string
	blobs     [][]byte
	batchVals []interface{}
}

func (f *fakeWindowCache) ZRevRangeByScore(ctx context.Context, key, max, min string, offset, count int64) ([]string, error) {
	return append([]string(nil), f.members...), nil
}

func (f *fakeWindowCache) MGetBytes(ctx context.Context, keys ...string) ([][]byte, error) {
	return append([][]byte(nil), f.blobs...), nil
}

func (f *fakeWindowCache) SetTimedJSONBatch(ctx context.Context, keys []string, vals []interface{}, ttl int64, opts ...icache.TimedOption) error {
	f.batchVals = append([]interface{}(nil), vals...)
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

	e.sendMessageCreateEvent(context.Background(), &model.Channel{Id: 99, Type: model.ChannelTypeThread}, &guildID, dto.Message{
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

	e.sendThreadCreateEvents(context.Background(), guildID, &model.Channel{Id: parentChannelID, Type: model.ChannelTypeGuild}, &threadCreateResult{
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

func TestSendReadStateUpdateAsyncPreservesRequestContext(t *testing.T) {
	transport := &fakeContextualUserTransport{ctxCh: make(chan context.Context, 1)}
	e := &entity{
		mqt: transport,
		log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	ctx := helper.ContextWithRequestID(context.Background(), "req-123")
	ctx = helper.ContextWithUserID(ctx, 42)

	e.sendReadStateUpdateAsync(ctx, 42, 77, 88)

	select {
	case sentCtx := <-transport.ctxCh:
		requestID, ok := helper.RequestIDFromContext(sentCtx)
		if !ok || requestID != "req-123" {
			t.Fatalf("expected propagated request id, got %q ok=%v", requestID, ok)
		}
		userID, ok := helper.UserIDFromContext(sentCtx)
		if !ok || userID != 42 {
			t.Fatalf("expected propagated user id, got %d ok=%v", userID, ok)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async read state update")
	}
}

func TestValidateReadPermissionsRejectsUnauthorizedDMParticipant(t *testing.T) {
	e := &entity{
		ch:  &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeDM}},
		dmc: &fakeAccessDMRepo{isParticipant: false},
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	_, _, err := e.validateReadPermissions(c, 9, 42)
	assertMessageFiberErrorCode(t, err, fiber.StatusForbidden)
}

func TestValidateSendPermissionsRejectsUnauthorizedGroupDMParticipant(t *testing.T) {
	e := &entity{
		ch:   &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGroupDM}},
		gdmc: &fakeAccessGroupDMRepo{isParticipant: false},
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	_, _, err := e.validateSendPermissions(c, 9, 42)
	assertMessageFiberErrorCode(t, err, fiber.StatusForbidden)
}

func TestValidateUploadPermissionsRejectsUnauthorizedGroupDMParticipant(t *testing.T) {
	e := &entity{
		ch:   &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGroupDM}},
		gdmc: &fakeAccessGroupDMRepo{isParticipant: false},
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	err := e.validateUploadPermissions(c, 9, 42)
	assertMessageFiberErrorCode(t, err, fiber.StatusForbidden)
}

func TestValidateMessageOwnershipRejectsFormerGuildMemberBeforeLookup(t *testing.T) {
	msgRepo := &fakeReplyMessageRepo{
		getMessage: model.Message{Id: 88, ChannelId: 9, UserId: 42, Type: int(model.MessageTypeChat)},
	}
	e := &entity{
		ch:  &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGuild}},
		gc:  &fakeAccessGuildChannelsRepo{guildByChannel: model.GuildChannel{GuildId: 77, ChannelId: 9}},
		m:   &fakeAccessMemberRepo{isMember: false},
		msg: msgRepo,
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	_, _, err := e.validateMessageOwnership(c, 88, 9, 42)
	assertMessageFiberErrorCode(t, err, fiber.StatusForbidden)
	if msgRepo.getCalls != 0 {
		t.Fatalf("expected access denial before message lookup, got %d lookups", msgRepo.getCalls)
	}
}

func TestValidateDeletePermissionRejectsFormerDMParticipantBeforeLookup(t *testing.T) {
	msgRepo := &fakeReplyMessageRepo{
		getMessage: model.Message{Id: 88, ChannelId: 9, UserId: 42, Type: int(model.MessageTypeChat)},
	}
	e := &entity{
		ch:  &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeDM}},
		dmc: &fakeAccessDMRepo{isParticipant: false},
		msg: msgRepo,
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	_, err := e.validateDeletePermission(c, 88, 9, 42)
	assertMessageFiberErrorCode(t, err, fiber.StatusForbidden)
	if msgRepo.getCalls != 0 {
		t.Fatalf("expected access denial before message lookup, got %d lookups", msgRepo.getCalls)
	}
}

func TestValidateDeletePermissionAllowsManageMessagesForOtherUserGuildMessage(t *testing.T) {
	msgRepo := &fakeReplyMessageRepo{
		getMessage: model.Message{Id: 88, ChannelId: 9, UserId: 99, Type: int(model.MessageTypeChat)},
	}
	permRepo := &fakeDeleteRoleCheck{allowed: true}
	e := &entity{
		ch:   &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGuild}},
		gc:   &fakeAccessGuildChannelsRepo{guildByChannel: model.GuildChannel{GuildId: 77, ChannelId: 9}},
		m:    &fakeAccessMemberRepo{isMember: true},
		msg:  msgRepo,
		perm: permRepo,
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	message, err := e.validateDeletePermission(c, 88, 9, 42)
	if err != nil {
		t.Fatalf("expected delete permission, got %v", err)
	}
	if message.Id != 88 {
		t.Fatalf("expected message 88, got %d", message.Id)
	}
	if permRepo.calls != 1 {
		t.Fatalf("expected one manage-messages permission check, got %d", permRepo.calls)
	}
	if len(permRepo.gotPerm) != 1 || permRepo.gotPerm[0] != permissions.PermTextManageMessages {
		t.Fatalf("expected manage-messages permission check, got %#v", permRepo.gotPerm)
	}
}

func TestValidateDeletePermissionRejectsOtherUserGuildMessageWithoutManageMessages(t *testing.T) {
	msgRepo := &fakeReplyMessageRepo{
		getMessage: model.Message{Id: 88, ChannelId: 9, UserId: 99, Type: int(model.MessageTypeChat)},
	}
	e := &entity{
		ch:   &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGuild}},
		gc:   &fakeAccessGuildChannelsRepo{guildByChannel: model.GuildChannel{GuildId: 77, ChannelId: 9}},
		m:    &fakeAccessMemberRepo{isMember: true},
		msg:  msgRepo,
		perm: &fakeDeleteRoleCheck{allowed: false},
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	_, err := e.validateDeletePermission(c, 88, 9, 42)
	assertMessageFiberErrorCode(t, err, fiber.StatusForbidden)
}

func TestValidateDeletePermissionDoesNotRequireManageMessagesForOwnMessage(t *testing.T) {
	msgRepo := &fakeReplyMessageRepo{
		getMessage: model.Message{Id: 88, ChannelId: 9, UserId: 42, Type: int(model.MessageTypeChat)},
	}
	permRepo := &fakeDeleteRoleCheck{allowed: false}
	e := &entity{
		ch:   &fakeReplyChannelRepo{channel: model.Channel{Id: 9, Type: model.ChannelTypeGuild}},
		gc:   &fakeAccessGuildChannelsRepo{guildByChannel: model.GuildChannel{GuildId: 77, ChannelId: 9}},
		m:    &fakeAccessMemberRepo{isMember: true},
		msg:  msgRepo,
		perm: permRepo,
	}
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	if _, err := e.validateDeletePermission(c, 88, 9, 42); err != nil {
		t.Fatalf("expected owner of message to delete without manage-messages check, got %v", err)
	}
	if permRepo.calls != 0 {
		t.Fatalf("expected no manage-messages check for own message, got %d", permRepo.calls)
	}
}

func assertMessageFiberErrorCode(t *testing.T, err error, want int) {
	t.Helper()
	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) {
		t.Fatalf("expected fiber error, got %v", err)
	}
	if fiberErr.Code != want {
		t.Fatalf("expected status %d, got %d", want, fiberErr.Code)
	}
}

func TestTryMessagesFromCacheSanitizesLegacyViewerFields(t *testing.T) {
	nonce := helper.MessageNonce([]byte(`"draft-1"`))
	blob, err := json.Marshal(dto.Message{
		Id:        1,
		ChannelId: 99,
		Nonce:     &nonce,
		Reactions: []dto.MessageReaction{{
			Count: 2,
			Me:    true,
			Emoji: dto.MessageReactionEmoji{Name: "wave"},
		}},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	members := make([]string, messagecache.WindowSize)
	blobs := make([][]byte, messagecache.WindowSize)
	for i := range members {
		members[i] = messagecache.IDToMember(int64(i + 1))
		blobs[i] = blob
	}

	cacheStore := &fakeWindowCache{members: members, blobs: blobs}
	e := &entity{cache: cacheStore}

	msgs, ok := e.tryMessagesFromCache(context.Background(), 99)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(msgs) != messagecache.WindowSize {
		t.Fatalf("expected %d messages, got %d", messagecache.WindowSize, len(msgs))
	}
	if msgs[0].Nonce != nil {
		t.Fatalf("expected nonce to be stripped from cached message, got %#v", msgs[0].Nonce)
	}
	if len(msgs[0].Reactions) != 1 || msgs[0].Reactions[0].Me {
		t.Fatalf("expected cached reaction state to be viewer-neutral, got %#v", msgs[0].Reactions)
	}
}

func TestBackfillMessagesCacheStripsViewerSpecificFields(t *testing.T) {
	nonce := helper.MessageNonce([]byte(`"draft-1"`))
	cacheStore := &fakeWindowCache{}
	e := &entity{cache: cacheStore}

	e.backfillMessagesCache(context.Background(), 99, []dto.Message{{
		Id:        1,
		ChannelId: 99,
		Nonce:     &nonce,
		Reactions: []dto.MessageReaction{{
			Count: 1,
			Me:    true,
			Emoji: dto.MessageReactionEmoji{Name: "wave"},
		}},
	}})

	if len(cacheStore.batchVals) != 1 {
		t.Fatalf("expected one cached message, got %d", len(cacheStore.batchVals))
	}
	cached, ok := cacheStore.batchVals[0].(dto.Message)
	if !ok {
		t.Fatalf("expected dto.Message in cache batch, got %T", cacheStore.batchVals[0])
	}
	if cached.Nonce != nil {
		t.Fatalf("expected nonce to be stripped before caching, got %#v", cached.Nonce)
	}
	if len(cached.Reactions) != 1 || cached.Reactions[0].Me {
		t.Fatalf("expected cached reactions to clear Me flags, got %#v", cached.Reactions)
	}
}
