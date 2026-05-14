package guild

import (
	"context"
	"encoding/binary"
	"hash/fnv"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

const (
	discoverySnapshotTTL        = 2 * time.Second
	discoverySnapshotStaleTTL   = 15 * time.Second
	discoveryRegionsCacheTTL    = 30 * time.Second
	selectionReservationTTL     = 10 * time.Second
	selectionReservationPenalty = int64(8)
	selectionSpreadWindow       = int64(2)
	voiceRouteInitialTTLSeconds = int64(180)
)

type discoverySnapshot struct {
	instances []discovery.Instance
	fetchedAt time.Time
	expiresAt time.Time
}

type discoveredRegionsSnapshot struct {
	regions   []string
	fetchedAt time.Time
	expiresAt time.Time
}

type voiceSelector struct {
	mu                sync.Mutex
	regionSnapshots   map[string]discoverySnapshot
	regionRefreshWait map[string]chan struct{}
	knownRegions      discoveredRegionsSnapshot
	knownRegionsWait  chan struct{}
	reservations      map[string][]time.Time
}

func newVoiceSelector(log *slog.Logger) *voiceSelector {
	return &voiceSelector{
		regionSnapshots:   make(map[string]discoverySnapshot),
		regionRefreshWait: make(map[string]chan struct{}),
		reservations:      make(map[string][]time.Time),
	}
}

func (e *entity) preferredVoiceRegion(ctx context.Context, channelID int64) string {
	return e.preferredVoiceRegionForUser(ctx, channelID, 0)
}

func (e *entity) preferredVoiceRegionForUser(ctx context.Context, channelID, userID int64) string {
	var region string
	if e != nil {
		region = e.defaultVoiceRegion
	}
	if e == nil || e.ch == nil {
		return region
	}
	if dbreg, err := e.ch.GetChannelVoiceRegion(ctx, channelID); err == nil && dbreg != nil && *dbreg != "" {
		region = *dbreg
		return region
	}
	if userID != 0 && e.uset != nil {
		settings, err := e.uset.GetUserSettings(ctx, userID, 0)
		if err == nil {
			data, err := model.UnmarshalStoredUserSettingsData(settings.Settings)
			if err == nil {
				preferred := strings.TrimSpace(data.Voice.PreferredRegion)
				if preferred != "" && preferred != "auto" {
					if len(e.allowedRegions) == 0 {
						return preferred
					}
					if _, ok := e.allowedRegions[preferred]; ok {
						return preferred
					}
				}
			}
		}
	}
	return region
}

func (e *entity) cachedChannelBinding(ctx context.Context, channelID int64) voiceRouteBinding {
	var binding voiceRouteBinding
	if e == nil || e.cache == nil {
		return binding
	}
	_ = e.cache.GetJSON(ctx, bindingKey(channelID), &binding)
	return binding
}

func (e *entity) bindChannelRoute(ctx context.Context, channelID int64, binding voiceRouteBinding) voiceRouteBinding {
	if e == nil || e.cache == nil || binding.URL == "" {
		return binding
	}

	set, err := e.cache.SetTimedJSONNX(ctx, bindingKey(channelID), binding, voiceRouteInitialTTLSeconds, cache.NoneProactive())
	if err == nil && set {
		return binding
	}

	if winner := e.cachedChannelBinding(ctx, channelID); winner.URL != "" {
		return winner
	}

	return binding
}

func (e *entity) channelBindingForJoin(ctx context.Context, channelID int64) (voiceRouteBinding, error) {
	return e.channelBindingForJoinByUser(ctx, channelID, 0)
}

func (e *entity) channelBindingForJoinByUser(ctx context.Context, channelID, userID int64) (voiceRouteBinding, error) {
	if binding := e.cachedChannelBinding(ctx, channelID); binding.URL != "" {
		return binding, nil
	}

	binding, err := e.selectSFUBinding(ctx, channelID, e.preferredVoiceRegionForUser(ctx, channelID, userID), true)
	if err != nil {
		return voiceRouteBinding{}, err
	}
	if binding.URL == "" {
		return voiceRouteBinding{}, nil
	}

	return e.bindChannelRoute(ctx, channelID, binding), nil
}

func (e *entity) selectSFUBinding(ctx context.Context, channelID int64, preferredRegion string, allowFallback bool) (voiceRouteBinding, error) {
	if e == nil || e.disco == nil {
		return voiceRouteBinding{}, nil
	}

	regions, err := e.selectionRegions(ctx, preferredRegion, allowFallback)
	if err != nil {
		return voiceRouteBinding{}, err
	}

	var firstErr error
	for _, region := range regions {
		instances, err := e.cachedRegionInstances(ctx, region)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if len(instances) == 0 {
			continue
		}
		if binding, ok := e.voiceSelector.pickBinding(channelID, region, instances); ok {
			return binding, nil
		}
	}

	return voiceRouteBinding{}, firstErr
}

func (e *entity) selectionRegions(ctx context.Context, preferredRegion string, allowFallback bool) ([]string, error) {
	seen := make(map[string]struct{}, len(e.allowedRegionIDs)+1)
	regions := make([]string, 0, len(e.allowedRegionIDs)+1)
	add := func(region string) {
		if region == "" {
			return
		}
		if _, ok := seen[region]; ok {
			return
		}
		seen[region] = struct{}{}
		regions = append(regions, region)
	}

	add(preferredRegion)
	if !allowFallback {
		return regions, nil
	}

	if len(e.allowedRegionIDs) > 0 {
		for _, region := range e.allowedRegionIDs {
			add(region)
		}
		return regions, nil
	}

	discovered, err := e.cachedDiscoveryRegions(ctx)
	if err != nil {
		if len(regions) > 0 {
			return regions, nil
		}
		return nil, err
	}
	for _, region := range discovered {
		add(region)
	}

	return regions, nil
}

func (e *entity) cachedRegionInstances(ctx context.Context, region string) ([]discovery.Instance, error) {
	if e == nil || e.disco == nil || e.voiceSelector == nil || region == "" {
		return nil, nil
	}

	for attempts := 0; attempts < 2; attempts++ {
		if instances, ok := e.voiceSelector.regionSnapshot(region, time.Now()); ok {
			return instances, nil
		}

		waitCh, refresh := e.voiceSelector.beginRegionRefresh(region)
		if !refresh {
			<-waitCh
			continue
		}

		instances, err := e.disco.List(ctx, region)
		if err != nil {
			e.voiceSelector.finishRegionRefresh(region, nil, false)
			if stale, ok := e.voiceSelector.staleRegionSnapshot(region, time.Now()); ok {
				if e.log != nil {
					e.log.Warn("voice discovery using stale region snapshot",
						slog.String("region", region),
						slog.String("error", err.Error()))
				}
				return stale, nil
			}
			return nil, err
		}

		e.voiceSelector.finishRegionRefresh(region, instances, true)
		return cloneDiscoveryInstances(instances), nil
	}

	return nil, nil
}

func (e *entity) cachedDiscoveryRegions(ctx context.Context) ([]string, error) {
	if e == nil || e.disco == nil || e.voiceSelector == nil {
		return nil, nil
	}

	for attempts := 0; attempts < 2; attempts++ {
		if regions, ok := e.voiceSelector.discoveryRegions(time.Now()); ok {
			return regions, nil
		}

		waitCh, refresh := e.voiceSelector.beginRegionsRefresh()
		if !refresh {
			<-waitCh
			continue
		}

		regions, err := e.disco.Regions(ctx)
		if err != nil {
			e.voiceSelector.finishRegionsRefresh(nil, false)
			if stale, ok := e.voiceSelector.staleDiscoveryRegions(time.Now()); ok {
				if e.log != nil {
					e.log.Warn("voice discovery using stale region list", slog.String("error", err.Error()))
				}
				return stale, nil
			}
			return nil, err
		}

		sort.Strings(regions)
		e.voiceSelector.finishRegionsRefresh(regions, true)
		return append([]string(nil), regions...), nil
	}

	return nil, nil
}

func (s *voiceSelector) regionSnapshot(region string, now time.Time) ([]discovery.Instance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, ok := s.regionSnapshots[region]
	if !ok || now.After(snapshot.expiresAt) {
		return nil, false
	}

	return cloneDiscoveryInstances(snapshot.instances), true
}

func (s *voiceSelector) staleRegionSnapshot(region string, now time.Time) ([]discovery.Instance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, ok := s.regionSnapshots[region]
	if !ok || now.Sub(snapshot.fetchedAt) > discoverySnapshotStaleTTL {
		return nil, false
	}

	return cloneDiscoveryInstances(snapshot.instances), true
}

func (s *voiceSelector) beginRegionRefresh(region string) (<-chan struct{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if waitCh, ok := s.regionRefreshWait[region]; ok {
		return waitCh, false
	}

	waitCh := make(chan struct{})
	s.regionRefreshWait[region] = waitCh
	return waitCh, true
}

func (s *voiceSelector) finishRegionRefresh(region string, instances []discovery.Instance, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ok {
		now := time.Now()
		s.regionSnapshots[region] = discoverySnapshot{
			instances: cloneDiscoveryInstances(instances),
			fetchedAt: now,
			expiresAt: now.Add(discoverySnapshotTTL),
		}
	}

	if waitCh, exists := s.regionRefreshWait[region]; exists {
		close(waitCh)
		delete(s.regionRefreshWait, region)
	}
}

func (s *voiceSelector) discoveryRegions(now time.Time) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.knownRegions.regions) == 0 || now.After(s.knownRegions.expiresAt) {
		return nil, false
	}

	return append([]string(nil), s.knownRegions.regions...), true
}

