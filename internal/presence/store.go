package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

type Store struct {
	c cache.Cache
}

func NewStore(c cache.Cache) *Store { return &Store{c: c} }

func sessionsKey(userID int64) string { return fmt.Sprintf("presence:sessions:%d", userID) }
func aggKey(userID int64) string      { return fmt.Sprintf("presence:agg:%d", userID) }
func overrideKey(userID int64) string { return fmt.Sprintf("presence:override:%d", userID) }
func streamKey(userID int64) string   { return fmt.Sprintf("presence:stream:%d", userID) }

// GetSession returns a single session presence record if it exists.
func (s *Store) GetSession(ctx context.Context, userID int64, sessionID string) (SessionPresence, bool, error) {
	val, err := s.c.HGet(ctx, sessionsKey(userID), sessionID)
	if err != nil {
		return SessionPresence{}, false, err
	}
	if val == "" {
		return SessionPresence{}, false, nil
	}

	var sp SessionPresence
	if err := json.Unmarshal([]byte(val), &sp); err != nil {
		return SessionPresence{}, false, err
	}
	return sp, true, nil
}

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

// SetSessionVoiceChannel sets or clears the session's current voice channel ID and refreshes TTLs.
func (s *Store) SetSessionVoiceChannel(ctx context.Context, userID int64, sessionID string, channelID *int64, ttlSeconds int64) error {
	val, err := s.c.HGet(ctx, sessionsKey(userID), sessionID)
	var sp SessionPresence
	if err == nil && val != "" {
		_ = json.Unmarshal([]byte(val), &sp)
	}
	sp.SessionID = sessionID
	sp.UpdatedAt = time.Now().Unix()
	sp.ExpiresAt = time.Now().Unix() + ttlSeconds
	sp.VoiceChannelID = channelID
	b, _ := json.Marshal(sp)
	if err := s.c.HSet(ctx, sessionsKey(userID), sessionID, string(b)); err != nil {
		return err
	}
	_ = s.c.SetTTL(ctx, sessionsKey(userID), ttlSeconds)
	_ = s.c.SetTTL(ctx, aggKey(userID), ttlSeconds)
	return nil
}

// SetSessionVoiceState sets the session's voice state and refreshes TTLs.
func (s *Store) SetSessionVoiceState(ctx context.Context, userID int64, sessionID string, mute, deafen, selfVideo bool, ttlSeconds int64) error {
	val, err := s.c.HGet(ctx, sessionsKey(userID), sessionID)
	var sp SessionPresence
	if err == nil && val != "" {
		_ = json.Unmarshal([]byte(val), &sp)
	}
	sp.SessionID = sessionID
	sp.UpdatedAt = time.Now().Unix()
	sp.ExpiresAt = time.Now().Unix() + ttlSeconds
	sp.Mute = mute
	sp.Deafen = deafen
	sp.SelfVideo = selfVideo
	b, _ := json.Marshal(sp)
	if err := s.c.HSet(ctx, sessionsKey(userID), sessionID, string(b)); err != nil {
		return err
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
	var voiceID *int64
	var voiceIDUpdated int64
	var mute, deafen, selfVideo bool
	var voiceStateUpdated int64
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
		switch sp.Status {
		case StatusDND:
			best = StatusDND
		case StatusOnline:
			if best != StatusDND {
				best = StatusOnline
			}
		case StatusIdle:
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
		if sp.VoiceChannelID != nil && sp.UpdatedAt >= voiceIDUpdated {
			vid := *sp.VoiceChannelID
			voiceID = &vid
			voiceIDUpdated = sp.UpdatedAt
		}
		// Aggregate voice state from the most recently updated session in a voice channel
		if sp.VoiceChannelID != nil && sp.UpdatedAt >= voiceStateUpdated {
			mute = sp.Mute
			deafen = sp.Deafen
			selfVideo = sp.SelfVideo
			voiceStateUpdated = sp.UpdatedAt
		}
	}
	if !any {
		p := Presence{UserID: userID, Status: StatusOffline, Since: nowUnix, CustomStatusText: bestText, VoiceChannelID: voiceID, Mute: mute, Deafen: deafen, SelfVideo: selfVideo}
		if err := s.mergeActiveStream(ctx, &p); err != nil {
			return Presence{}, false, err
		}
		return p, false, nil
	}
	p := Presence{UserID: userID, Status: best, Since: since, CustomStatusText: bestText, VoiceChannelID: voiceID, Mute: mute, Deafen: deafen, SelfVideo: selfVideo}
	if err := s.mergeActiveStream(ctx, &p); err != nil {
		return Presence{}, false, err
	}
	return p, true, nil
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

func (s *Store) SetActiveStream(ctx context.Context, userID int64, stream streammeta.ActiveStream, ttlSeconds int64) error {
	return s.c.SetTimedJSON(ctx, streamKey(userID), stream, ttlSeconds)
}

func (s *Store) ClearActiveStream(ctx context.Context, userID int64) error {
	return s.c.Delete(ctx, streamKey(userID))
}

func (s *Store) GetActiveStream(ctx context.Context, userID int64) (*streammeta.ActiveStream, bool, error) {
	var stream streammeta.ActiveStream
	if err := s.c.GetJSON(ctx, streamKey(userID), &stream); err != nil {
		return nil, false, nil
	}
	if stream.ID == 0 {
		return nil, false, nil
	}
	return &stream, true, nil
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

func (s *Store) mergeActiveStream(ctx context.Context, p *Presence) error {
	if p == nil || p.UserID == 0 || p.VoiceChannelID == nil {
		if p != nil {
			p.ActiveStream = nil
		}
		return nil
	}

	stream, ok, err := s.GetActiveStream(ctx, p.UserID)
	if err != nil {
		return err
	}
	if !ok || stream == nil || stream.ChannelID != *p.VoiceChannelID {
		p.ActiveStream = nil
		return nil
	}

	p.ActiveStream = stream
	return nil
}
