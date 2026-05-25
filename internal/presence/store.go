package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

type Store struct {
	c cache.Cache
}

func NewStore(c cache.Cache) *Store { return &Store{c: c} }

func legacySessionsKey(userID int64) string { return fmt.Sprintf("presence:sessions:%d", userID) }
func sessionsKey(userID int64) string       { return fmt.Sprintf("presence:sessions:v2:%d", userID) }
func sessionKey(userID int64, leaseID string) string {
	return fmt.Sprintf("presence:session:%d:%s", userID, leaseID)
}
func aggKey(userID int64) string      { return fmt.Sprintf("presence:agg:%d", userID) }
func overrideKey(userID int64) string { return fmt.Sprintf("presence:override:%d", userID) }
func streamKey(userID int64) string   { return fmt.Sprintf("presence:stream:%d", userID) }
func touchedUsersKey() string         { return "presence:touched_users" }

func leaseID(sessionID string, generation int64) string {
	if generation <= 0 {
		return sessionID
	}
	return fmt.Sprintf("%s:%d", sessionID, generation)
}

func sessionIDFromLeaseID(id string) string {
	if idx := strings.LastIndex(id, ":"); idx > 0 {
		if _, err := strconv.ParseInt(id[idx+1:], 10, 64); err == nil {
			return id[:idx]
		}
	}
	return id
}

func (s *Store) liveSessions(ctx context.Context, userID int64, nowUnix int64) ([]SessionPresence, error) {
	members, err := s.c.ZRevRangeByScore(ctx, sessionsKey(userID), "+inf", fmt.Sprintf("(%d", nowUnix), 0, 1000)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(members))
	for _, member := range members {
		if member == "" {
			continue
		}
		keys = append(keys, sessionKey(userID, member))
	}

	raw, err := s.c.MGetBytes(ctx, keys...)
	if err != nil {
		return nil, err
	}

	sessions := make([]SessionPresence, 0, len(raw))
	for i, body := range raw {
		if len(body) == 0 {
			if i < len(members) {
				_ = s.c.ZRem(ctx, sessionsKey(userID), members[i])
			}
			continue
		}
		var sp SessionPresence
		if err := json.Unmarshal(body, &sp); err != nil {
			continue
		}
		if sp.ExpiresAt <= nowUnix {
			if i < len(members) {
				_ = s.c.ZRem(ctx, sessionsKey(userID), members[i])
				_ = s.c.Delete(ctx, sessionKey(userID, members[i]))
			}
			continue
		}
		sessions = append(sessions, sp)
	}

	legacy, err := s.legacyLiveSessions(ctx, userID, nowUnix)
	if err != nil {
		return nil, err
	}
	sessions = append(sessions, legacy...)
	return sessions, nil
}

func (s *Store) legacyLiveSessions(ctx context.Context, userID int64, nowUnix int64) ([]SessionPresence, error) {
	m, err := s.c.HGetAll(ctx, legacySessionsKey(userID))
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}
	sessions := make([]SessionPresence, 0, len(m))
	for _, v := range m {
		if v == "" {
			continue
		}
		var sp SessionPresence
		if json.Unmarshal([]byte(v), &sp) != nil || sp.ExpiresAt <= nowUnix {
			continue
		}
		sessions = append(sessions, sp)
	}
	return sessions, nil
}

func (s *Store) getSessionByLease(ctx context.Context, userID int64, sessionID string, generation int64) (SessionPresence, bool, error) {
	var sp SessionPresence
	if sessionID == "" {
		return sp, false, nil
	}
	if err := s.c.GetJSON(ctx, sessionKey(userID, leaseID(sessionID, generation)), &sp); err == nil && sp.SessionID != "" {
		return sp, true, nil
	}
	return SessionPresence{}, false, nil
}

