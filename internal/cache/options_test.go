package cache

import "testing"

func TestResolveTimedOptionsDefaultsToProactive(t *testing.T) {
	opts := ResolveTimedOptions()
	if !opts.Proactive {
		t.Fatalf("expected proactive refresh to be enabled by default")
	}
}

func TestResolveTimedOptionsCanDisableProactive(t *testing.T) {
	opts := ResolveTimedOptions(NoneProactive())
	if opts.Proactive {
		t.Fatalf("expected NoneProactive to disable proactive refresh")
	}
}
