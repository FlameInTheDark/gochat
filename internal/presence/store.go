package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
)

type Store struct {
	c cache.Cache
}

func NewStore(c cache.Cache) *Store { return &Store{c: c} }

func sessionsKey(userID int64) string { return fmt.Sprintf("presence:sessions:%d", userID) }
func aggKey(userID int64) string      { return fmt.Sprintf("presence:agg:%d", userID) }
func overrideKey(userID int64) string { return fmt.Sprintf("presence:override:%d", userID) }

// UpsertSession creates or updates a session presence and refreshes TTLs.
func (s *Store) UpsertSession(ctx context.Context, userID int64, sessionID string, p SessionPresence, ttlSeconds int64) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := s.c.HSet(ctx, sessionsKey(userID), sessionID, string(b)); err != nil {
		return err
	}
	_ = s.c.SetTTL(ctx, sessionsKey(userID), ttlSeconds)
	_ = s.c.SetTTL(ctx, aggKey(userID), ttlSeconds)
	return nil
}

// TouchSessionTTL refreshes only TTLs. If the session exists, extend ExpiresAt too.
func (s *Store) TouchSessionTTL(ctx context.Context, userID int64, sessionID string, ttlSeconds int64) error {
	val, err := s.c.HGet(ctx, sessionsKey(userID), sessionID)
	if err == nil && val != "" {
		var sp SessionPresence
		if json.Unmarshal([]byte(val), &sp) == nil {
			sp.ExpiresAt = time.Now().Unix() + ttlSeconds
			sp.UpdatedAt = time.Now().Unix()
			b, _ := json.Marshal(sp)
			_ = s.c.HSet(ctx, sessionsKey(userID), sessionID, string(b))
		}
	}
	_ = s.c.SetTTL(ctx, sessionsKey(userID), ttlSeconds)
	_ = s.c.SetTTL(ctx, aggKey(userID), ttlSeconds)
	return nil
}

// RemoveSession logically removes session by blanking its field; then refresh TTL.
func (s *Store) RemoveSession(ctx context.Context, userID int64, sessionID string, ttlSeconds int64) error {
	if err := s.c.HDel(ctx, sessionsKey(userID), sessionID); err != nil {
		return err
	}

	if m, err := s.c.HGetAll(ctx, sessionsKey(userID)); err == nil {
		empty := true
		for _, v := range m {
			if v != "" {
				empty = false
				break
			}
		}
		if empty {
			_ = s.c.Delete(ctx, sessionsKey(userID))
		} else {
			_ = s.c.SetTTL(ctx, sessionsKey(userID), ttlSeconds)
		}
	}
	_ = s.c.SetTTL(ctx, aggKey(userID), ttlSeconds)
	return nil
}

// Aggregate reads all valid sessions and returns aggregated presence and if any sessions present.
func (s *Store) Aggregate(ctx context.Context, userID int64, nowUnix int64) (Presence, bool, error) {
	// Check global override first (e.g., manual offline/invisible)
	var ov Presence
	if err := s.c.GetJSON(ctx, overrideKey(userID), &ov); err == nil && ov.Status != "" {
		// Honor override including custom text
		return Presence{UserID: userID, Status: ov.Status, Since: ov.Since, CustomStatusText: ov.CustomStatusText}, true, nil
	}
	m, err := s.c.HGetAll(ctx, sessionsKey(userID))
	if err != nil {
		return Presence{}, false, err
	}
	best := StatusOffline
	since := nowUnix
	any := false
	var bestText string
	var bestTextUpdated int64
	for _, v := range m {
		if v == "" {
			continue
		}
		var sp SessionPresence
		if json.Unmarshal([]byte(v), &sp) != nil {
			continue
		}
		if sp.ExpiresAt <= nowUnix { // expired
			continue
		}
		any = true
		if sp.Status == StatusDND {
			best = StatusDND
		} else if sp.Status == StatusOnline {
			if best != StatusDND {
				best = StatusOnline
			}
		} else if sp.Status == StatusIdle {
			if best != StatusDND && best != StatusOnline {
				best = StatusIdle
			}
		}
		if sp.Since > 0 && sp.Since < since {
			since = sp.Since
		}
		if sp.CustomStatusText != "" && sp.UpdatedAt >= bestTextUpdated {
			bestText = sp.CustomStatusText
			bestTextUpdated = sp.UpdatedAt
		}
	}
	if !any {
		return Presence{UserID: userID, Status: StatusOffline, Since: nowUnix, CustomStatusText: bestText}, false, nil
	}
	return Presence{UserID: userID, Status: best, Since: since, CustomStatusText: bestText}, true, nil
}

// Get returns aggregated presence (from cache if exists; falls back to recompute).
func (s *Store) Get(ctx context.Context, userID int64) (Presence, bool, error) {
	var p Presence
	if err := s.c.GetJSON(ctx, aggKey(userID), &p); err == nil && p.UserID != 0 {
		return p, true, nil
	}
	return s.Aggregate(ctx, userID, time.Now().Unix())
}

// SetAggregated stores aggregated presence with TTL.
func (s *Store) SetAggregated(ctx context.Context, p Presence, ttlSeconds int64) error {
	return s.c.SetTimedJSON(ctx, aggKey(p.UserID), p, ttlSeconds)
}

// Override APIs
func (s *Store) SetOverride(ctx context.Context, userID int64, status string, since int64, text string) error {
	return s.c.SetJSON(ctx, overrideKey(userID), Presence{UserID: userID, Status: status, Since: since, CustomStatusText: text})
}

func (s *Store) ClearOverride(ctx context.Context, userID int64) error {
	return s.c.Delete(ctx, overrideKey(userID))
}

func (s *Store) GetOverride(ctx context.Context, userID int64) (Presence, bool, error) {
	var p Presence
	if err := s.c.GetJSON(ctx, overrideKey(userID), &p); err != nil {
		return Presence{}, false, nil
	}
	return p, true, nil
}