// GetSession returns the newest live lease for a logical session id.
func (s *Store) GetSession(ctx context.Context, userID int64, sessionID string) (SessionPresence, bool, error) {
	now := time.Now().Unix()
	sessions, err := s.liveSessions(ctx, userID, now)
	if err != nil {
		return SessionPresence{}, false, err
	}
	var best SessionPresence
	var found bool
	for _, sp := range sessions {
		if sp.SessionID != sessionID {
			continue
		}
		if !found || sp.Generation > best.Generation || sp.UpdatedAt > best.UpdatedAt {
			best = sp
			found = true
		}
	}
	if found {
		return best, true, nil
	}

	val, err := s.c.HGet(ctx, legacySessionsKey(userID), sessionID)
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

// UpsertSession creates or updates a fenced per-session presence lease.
func (s *Store) UpsertSession(ctx context.Context, userID int64, sessionID string, p SessionPresence, ttlSeconds int64) error {
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}
	if p.SessionID == "" {
		p.SessionID = sessionID
	}
	member := leaseID(p.SessionID, p.Generation)
	if err := s.c.SetTimedJSON(ctx, sessionKey(userID, member), p, ttlSeconds); err != nil {
		return err
	}
	if err := s.c.ZAdd(ctx, sessionsKey(userID), float64(p.ExpiresAt), member); err != nil {
		return err
	}
	_ = s.markTouched(ctx, userID)
	return s.c.SetTTL(ctx, sessionsKey(userID), ttlSeconds)
}

func (s *Store) TouchSessionTTL(ctx context.Context, userID int64, sessionID string, ttlSeconds int64) error {
	sp, ok, err := s.GetSession(ctx, userID, sessionID)
	if err != nil || !ok {
		return err
	}
	return s.touchSession(ctx, userID, sp, ttlSeconds)
}

func (s *Store) TouchSessionTTLIfOwner(ctx context.Context, userID int64, sessionID, connectionID string, generation, ttlSeconds int64) error {
	sp, ok, err := s.getSessionByLease(ctx, userID, sessionID, generation)
	if err != nil || !ok {
		return err
	}
	if sp.ConnectionID != connectionID || sp.Generation != generation {
		return nil
	}
	return s.touchSession(ctx, userID, sp, ttlSeconds)
}

func (s *Store) touchSession(ctx context.Context, userID int64, sp SessionPresence, ttlSeconds int64) error {
	now := time.Now().Unix()
	sp.ExpiresAt = now + ttlSeconds
	sp.UpdatedAt = now
	return s.UpsertSession(ctx, userID, sp.SessionID, sp, ttlSeconds)
}

func (s *Store) SetSessionVoiceChannel(ctx context.Context, userID int64, sessionID string, channelID *int64, ttlSeconds int64) error {
	sp, _, _ := s.GetSession(ctx, userID, sessionID)
	sp.SessionID = sessionID
	sp.UpdatedAt = time.Now().Unix()
	sp.ExpiresAt = time.Now().Unix() + ttlSeconds
	sp.VoiceChannelID = channelID
	return s.UpsertSession(ctx, userID, sessionID, sp, ttlSeconds)
}

func (s *Store) SetSessionVoiceChannelIfOwner(ctx context.Context, userID int64, sessionID, connectionID string, generation int64, channelID *int64, ttlSeconds int64) error {
	sp, ok, err := s.getSessionByLease(ctx, userID, sessionID, generation)
	if err != nil || !ok {
		return err
	}
	if sp.ConnectionID != connectionID || sp.Generation != generation {
		return nil
	}
	sp.UpdatedAt = time.Now().Unix()
	sp.ExpiresAt = time.Now().Unix() + ttlSeconds
	sp.VoiceChannelID = channelID
	return s.UpsertSession(ctx, userID, sessionID, sp, ttlSeconds)
}