func (s *voiceSelector) staleDiscoveryRegions(now time.Time) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.knownRegions.regions) == 0 || now.Sub(s.knownRegions.fetchedAt) > discoverySnapshotStaleTTL {
		return nil, false
	}

	return append([]string(nil), s.knownRegions.regions...), true
}

func (s *voiceSelector) beginRegionsRefresh() (<-chan struct{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.knownRegionsWait != nil {
		return s.knownRegionsWait, false
	}

	s.knownRegionsWait = make(chan struct{})
	return s.knownRegionsWait, true
}

func (s *voiceSelector) finishRegionsRefresh(regions []string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ok {
		now := time.Now()
		s.knownRegions = discoveredRegionsSnapshot{
			regions:   append([]string(nil), regions...),
			fetchedAt: now,
			expiresAt: now.Add(discoveryRegionsCacheTTL),
		}
	}

	if s.knownRegionsWait != nil {
		close(s.knownRegionsWait)
		s.knownRegionsWait = nil
	}
}

func (s *voiceSelector) pickBinding(channelID int64, region string, instances []discovery.Instance) (voiceRouteBinding, bool) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	effective := make(map[string]int64, len(instances))
	var best int64
	bestSet := false
	for _, inst := range instances {
		if inst.ID == "" || inst.URL == "" {
			continue
		}
		pending := s.pendingReservationsLocked(inst.ID, now)
		score := inst.Load + int64(pending)*selectionReservationPenalty
		effective[inst.ID] = score
		if !bestSet || score < best {
			best = score
			bestSet = true
		}
	}
	if !bestSet {
		return voiceRouteBinding{}, false
	}

	var (
		chosen    discovery.Instance
		chosenSet bool
		bestHash  uint64
	)
	for _, inst := range instances {
		score, ok := effective[inst.ID]
		if !ok || score > best+selectionSpreadWindow {
			continue
		}
		hash := rendezvousScore(channelID, inst.ID)
		if !chosenSet || hash > bestHash {
			chosen = inst
			chosenSet = true
			bestHash = hash
		}
	}
	if !chosenSet {
		return voiceRouteBinding{}, false
	}

	s.reservations[chosen.ID] = append(s.reservations[chosen.ID], now.Add(selectionReservationTTL))
	return voiceRouteBinding{ID: chosen.ID, URL: chosen.URL, Region: region}, true
}

func (s *voiceSelector) pendingReservationsLocked(instanceID string, now time.Time) int {
	expiries := s.reservations[instanceID]
	if len(expiries) == 0 {
		return 0
	}

	alive := expiries[:0]
	for _, expiry := range expiries {
		if expiry.After(now) {
			alive = append(alive, expiry)
		}
	}
	if len(alive) == 0 {
		delete(s.reservations, instanceID)
		return 0
	}

	s.reservations[instanceID] = alive
	return len(alive)
}

func cloneDiscoveryInstances(instances []discovery.Instance) []discovery.Instance {
	if len(instances) == 0 {
		return nil
	}
	out := make([]discovery.Instance, len(instances))
	copy(out, instances)
	return out
}

func rendezvousScore(channelID int64, instanceID string) uint64 {
	hasher := fnv.New64a()
	var raw [8]byte
	binary.LittleEndian.PutUint64(raw[:], uint64(channelID))
	_, _ = hasher.Write(raw[:])
	_, _ = hasher.Write([]byte(instanceID))
	return hasher.Sum64()
}
