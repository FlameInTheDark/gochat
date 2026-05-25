package gatewaystate

import (
	"context"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
)

const defaultClientStateTTL = int64(24 * time.Hour / time.Second)

type Store struct {
	c cache.Cache
}

type ConnectionState struct {
	ConnectionID     string `json:"connection_id"`
	UserID           int64  `json:"user_id"`
	SessionID        string `json:"session_id"`
	ClientInstanceID string `json:"client_instance_id,omitempty"`
	Generation       int64  `json:"generation"`
	InstanceID       string `json:"instance_id,omitempty"`
	ProtocolVersion  int    `json:"protocol_version"`
	LastHeartbeat    int64  `json:"last_heartbeat"`
	LastSeq          int64  `json:"last_seq"`
	ExpiresAt        int64  `json:"expires_at"`
}

type ClientState struct {
	Channels      []int64 `json:"channels,omitempty"`
	PresenceSet   []int64 `json:"presence_set,omitempty"`
	ActiveContext string  `json:"active_context,omitempty"`
	UpdatedAt     int64   `json:"updated_at"`
}

func NewStore(c cache.Cache) *Store {
	return &Store{c: c}
}

func (s *Store) UpsertConnection(ctx context.Context, st ConnectionState, ttlSeconds int64) error {
	if s == nil || s.c == nil || st.ConnectionID == "" || st.UserID == 0 {
		return nil
	}
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}
	now := time.Now().Unix()
	if st.LastHeartbeat == 0 {
		st.LastHeartbeat = now
	}
	st.ExpiresAt = now + ttlSeconds
	if err := s.c.SetTimedJSON(ctx, connKey(st.ConnectionID), st, ttlSeconds); err != nil {
		return err
	}
	if err := s.c.ZAdd(ctx, userConnsKey(st.UserID), float64(st.ExpiresAt), st.ConnectionID); err != nil {
		return err
	}
	return s.c.SetTTL(ctx, userConnsKey(st.UserID), ttlSeconds)
}

func (s *Store) TouchConnection(ctx context.Context, connectionID string, lastSeq int64, ttlSeconds int64) error {
	if s == nil || s.c == nil || connectionID == "" {
		return nil
	}
	var st ConnectionState
	if err := s.c.GetJSON(ctx, connKey(connectionID), &st); err != nil || st.ConnectionID == "" {
		return nil
	}
	st.LastHeartbeat = time.Now().Unix()
	st.LastSeq = lastSeq
	return s.UpsertConnection(ctx, st, ttlSeconds)
}

func (s *Store) DeleteConnection(ctx context.Context, connectionID string, userID int64) error {
	if s == nil || s.c == nil || connectionID == "" {
		return nil
	}
	if userID != 0 {
		_ = s.c.ZRem(ctx, userConnsKey(userID), connectionID)
	}
	return s.c.Delete(ctx, connKey(connectionID))
}

func (s *Store) SetClientState(ctx context.Context, userID int64, clientInstanceID string, st ClientState, ttlSeconds int64) error {
	if s == nil || s.c == nil || userID == 0 || clientInstanceID == "" {
		return nil
	}
	if ttlSeconds < 1 {
		ttlSeconds = defaultClientStateTTL
	}
	st.UpdatedAt = time.Now().Unix()
	return s.c.SetTimedJSON(ctx, clientStateKey(userID, clientInstanceID), st, ttlSeconds)
}

func (s *Store) GetClientState(ctx context.Context, userID int64, clientInstanceID string) (ClientState, bool, error) {
	if s == nil || s.c == nil || userID == 0 || clientInstanceID == "" {
		return ClientState{}, false, nil
	}
	var st ClientState
	if err := s.c.GetJSON(ctx, clientStateKey(userID, clientInstanceID), &st); err != nil {
		return ClientState{}, false, nil
	}
	return st, true, nil
}

func connKey(connectionID string) string {
	return fmt.Sprintf("ws:conn:%s", connectionID)
}

func userConnsKey(userID int64) string {
	return fmt.Sprintf("ws:user_conns:%d", userID)
}

func clientStateKey(userID int64, clientInstanceID string) string {
	return fmt.Sprintf("ws:client:%d:%s:state", userID, clientInstanceID)
}