func (s *Store) SetSessionVoiceState(ctx context.Context, userID int64, sessionID string, mute, deafen, selfVideo bool, ttlSeconds int64) error {
	sp, _, _ := s.GetSession(ctx, userID, sessionID)
	sp.SessionID = sessionID
	sp.UpdatedAt = time.Now().Unix()
	sp.ExpiresAt = time.Now().Unix() + ttlSeconds
	sp.Mute = mute
	sp.Deafen = deafen
	sp.SelfVideo = selfVideo
	return s.UpsertSession(ctx, userID, sessionID, sp, ttlSeconds)
}

func (s *Store) RemoveSession(ctx context.Context, userID int64, sessionID string, ttlSeconds int64) error {
	now := time.Now().Unix()
	sessions, err := s.liveSessions(ctx, userID, now)
	if err != nil {
		return err
	}
	for _, sp := range sessions {
		if sp.SessionID != sessionID {
			continue
		}
		member := leaseID(sp.SessionID, sp.Generation)
		_ = s.c.ZRem(ctx, sessionsKey(userID), member)
		_ = s.c.Delete(ctx, sessionKey(userID, member))
	}
	_ = s.c.HDel(ctx, legacySessionsKey(userID), sessionID)
	_ = s.markTouched(ctx, userID)
	_ = s.c.SetTTL(ctx, aggKey(userID), ttlSeconds)
	return nil
}

func (s *Store) RemoveSessionIfOwner(ctx context.Context, userID int64, sessionID, connectionID string, generation, ttlSeconds int64) (bool, error) {
	sp, ok, err := s.getSessionByLease(ctx, userID, sessionID, generation)
	if err != nil || !ok {
		return false, err
	}
	if sp.ConnectionID != connectionID || sp.Generation != generation {
		return false, nil
	}
	member := leaseID(sp.SessionID, sp.Generation)
	if err := s.c.ZRem(ctx, sessionsKey(userID), member); err != nil {
		return false, err
	}
	if err := s.c.Delete(ctx, sessionKey(userID, member)); err != nil {
		return false, err
	}
	_ = s.markTouched(ctx, userID)
	_ = s.c.SetTTL(ctx, aggKey(userID), ttlSeconds)
	return true, nil
}

