package guild

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

type fakeDiscoveryManager struct {
	lists     map[string][]discovery.Instance
	listCalls map[string]int
	regions   []string
}

func (f *fakeDiscoveryManager) Register(ctx context.Context, region string, inst discovery.Instance) error {
	return nil
}

func (f *fakeDiscoveryManager) List(ctx context.Context, region string) ([]discovery.Instance, error) {
	if f.listCalls == nil {
		f.listCalls = make(map[string]int)
	}
	f.listCalls[region]++
	return cloneDiscoveryInstances(f.lists[region]), nil
}

func (f *fakeDiscoveryManager) Regions(ctx context.Context) ([]string, error) {
	return append([]string(nil), f.regions...), nil
}

func newVoiceTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestChannelBindingForJoinFallsBackToAnotherRegion(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	disco := &fakeDiscoveryManager{
		lists: map[string][]discovery.Instance{
			"us-east":    nil,
			"eu-central": {{ID: "sfu-eu-1", Region: "eu-central", URL: "wss://eu.example/signal", Load: 3}},
		},
	}
	e := &entity{
		cache:              cache,
		ch:                 &fakeCreateChannelRepo{},
		defaultVoiceRegion: "us-east",
		disco:              disco,
		allowedRegions:     map[string]struct{}{"us-east": {}, "eu-central": {}},
		allowedRegionIDs:   []string{"eu-central", "us-east"},
		voiceSelector:      newVoiceSelector(newVoiceTestLogger()),
	}

	binding, err := e.channelBindingForJoin(context.Background(), 42)
	if err != nil {
		t.Fatalf("channelBindingForJoin returned error: %v", err)
	}
	if binding.Region != "eu-central" || binding.ID != "sfu-eu-1" {
		t.Fatalf("unexpected binding: %#v", binding)
	}
	if got := disco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected one preferred-region lookup, got %d", got)
	}
	if got := disco.listCalls["eu-central"]; got != 1 {
		t.Fatalf("expected one fallback-region lookup, got %d", got)
	}

	again, err := e.channelBindingForJoin(context.Background(), 42)
	if err != nil {
		t.Fatalf("second channelBindingForJoin returned error: %v", err)
	}
	if again != binding {
		t.Fatalf("expected cached binding %#v, got %#v", binding, again)
	}
	if got := disco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected cached route to avoid additional preferred lookups, got %d", got)
	}
	if got := disco.listCalls["eu-central"]; got != 1 {
		t.Fatalf("expected cached route to avoid additional fallback lookups, got %d", got)
	}
}

func TestSelectSFUBindingReusesDiscoverySnapshotAcrossColdChannels(t *testing.T) {
	disco := &fakeDiscoveryManager{
		lists: map[string][]discovery.Instance{
			"us-east": {
				{ID: "sfu-1", Region: "us-east", URL: "wss://1.example/signal", Load: 4},
				{ID: "sfu-2", Region: "us-east", URL: "wss://2.example/signal", Load: 4},
			},
		},
	}
	e := &entity{
		ch:                 &fakeCreateChannelRepo{},
		defaultVoiceRegion: "us-east",
		disco:              disco,
		allowedRegions:     map[string]struct{}{"us-east": {}},
		allowedRegionIDs:   []string{"us-east"},
		voiceSelector:      newVoiceSelector(newVoiceTestLogger()),
	}

	first, err := e.selectSFUBinding(context.Background(), 1001, "us-east", true)
	if err != nil {
		t.Fatalf("first selectSFUBinding returned error: %v", err)
	}
	second, err := e.selectSFUBinding(context.Background(), 1002, "us-east", true)
	if err != nil {
		t.Fatalf("second selectSFUBinding returned error: %v", err)
	}
	if first.URL == "" || second.URL == "" {
		t.Fatalf("expected both selections to resolve an SFU, got first=%#v second=%#v", first, second)
	}
	if got := disco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected one discovery refresh shared across cold joins, got %d", got)
	}
}

func TestVoiceSelectorReservationsSpreadBurstAcrossNearEqualNodes(t *testing.T) {
	selector := newVoiceSelector(newVoiceTestLogger())
	instances := []discovery.Instance{
		{ID: "sfu-a", Region: "eu-central", URL: "wss://a.example/signal", Load: 0},
		{ID: "sfu-b", Region: "eu-central", URL: "wss://b.example/signal", Load: 0},
	}

	first, ok := selector.pickBinding(5001, "eu-central", instances)
	if !ok {
		t.Fatal("expected first pick to succeed")
	}
	second, ok := selector.pickBinding(5002, "eu-central", instances)
	if !ok {
		t.Fatal("expected second pick to succeed")
	}
	if first.ID == second.ID {
		t.Fatalf("expected reservations to steer the second cold join to another SFU, got %q twice", first.ID)
	}
}
