package handler

import (
	"reflect"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

func TestResolveRequestedChannelsPrefersChannelsList(t *testing.T) {
	t.Parallel()

	legacy := int64(999)
	got, ok := resolveRequestedChannels(mqmsg.Subscribe{
		Channel:  &legacy,
		Channels: []int64{22, 11, 22, -1, 0},
	})
	if !ok {
		t.Fatal("expected channels selection to be present")
	}

	want := []int64{11, 22}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected channel selection: got %v want %v", got, want)
	}
}

func TestResolveRequestedChannelsSupportsLegacySingleChannel(t *testing.T) {
	t.Parallel()

	channelID := int64(42)
	got, ok := resolveRequestedChannels(mqmsg.Subscribe{Channel: &channelID})
	if !ok {
		t.Fatal("expected legacy channel selection to be present")
	}

	want := []int64{42}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected channel selection: got %v want %v", got, want)
	}
}

func TestResolveRequestedChannelsIgnoresInvalidLegacyChannel(t *testing.T) {
	t.Parallel()

	channelID := int64(0)
	got, ok := resolveRequestedChannels(mqmsg.Subscribe{Channel: &channelID})
	if ok {
		t.Fatalf("expected invalid legacy channel to be ignored, got %v", got)
	}
}

func TestResolveRequestedChannelsAllowsClearingWithEmptyList(t *testing.T) {
	t.Parallel()

	got, ok := resolveRequestedChannels(mqmsg.Subscribe{Channels: []int64{}})
	if !ok {
		t.Fatal("expected empty channels selection to be present")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty channel selection, got %v", got)
	}
}

func TestBuildChannelSubscriptionDiffReplacesExactSet(t *testing.T) {
	t.Parallel()

	current := map[int64]struct{}{
		10: {},
		20: {},
	}
	subscribe, unsubscribe := buildChannelSubscriptionDiff(current, []int64{20, 30, 30})

	wantSubscribe := []int64{30}
	if !reflect.DeepEqual(subscribe, wantSubscribe) {
		t.Fatalf("unexpected subscribe diff: got %v want %v", subscribe, wantSubscribe)
	}

	wantUnsubscribe := []int64{10}
	if !reflect.DeepEqual(unsubscribe, wantUnsubscribe) {
		t.Fatalf("unexpected unsubscribe diff: got %v want %v", unsubscribe, wantUnsubscribe)
	}
}