func (s *Store) Aggregate(ctx context.Context, userID int64, nowUnix int64) (Presence, bool, error) {
	var ov Presence
	if err := s.c.GetJSON(ctx, overrideKey(userID), &ov); err == nil && ov.Status != "" {
		return Presence{UserID: userID, Status: ov.Status, Since: ov.Since, CustomStatusText: ov.CustomStatusText}, true, nil
	}

	sessions, err := s.liveSessions(ctx, userID, nowUnix)
	if err != nil {
		return Presence{}, false, err
	}
	best := StatusOffline
	since := nowUnix
	any := false
	clientStatus := make(map[string]string)
	var bestText string
	var bestTextUpdated int64
	var voiceID *int64
	var voiceIDUpdated int64
	var mute, deafen, selfVideo bool
	var voiceStateUpdated int64

	for _, sp := range sessions {
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
		if sp.Platform != "" {
			if current, ok := clientStatus[sp.Platform]; !ok || statusRank(sp.Status) > statusRank(current) {
				clientStatus[sp.Platform] = sp.Status
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
		if sp.VoiceChannelID != nil && sp.UpdatedAt >= voiceStateUpdated {
			mute = sp.Mute
			deafen = sp.Deafen
			selfVideo = sp.SelfVideo
			voiceStateUpdated = sp.UpdatedAt
		}
	}
	if !any {
		p := Presence{UserID: userID, Status: StatusOffline, Since: nowUnix, CustomStatusText: bestText}
		if err := s.mergeActiveStream(ctx, &p); err != nil {
			return Presence{}, false, err
		}
		return p, false, nil
	}
	p := Presence{UserID: userID, Status: best, Since: since, CustomStatusText: bestText, ClientStatus: clientStatus, VoiceChannelID: voiceID, Mute: mute, Deafen: deafen, SelfVideo: selfVideo}
	if err := s.mergeActiveStream(ctx, &p); err != nil {
		return Presence{}, false, err
	}
	return p, true, nil
}

func statusRank(status string) int {
	switch status {
	case StatusDND:
		return 3
	case StatusOnline:
		return 2
	case StatusIdle:
		return 1
	default:
		return 0
	}
}

func (s *Store) Get(ctx context.Context, userID int64) (Presence, bool, error) {
	var p Presence
	if err := s.c.GetJSON(ctx, aggKey(userID), &p); err == nil && p.UserID != 0 {
		return p, true, nil
	}
	return s.Aggregate(ctx, userID, time.Now().Unix())
}

func (s *Store) SetAggregated(ctx context.Context, p Presence, ttlSeconds int64) error {
	return s.c.SetTimedJSON(ctx, aggKey(p.UserID), p, ttlSeconds)
}

func (s *Store) SetActiveStream(ctx context.Context, userID int64, stream streammeta.ActiveStream, ttlSeconds int64) error {
	_ = s.markTouched(ctx, userID)
	return s.c.SetTimedJSON(ctx, streamKey(userID), stream, ttlSeconds)
}

func (s *Store) ClearActiveStream(ctx context.Context, userID int64) error {
	_ = s.markTouched(ctx, userID)
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

func (s *Store) SetOverride(ctx context.Context, userID int64, status string, since int64, text string) error {
	_ = s.markTouched(ctx, userID)
	return s.c.SetJSON(ctx, overrideKey(userID), Presence{UserID: userID, Status: status, Since: since, CustomStatusText: text})
}

func (s *Store) ClearOverride(ctx context.Context, userID int64) error {
	_ = s.markTouched(ctx, userID)
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

func (s *Store) RecentlyTouched(ctx context.Context, limit int64) ([]int64, error) {
	if limit <= 0 {
		limit = 1000
	}
	members, err := s.c.ZRevRangeByScore(ctx, touchedUsersKey(), "+inf", "-inf", 0, limit)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseInt(member, 10, 64)
		if err != nil || userID == 0 {
			continue
		}
		out = append(out, userID)
	}
	return out, nil
}

func (s *Store) CachedAggregate(ctx context.Context, userID int64) (Presence, bool, error) {
	var p Presence
	if userID == 0 {
		return Presence{}, false, nil
	}
	if err := s.c.GetJSON(ctx, aggKey(userID), &p); err != nil || p.UserID == 0 {
		return Presence{}, false, nil
	}
	return p, true, nil
}

func (s *Store) PruneExpiredSessions(ctx context.Context, userID int64, nowUnix, limit int64) (int, error) {
	if userID == 0 {
		return 0, nil
	}
	if limit <= 0 {
		limit = 1000
	}
	members, err := s.c.ZRevRangeByScore(ctx, sessionsKey(userID), fmt.Sprintf("%d", nowUnix), "-inf", 0, limit)
	if err != nil {
		return 0, err
	}
	for _, member := range members {
		if member == "" {
			continue
		}
		_ = s.c.Delete(ctx, sessionKey(userID, member))
		_ = s.c.ZRem(ctx, sessionsKey(userID), member)
	}
	if len(members) > 0 {
		_ = s.markTouched(ctx, userID)
	}
	return len(members), nil
}

func (s *Store) markTouched(ctx context.Context, userID int64) error {
	if userID == 0 {
		return nil
	}
	if err := s.c.ZAdd(ctx, touchedUsersKey(), float64(time.Now().Unix()), strconv.FormatInt(userID, 10)); err != nil {
		return err
	}
	return s.c.SetTTL(ctx, touchedUsersKey(), int64((24*time.Hour)/time.Second))
}

func LeaseIDForTest(sessionID string, generation int64) string {
	return leaseID(sessionID, generation)
}

func SessionIDFromLeaseIDForTest(id string) string {
	return sessionIDFromLeaseID(id)
}
