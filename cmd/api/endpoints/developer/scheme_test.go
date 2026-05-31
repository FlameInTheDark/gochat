package developer

import (
	"reflect"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

func TestNormalizeBotDiscoveryTags(t *testing.T) {
	got := normalizeBotDiscoveryTags([]string{"  Moderation ", "music", "moderation", "", "TOOLS"})
	want := []string{"moderation", "music", "tools"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeBotDiscoveryTags() = %#v, want %#v", got, want)
	}
}

func TestValidateBotDiscoveryTags(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantErr bool
	}{
		{name: "empty", tags: nil},
		{name: "valid", tags: []string{"moderation", "music-bot", "ai_tools"}},
		{name: "too many", tags: []string{"aa", "bb", "cc", "dd", "ee", "ff", "gg", "hh", "ii", "jj", "kk"}, wantErr: true},
		{name: "too short", tags: []string{"a"}, wantErr: true},
		{name: "uppercase", tags: []string{"Moderation"}, wantErr: true},
		{name: "symbol", tags: []string{"music!"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBotDiscoveryTags(tt.tags)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateBotDiscoveryTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeGrantRequestMaxUses(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "omitted default", in: 0, want: 1},
		{name: "positive", in: 5, want: 5},
		{name: "unlimited sentinel", in: -1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateGrantRequest{MaxUses: tt.in}

			normalizeGrantRequest(&req, 123)

			if req.MaxUses != tt.want {
				t.Fatalf("MaxUses = %d, want %d", req.MaxUses, tt.want)
			}
		})
	}
}

func TestGrantHasUsesRemainingUnlimited(t *testing.T) {
	grant := model.BotInstallGrant{MaxUses: 0, Uses: 100}

	if !grantHasUsesRemaining(grant) {
		t.Fatal("grantHasUsesRemaining() = false, want true")
	}
}
