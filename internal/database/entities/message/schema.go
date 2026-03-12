package message

import (
	"context"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gocql/gocql"
)

const (
	createMessage                 = `INSERT INTO gochat.messages (channel_id, bucket, id, user_id, content, position, attachments, embeds, auto_embeds, flags, type, reference_channel, reference, thread) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	createSystemMessage           = `INSERT INTO gochat.messages (channel_id, bucket, id, user_id, content, position, flags, type) VALUES (?, ?, ?, ?, ?, ?, 0, ?)`
	claimThread                   = `INSERT INTO gochat.message_threads (channel_id, message_id, thread_id) VALUES (?, ?, ?) IF NOT EXISTS`
	createThreadCreatedMessageRef = `INSERT INTO gochat.thread_created_messages (thread_id, channel_id, message_id) VALUES (?, ?, ?)`
	deleteThreadCreatedMessageRef = `DELETE FROM gochat.thread_created_messages WHERE thread_id = ?`
	getThreadCreatedMessageRef    = `SELECT channel_id, message_id FROM gochat.thread_created_messages WHERE thread_id = ?`
	releaseThreadClaim            = `DELETE FROM gochat.message_threads WHERE channel_id = ? AND message_id = ?`
	setThread                     = `UPDATE gochat.messages SET thread = ? WHERE channel_id = ? AND id = ? AND bucket = ?`
	updateMessageContent          = `UPDATE gochat.messages SET content = ? WHERE channel_id = ? AND id = ? AND bucket = ?`
	updateMessage                 = `UPDATE gochat.messages SET content = ?, embeds = ?, auto_embeds = ?, flags = ?, edited_at = toTimestamp(now()) WHERE channel_id = ? AND id = ? AND bucket = ?`
	updateGeneratedEmbeds         = `UPDATE gochat.messages SET auto_embeds = ? WHERE channel_id = ? AND id = ? AND bucket = ?`
	deleteMessage                 = `DELETE FROM gochat.messages WHERE channel_id = ? AND bucket = ? AND id = ?`
	deleteChannelMessages         = `DELETE FROM gochat.messages WHERE channel_id = ? AND bucket IN ?`
	getMessage                    = `SELECT id, channel_id, user_id, content, position, attachments, embeds, auto_embeds, flags, edited_at, type, reference_channel, reference, thread FROM gochat.messages WHERE id = ? AND channel_id = ? AND bucket = ?`
	getMessagesBefore             = `SELECT id, channel_id, user_id, content, position, attachments, embeds, auto_embeds, flags, edited_at, type, reference_channel, reference, thread FROM gochat.messages WHERE channel_id = ? AND id <= ? AND bucket = ? ORDER BY id DESC LIMIT ?`
	getMessagesAfter              = `SELECT id, channel_id, user_id, content, position, attachments, embeds, auto_embeds, flags, edited_at, type, reference_channel, reference, thread FROM gochat.messages WHERE channel_id = ? AND id >= ? AND bucket = ? ORDER BY id LIMIT ?`
	getMessagesList               = `SELECT id, channel_id, user_id, content, position, attachments, embeds, auto_embeds, flags, edited_at, type, reference_channel, reference, thread FROM gochat.messages WHERE id IN ?`
	getMessagesByIds              = `SELECT id, channel_id, user_id, content, position, attachments, embeds, auto_embeds, flags, edited_at, type, reference_channel, reference, thread FROM gochat.messages WHERE channel_id = ? AND bucket = ? AND id IN ?;
`
)

func (e *Entity) CreateMessage(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, position int64) error {
	return e.CreateMessageWithMeta(ctx, id, channelID, userID, content, attachments, embedsJSON, autoEmbedsJSON, 0, model.MessageTypeChat, 0, 0, 0, position)
}

func (e *Entity) CreateMessageWithMeta(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, flags int, msgType model.MessageType, referenceChannel, reference, thread, position int64) error {
	err := e.c.Session().
		Query(createMessage).
		WithContext(ctx).
		Bind(channelID, idgen.GetBucket(id), id, userID, content, position, attachments, embedsJSON, autoEmbedsJSON, flags, int(msgType), referenceChannel, reference, thread).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create message: %w", err)
	}
	return nil
}

func (e *Entity) CreateSystemMessage(ctx context.Context, id, channelID, userID int64, content string, msgType model.MessageType, position int64) error {
	err := e.c.Session().
		Query(createSystemMessage).
		WithContext(ctx).
		Bind(channelID, idgen.GetBucket(id), id, userID, content, position, int(msgType)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create message: %w", err)
	}
	return nil
}

func (e *Entity) CreateThreadCreatedMessageRef(ctx context.Context, threadID, channelID, messageID int64) error {
	err := e.c.Session().
		Query(createThreadCreatedMessageRef).
		WithContext(ctx).
		Bind(threadID, channelID, messageID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create thread-created message ref: %w", err)
	}
	return nil
}

func (e *Entity) ClaimThread(ctx context.Context, channelID, messageID, threadID int64) (bool, int64, error) {
	result := make(map[string]interface{})
	applied, err := e.c.Session().
		Query(claimThread).
		WithContext(ctx).
		Bind(channelID, messageID, threadID).
		MapScanCAS(result)
	if err != nil {
		return false, 0, fmt.Errorf("unable to claim message thread: %w", err)
	}
	if applied {
		return true, threadID, nil
	}

	var currentThread int64
	switch value := result["thread_id"].(type) {
	case nil:
		currentThread = 0
	case int64:
		currentThread = value
	case int:
		currentThread = int64(value)
	case int32:
		currentThread = int64(value)
	default:
		return false, 0, fmt.Errorf("unexpected thread CAS value type %T", value)
	}

	return false, currentThread, nil
}

func (e *Entity) DeleteThreadCreatedMessageRef(ctx context.Context, threadID int64) error {
	err := e.c.Session().
		Query(deleteThreadCreatedMessageRef).
		WithContext(ctx).
		Bind(threadID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to delete thread-created message ref: %w", err)
	}
	return nil
}

func (e *Entity) ReleaseThreadClaim(ctx context.Context, channelID, messageID int64) error {
	err := e.c.Session().
		Query(releaseThreadClaim).
		WithContext(ctx).
		Bind(channelID, messageID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to release message thread claim: %w", err)
	}
	return nil
}

func (e *Entity) GetThreadCreatedMessageRef(ctx context.Context, threadID int64) (int64, int64, error) {
	var channelID int64
	var messageID int64
	err := e.c.Session().
		Query(getThreadCreatedMessageRef).
		WithContext(ctx).
		Bind(threadID).
		Scan(&channelID, &messageID)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to get thread-created message ref: %w", err)
	}
	return channelID, messageID, nil
}

func (e *Entity) SetThread(ctx context.Context, id, channelID, threadID int64) error {
	err := e.c.Session().
		Query(setThread).
		WithContext(ctx).
		Bind(threadID, channelID, id, idgen.GetBucket(id)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set message thread: %w", err)
	}
	return nil
}

func (e *Entity) UpdateMessageContent(ctx context.Context, id, channelID int64, content string) error {
	err := e.c.Session().
		Query(updateMessageContent).
		WithContext(ctx).
		Bind(content, channelID, id, idgen.GetBucket(id)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update message content: %w", err)
	}
	return nil
}

func (e *Entity) UpdateMessage(ctx context.Context, id, channelID int64, content, embedsJSON, autoEmbedsJSON string, flags int) error {
	err := e.c.Session().
		Query(updateMessage).
		WithContext(ctx).
		Bind(content, embedsJSON, autoEmbedsJSON, flags, channelID, id, idgen.GetBucket(id)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update message: %w", err)
	}
	return nil
}

func (e *Entity) UpdateGeneratedEmbeds(ctx context.Context, id, channelID int64, autoEmbedsJSON string) error {
	err := e.c.Session().
		Query(updateGeneratedEmbeds).
		WithContext(ctx).
		Bind(autoEmbedsJSON, channelID, id, idgen.GetBucket(id)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update generated embeds: %w", err)
	}
	return nil
}

func (e *Entity) DeleteMessage(ctx context.Context, id, channelID int64) error {
	err := e.c.Session().
		Query(deleteMessage).
		WithContext(ctx).
		Bind(channelID, idgen.GetBucket(id), id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to delete message: %w", err)
	}
	return nil
}

func (e *Entity) DeleteChannelMessages(ctx context.Context, channelID, lastID int64) error {
	first, last := idgen.GetBucket(channelID), idgen.GetBucket(lastID)
	length := last - first + 1
	buckets := make([]int64, length)
	for i := int64(0); i < length; i++ {
		buckets[i] = first + i
	}
	err := e.c.Session().
		Query(deleteChannelMessages).
		WithContext(ctx).
		Bind(channelID, buckets).
		Exec()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return fmt.Errorf("unable to delete messages: %w", err)
	}
	return nil
}

func (e *Entity) GetMessage(ctx context.Context, id, channelID int64) (model.Message, error) {
	var m model.Message
	err := e.c.Session().
		Query(getMessage).
		WithContext(ctx).
		Bind(id, channelID, idgen.GetBucket(id)).
		Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Position, &m.Attachments, &m.EmbedsJSON, &m.AutoEmbedsJSON, &m.Flags, &m.EditedAt, &m.Type, &m.ReferenceChannel, &m.Reference, &m.Thread)
	if err != nil {
		return m, fmt.Errorf("unable to get message: %w", err)
	}
	return m, nil
}

func (e *Entity) GetMessagesBefore(ctx context.Context, channelID, msgID int64, limit int) ([]model.Message, []int64, error) {
	var msgs []model.Message
	users := make(map[int64]bool)
	if msgID <= channelID {
		return msgs, nil, nil
	}
	lastBucket := idgen.GetBucket(msgID)
	endBucket := idgen.GetBucket(channelID)
	for {
		iter := e.c.Session().
			Query(getMessagesBefore).
			WithContext(ctx).
			Bind(channelID, msgID, lastBucket, limit-len(msgs)).
			Iter()
		var m model.Message
		for iter.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Position, &m.Attachments, &m.EmbedsJSON, &m.AutoEmbedsJSON, &m.Flags, &m.EditedAt, &m.Type, &m.ReferenceChannel, &m.Reference, &m.Thread) {
			msgs = append(msgs, cloneMessageRow(m))
			users[m.UserId] = true
		}
		if err := iter.Close(); err != nil {
			return nil, nil, fmt.Errorf("unable to get messages before: %w", err)
		}
		if len(msgs) == limit || lastBucket <= endBucket {
			break
		}
		lastBucket--
	}
	var userIDs []int64
	for id := range users {
		userIDs = append(userIDs, id)
	}
	return msgs, userIDs, nil
}

func (e *Entity) GetMessagesAfter(ctx context.Context, channelID, msgID, lastChannelMessage int64, limit int) ([]model.Message, []int64, error) {
	var msgs []model.Message
	users := make(map[int64]bool)
	if msgID <= channelID {
		return msgs, nil, nil
	}
	lastBucket := idgen.GetBucket(msgID)
	endBucket := idgen.GetBucket(lastChannelMessage)
	for {
		iter := e.c.Session().
			Query(getMessagesAfter).
			WithContext(ctx).
			Bind(channelID, msgID, lastBucket, limit-len(msgs)).
			Iter()
		var m model.Message
		for iter.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Position, &m.Attachments, &m.EmbedsJSON, &m.AutoEmbedsJSON, &m.Flags, &m.EditedAt, &m.Type, &m.ReferenceChannel, &m.Reference, &m.Thread) {
			msgs = append(msgs, cloneMessageRow(m))
			users[m.UserId] = true
		}
		if err := iter.Close(); err != nil {
			return nil, nil, fmt.Errorf("unable to get messages before: %w", err)
		}
		if len(msgs) == limit || lastBucket >= endBucket {
			break
		}
		lastBucket++
	}
	var userIDs []int64
	for id := range users {
		userIDs = append(userIDs, id)
	}
	return msgs, userIDs, nil
}

func (e *Entity) GetMessagesAround(ctx context.Context, channelID, msgID, lastChannelMessage int64, limit int) ([]model.Message, []int64, error) {
	beforeMsgs, beforeUserIDs, err := e.GetMessagesBefore(ctx, channelID, msgID, limit/2)
	if err != nil {
		return nil, nil, err
	}

	afterMsgs, afterUserIDs, err := e.GetMessagesAfter(ctx, channelID, msgID, lastChannelMessage, limit/2)
	if err != nil {
		return nil, nil, err
	}

	var msgs []model.Message
	if len(afterMsgs) > 1 {
		msgs = append(beforeMsgs, afterMsgs[1:]...)
	} else {
		msgs = beforeMsgs
	}

	return msgs, append(beforeUserIDs, afterUserIDs...), nil
}

func (e *Entity) GetMessagesList(ctx context.Context, msgIDs []int64) ([]model.Message, error) {
	var msgs []model.Message
	iter := e.c.Session().
		Query(getMessagesList).
		WithContext(ctx).
		Bind(msgIDs).
		Iter()
	var m model.Message
	for iter.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Position, &m.Attachments, &m.EmbedsJSON, &m.AutoEmbedsJSON, &m.Flags, &m.EditedAt, &m.Type, &m.ReferenceChannel, &m.Reference, &m.Thread) {
		msgs = append(msgs, cloneMessageRow(m))
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("unable to get messages list: %w", err)
	}
	return msgs, nil
}

func (e *Entity) GetChannelMessagesByIDs(ctx context.Context, channelID int64, ids []int64) ([]model.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	byBucket := make(map[int][]int64)
	for _, id := range ids {
		bucket := int(idgen.GetBucket(id))
		byBucket[bucket] = append(byBucket[bucket], id)
	}

	const maxIN = 100
	chunk := func(src []int64, size int) [][]int64 {
		var out [][]int64
		for len(src) > 0 {
			n := size
			if len(src) < n {
				n = len(src)
			}
			out = append(out, src[:n])
			src = src[n:]
		}
		return out
	}

	results := make([]model.Message, 0, len(ids))
	for bucket, idList := range byBucket {
		for _, part := range chunk(idList, maxIN) {
			iter := e.c.Session().
				Query(getMessagesByIds, channelID, bucket, part).
				WithContext(ctx).
				Iter()

			var m model.Message
			for iter.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Position, &m.Attachments, &m.EmbedsJSON, &m.AutoEmbedsJSON, &m.Flags, &m.EditedAt, &m.Type, &m.ReferenceChannel, &m.Reference, &m.Thread) {
				results = append(results, cloneMessageRow(m))
			}
			if err := iter.Close(); err != nil {
				return nil, fmt.Errorf("unable to get messages list: %w", err)
			}
		}
	}

	byID := make(map[int64]model.Message, len(results))
	for _, message := range results {
		byID[message.Id] = message
	}

	out := make([]model.Message, 0, len(ids))
	for _, id := range ids {
		if message, ok := byID[id]; ok {
			out = append(out, message)
		}
	}
	return out, nil
}

func cloneMessageRow(m model.Message) model.Message {
	mm := m
	if mm.Attachments != nil {
		mm.Attachments = append([]int64(nil), mm.Attachments...)
	}
	if mm.EmbedsJSON != nil {
		embedsJSON := *mm.EmbedsJSON
		mm.EmbedsJSON = &embedsJSON
	}
	if mm.AutoEmbedsJSON != nil {
		autoEmbedsJSON := *mm.AutoEmbedsJSON
		mm.AutoEmbedsJSON = &autoEmbedsJSON
	}
	if mm.Flags != nil {
		flags := *mm.Flags
		mm.Flags = &flags
	}
	return mm
}
