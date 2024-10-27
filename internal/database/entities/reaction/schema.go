package reaction

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getReactions      = `SELECT message_id, user_id, emote_id FROM gochat.reactions WHERE message_id = ? ORDER BY user_id LIMIT 10`
	getReactionsAfter = `SELECT message_id, user_id, emote_id FROM gochat.reactions WHERE message_id = ? AND token(user_id) > token(?) ORDER BY user_id LIMIT 10`
	addReaction       = `INSERT INTO gochat.reactions (message_id, user_id, emote_id) VALUES (?, ?, ?)`
	removeReaction    = `DELETE FROM gochat.reactions WHERE message_id = ? AND user_id = ?`
)

func (e *Entity) GetReactions(ctx context.Context, messageId int64) ([]model.Reaction, error) {
	var reactions []model.Reaction
	iter := e.c.Session().
		Query(getReactions).
		WithContext(ctx).
		Bind(messageId).
		Iter()
	var r model.Reaction
	for iter.Scan(&r.MessageId, &r.UserId, &r.EmoteId) {
		reactions = append(reactions, r)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get reactions: %w", err)
	}
	return reactions, nil
}

func (e *Entity) GetReactionsAfter(ctx context.Context, messageId, userId int64) ([]model.Reaction, error) {
	var reactions []model.Reaction
	iter := e.c.Session().
		Query(getReactionsAfter).
		WithContext(ctx).
		Bind(messageId, userId).
		Iter()
	var r model.Reaction
	for iter.Scan(&r.MessageId, &r.UserId, &r.EmoteId) {
		reactions = append(reactions, r)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get reactions: %w", err)
	}
	return reactions, nil
}

func (e *Entity) AddReaction(ctx context.Context, messageId, userId, emoteId int64) error {
	err := e.c.Session().
		Query(addReaction).
		WithContext(ctx).
		Bind(messageId, userId, emoteId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add reaction: %w", err)
	}
	return nil
}

func (e *Entity) RemoveReaction(ctx context.Context, messageId, userId int64) error {
	err := e.c.Session().
		Query(removeReaction).
		WithContext(ctx).
		Bind(messageId, userId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove reaction: %w", err)
	}
	return nil
}
