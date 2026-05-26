package botgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/redis/go-redis/v9"
)

type Session struct {
	SessionID        string `json:"session_id"`
	BotUserID        int64  `json:"bot_user_id"`
	InstanceID       string `json:"instance_id"`
	ShardID          int    `json:"shard_id"`
	ShardCount       int    `json:"shard_count"`
	ReceivesDMEvents bool   `json:"receives_dm_events"`
}

type Registry struct {
	cache *kvs.Cache
	ttl   time.Duration
}

func NewRegistry(cache *kvs.Cache, ttl time.Duration) *Registry {
	if ttl <= 0 {
		ttl = 75 * time.Second
	}
	return &Registry{cache: cache, ttl: ttl}
}

func (r *Registry) Register(ctx context.Context, session Session) error {
	raw, err := json.Marshal(session)
	if err != nil {
		return err
	}
	pipe := r.cache.Client().Pipeline()
	pipe.Set(ctx, sessionKey(session.SessionID), raw, r.ttl)
	pipe.SAdd(ctx, botSessionsKey(session.BotUserID), session.SessionID)
	pipe.SAdd(ctx, instanceSessionsKey(session.InstanceID), session.SessionID)
	pipe.Set(ctx, botShardCountKey(session.BotUserID), strconv.Itoa(session.ShardCount), r.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *Registry) Touch(ctx context.Context, session Session) error {
	pipe := r.cache.Client().Pipeline()
	pipe.Expire(ctx, sessionKey(session.SessionID), r.ttl)
	pipe.Expire(ctx, botShardCountKey(session.BotUserID), r.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Registry) Unregister(ctx context.Context, session Session) error {
	pipe := r.cache.Client().Pipeline()
	pipe.Del(ctx, sessionKey(session.SessionID))
	pipe.SRem(ctx, botSessionsKey(session.BotUserID), session.SessionID)
	pipe.SRem(ctx, instanceSessionsKey(session.InstanceID), session.SessionID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Registry) ExistingShardCount(ctx context.Context, botUserID int64) (int, bool, error) {
	raw, err := r.cache.Client().Get(ctx, botShardCountKey(botUserID)).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	count, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, err
	}
	return count, true, nil
}

func (r *Registry) BotSessions(ctx context.Context, botUserID int64) ([]Session, error) {
	ids, err := r.cache.Client().SMembers(ctx, botSessionsKey(botUserID)).Result()
	if err != nil {
		return nil, err
	}
	return r.sessionsByID(ctx, ids, botUserID, "")
}

func (r *Registry) sessionsByID(ctx context.Context, ids []string, botUserID int64, instanceID string) ([]Session, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, sessionKey(id))
	}
	values, err := r.cache.Client().MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(values))
	stale := make([]string, 0)
	for i, value := range values {
		if value == nil {
			stale = append(stale, ids[i])
			continue
		}
		raw, ok := value.(string)
		if !ok {
			stale = append(stale, ids[i])
			continue
		}
		var session Session
		if err := json.Unmarshal([]byte(raw), &session); err != nil {
			stale = append(stale, ids[i])
			continue
		}
		sessions = append(sessions, session)
	}
	if len(stale) > 0 {
		pipe := r.cache.Client().Pipeline()
		if botUserID != 0 {
			args := make([]interface{}, 0, len(stale))
			for _, id := range stale {
				args = append(args, id)
			}
			pipe.SRem(ctx, botSessionsKey(botUserID), args...)
		}
		if instanceID != "" {
			args := make([]interface{}, 0, len(stale))
			for _, id := range stale {
				args = append(args, id)
			}
			pipe.SRem(ctx, instanceSessionsKey(instanceID), args...)
		}
		_, _ = pipe.Exec(ctx)
	}
	return sessions, nil
}

func TargetGuildSessions(sessions []Session, guildID int64) []Session {
	targets := make([]Session, 0, len(sessions))
	for _, session := range sessions {
		if session.ShardCount <= 1 {
			targets = append(targets, session)
			continue
		}
		if int(guildID%int64(session.ShardCount)) == session.ShardID {
			targets = append(targets, session)
		}
	}
	return targets
}

func TargetDMSessions(sessions []Session) []Session {
	targets := make([]Session, 0, len(sessions))
	for _, session := range sessions {
		if !session.ReceivesDMEvents {
			continue
		}
		if session.ShardCount <= 1 || session.ShardID == 0 {
			targets = append(targets, session)
		}
	}
	return targets
}

func GroupByInstance(sessions []Session) map[string][]string {
	grouped := make(map[string][]string)
	for _, session := range sessions {
		if session.InstanceID == "" || session.SessionID == "" {
			continue
		}
		grouped[session.InstanceID] = append(grouped[session.InstanceID], session.SessionID)
	}
	return grouped
}

func sessionKey(sessionID string) string {
	return "botgw:session:" + sessionID
}

func botSessionsKey(botUserID int64) string {
	return fmt.Sprintf("botgw:sessions:bot:%d", botUserID)
}

func instanceSessionsKey(instanceID string) string {
	return "botgw:sessions:instance:" + instanceID
}

func botShardCountKey(botUserID int64) string {
	return fmt.Sprintf("botgw:shard_count:%d", botUserID)
}
