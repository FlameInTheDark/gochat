package reaction

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gocql/gocql"
)

const (
	getUserReaction = `SELECT message_id, reaction_id, user_id, bucket_key, custom, emoji_id, emoji_name
FROM gochat.reactions_by_message_user WHERE message_id = ? AND user_id = ? AND bucket_key = ? LIMIT 1`
	getUserReactions = `SELECT message_id, reaction_id, user_id, bucket_key, custom, emoji_id, emoji_name
FROM gochat.reactions_by_message_user WHERE message_id = ? AND user_id = ?`
	listBucketReactions = `SELECT message_id, reaction_id, user_id, bucket_key, custom, emoji_id, emoji_name
FROM gochat.reaction_members_by_bucket WHERE message_id = ? AND bucket_key = ? LIMIT ?`
	listBucketReactionsAfter = `SELECT message_id, reaction_id, user_id, bucket_key, custom, emoji_id, emoji_name
FROM gochat.reaction_members_by_bucket WHERE message_id = ? AND bucket_key = ? AND reaction_id < ? LIMIT ?`
	listMessageSummaries = `SELECT message_id, bucket_key, custom, emoji_id, emoji_name, count
FROM gochat.reaction_summaries_by_message WHERE message_id = ?`
	upsertReactionByUser = `INSERT INTO gochat.reactions_by_message_user
(message_id, user_id, bucket_key, reaction_id, custom, emoji_id, emoji_name) VALUES (?, ?, ?, ?, ?, ?, ?)`
	upsertReactionByBucket = `INSERT INTO gochat.reaction_members_by_bucket
(message_id, bucket_key, reaction_id, user_id, custom, emoji_id, emoji_name) VALUES (?, ?, ?, ?, ?, ?, ?)`
	deleteReactionByUser   = `DELETE FROM gochat.reactions_by_message_user WHERE message_id = ? AND user_id = ? AND bucket_key = ?`
	deleteReactionByBucket = `DELETE FROM gochat.reaction_members_by_bucket WHERE message_id = ? AND bucket_key = ? AND reaction_id = ?`
	upsertSummary          = `INSERT INTO gochat.reaction_summaries_by_message
(message_id, bucket_key, custom, emoji_id, emoji_name, count) VALUES (?, ?, ?, ?, ?, ?)`
	deleteSummary = `DELETE FROM gochat.reaction_summaries_by_message WHERE message_id = ? AND bucket_key = ?`
)

func (e *Entity) GetUserReaction(ctx context.Context, messageId, userId int64, bucketKey string) (model.Reaction, error) {
	var reaction model.Reaction
	err := e.c.Session().
		Query(getUserReaction).
		WithContext(ctx).
		Bind(messageId, userId, bucketKey).
		Scan(&reaction.MessageId, &reaction.ReactionId, &reaction.UserId, &reaction.BucketKey, &reaction.Custom, &reaction.EmojiId, &reaction.EmojiName)
	if err != nil {
		return reaction, fmt.Errorf("unable to get user reaction: %w", err)
	}
	return reaction, nil
}

func (e *Entity) GetUserReactions(ctx context.Context, messageId, userId int64) ([]model.Reaction, error) {
	iter := e.c.Session().
		Query(getUserReactions).
		WithContext(ctx).
		Bind(messageId, userId).
		Iter()
	reactions, err := scanReactions(iter)
	if err != nil {
		return nil, fmt.Errorf("unable to get user reactions: %w", err)
	}
	return reactions, nil
}

func (e *Entity) ListBucketReactions(ctx context.Context, messageId int64, bucketKey string, after *int64, limit int) ([]model.Reaction, error) {
	query := e.c.Session().
		Query(listBucketReactions).
		WithContext(ctx)
	if after != nil {
		query = e.c.Session().
			Query(listBucketReactionsAfter).
			WithContext(ctx).
			Bind(messageId, bucketKey, *after, limit)
	} else {
		query = query.Bind(messageId, bucketKey, limit)
	}
	reactions, err := scanReactions(query.Iter())
	if err != nil {
		return nil, fmt.Errorf("unable to list bucket reactions: %w", err)
	}
	return reactions, nil
}

func (e *Entity) ListMessageSummaries(ctx context.Context, messageId int64) ([]model.ReactionSummary, error) {
	iter := e.c.Session().
		Query(listMessageSummaries).
		WithContext(ctx).
		Bind(messageId).
		Iter()
	var summaries []model.ReactionSummary
	var summary model.ReactionSummary
	for iter.Scan(&summary.MessageId, &summary.BucketKey, &summary.Custom, &summary.EmojiId, &summary.EmojiName, &summary.Count) {
		summaries = append(summaries, summary)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("unable to list message summaries: %w", err)
	}
	return summaries, nil
}

func (e *Entity) UpsertReaction(ctx context.Context, reaction model.Reaction) error {
	if err := e.c.Session().
		Query(upsertReactionByUser).
		WithContext(ctx).
		Bind(reaction.MessageId, reaction.UserId, reaction.BucketKey, reaction.ReactionId, reaction.Custom, reaction.EmojiId, reaction.EmojiName).
		Exec(); err != nil {
		return fmt.Errorf("unable to upsert user reaction: %w", err)
	}
	if err := e.c.Session().
		Query(upsertReactionByBucket).
		WithContext(ctx).
		Bind(reaction.MessageId, reaction.BucketKey, reaction.ReactionId, reaction.UserId, reaction.Custom, reaction.EmojiId, reaction.EmojiName).
		Exec(); err != nil {
		return fmt.Errorf("unable to upsert bucket reaction: %w", err)
	}
	return nil
}

func (e *Entity) DeleteReaction(ctx context.Context, reaction model.Reaction) error {
	if err := e.c.Session().
		Query(deleteReactionByUser).
		WithContext(ctx).
		Bind(reaction.MessageId, reaction.UserId, reaction.BucketKey).
		Exec(); err != nil {
		return fmt.Errorf("unable to delete user reaction: %w", err)
	}
	if reaction.ReactionId != 0 {
		if err := e.c.Session().
			Query(deleteReactionByBucket).
			WithContext(ctx).
			Bind(reaction.MessageId, reaction.BucketKey, reaction.ReactionId).
			Exec(); err != nil {
			return fmt.Errorf("unable to delete bucket reaction: %w", err)
		}
	}
	return nil
}

func (e *Entity) SetSummary(ctx context.Context, summary model.ReactionSummary) error {
	if summary.Count <= 0 {
		if err := e.c.Session().
			Query(deleteSummary).
			WithContext(ctx).
			Bind(summary.MessageId, summary.BucketKey).
			Exec(); err != nil {
			return fmt.Errorf("unable to delete reaction summary: %w", err)
		}
		return nil
	}
	if err := e.c.Session().
		Query(upsertSummary).
		WithContext(ctx).
		Bind(summary.MessageId, summary.BucketKey, summary.Custom, summary.EmojiId, summary.EmojiName, summary.Count).
		Exec(); err != nil {
		return fmt.Errorf("unable to upsert reaction summary: %w", err)
	}
	return nil
}

func scanReactions(iter *gocql.Iter) ([]model.Reaction, error) {
	var reactions []model.Reaction
	var reaction model.Reaction
	for iter.Scan(&reaction.MessageId, &reaction.ReactionId, &reaction.UserId, &reaction.BucketKey, &reaction.Custom, &reaction.EmojiId, &reaction.EmojiName) {
		reactions = append(reactions, reaction)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return reactions, nil
}
