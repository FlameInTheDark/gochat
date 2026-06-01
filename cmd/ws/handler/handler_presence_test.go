package handler

import (
	"testing"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/presence"
)

func strptr(s string) *string { return &s }

func TestMergePresenceUpdatePreservesVoiceFieldsWhenOmitted(t *testing.T) {
	voiceID := int64(42)
	existing := presence.SessionPresence{
		SessionID:      "session-1",
		Status:         presence.StatusOnline,
		Platform:       "desktop",
		VoiceChannelID: &voiceID,
		Mute:           true,
		Deafen:         true,
		SelfVideo:      true,
	}

	got := mergePresenceUpdate(existing, "session-1", mqmsg.PresenceUpdateRequest{
		Status:           presence.StatusOnline,
		Platform:         "web",
		CustomStatusText: strptr("busy"),
	}, 100, 30)

	if got.VoiceChannelID == nil || *got.VoiceChannelID != voiceID {
		t.Fatalf("expected voice channel %d to be preserved, got %#v", voiceID, got.VoiceChannelID)
	}
	if !got.Mute {
		t.Fatalf("expected mute=true to be preserved")
	}
	if !got.Deafen {
		t.Fatalf("expected deafen=true to be preserved")
	}
	if !got.SelfVideo {
		t.Fatalf("expected self_video=true to be preserved")
	}
	if got.Platform != "web" {
		t.Fatalf("expected platform to update, got %q", got.Platform)
	}
}

func TestMergePresenceUpdatePreservesCustomStatusWhenOmitted(t *testing.T) {
	existing := presence.SessionPresence{
		SessionID:        "session-1",
		Status:           presence.StatusOnline,
		CustomStatusText: "busy",
	}

	got := mergePresenceUpdate(existing, "session-1", mqmsg.PresenceUpdateRequest{
		Status:   presence.StatusIdle,
		Platform: "web",
	}, 100, 30)

	if got.CustomStatusText != "busy" {
		t.Fatalf("expected custom status to be preserved, got %q", got.CustomStatusText)
	}
}

func TestMergePresenceUpdateClearsCustomStatusWhenExplicitlyEmpty(t *testing.T) {
	existing := presence.SessionPresence{
		SessionID:        "session-1",
		Status:           presence.StatusOnline,
		CustomStatusText: "busy",
	}

	got := mergePresenceUpdate(existing, "session-1", mqmsg.PresenceUpdateRequest{
		Status:           presence.StatusOnline,
		Platform:         "web",
		CustomStatusText: strptr(""),
	}, 100, 30)

	if got.CustomStatusText != "" {
		t.Fatalf("expected custom status to be cleared, got %q", got.CustomStatusText)
	}
}

func TestPresenceChangedIgnoresSinceOnlyChanges(t *testing.T) {
	prev := presence.Presence{UserID: 1, Status: presence.StatusOnline, Since: 100}
	next := presence.Presence{UserID: 1, Status: presence.StatusOnline, Since: 200}

	if presenceChanged(prev, next) {
		t.Fatalf("expected presenceChanged to ignore since-only updates")
	}
}

func TestSameInt64Ptr(t *testing.T) {
	value := int64(1)
	other := int64(2)
	cases := []struct {
		name string
		a    *int64
		b    *int64
		want bool
	}{
		{name: "both nil", want: true},
		{name: "left nil", b: &value},
		{name: "same value", a: &value, b: &value, want: true},
		{name: "different value", a: &value, b: &other},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sameInt64Ptr(tc.a, tc.b); got != tc.want {
				t.Fatalf("sameInt64Ptr(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestPresenceCustomText(t *testing.T) {
	if got := presenceCustomText(nil, "existing"); got != "existing" {
		t.Fatalf("expected existing text, got %q", got)
	}
	if got := presenceCustomText(strptr(""), "existing"); got != "" {
		t.Fatalf("expected explicit empty text, got %q", got)
	}
	if got := presenceCustomText(strptr("new"), "existing"); got != "new" {
		t.Fatalf("expected explicit new text")
	}
}

func TestMergePresenceUpdateClearsVoiceChannelOnlyWhenExplicitlyRequested(t *testing.T) {
	voiceID := int64(42)
	clearVoice := int64(0)
	mute := false
	selfVideo := false
	existing := presence.SessionPresence{
		SessionID:      "session-1",
		Status:         presence.StatusOnline,
		VoiceChannelID: &voiceID,
		Mute:           true,
		Deafen:         true,
		SelfVideo:      true,
	}

	got := mergePresenceUpdate(existing, "session-1", mqmsg.PresenceUpdateRequest{
		Status:         presence.StatusOnline,
		VoiceChannelID: &clearVoice,
		Mute:           &mute,
		SelfVideo:      &selfVideo,
	}, 100, 30)

	if got.VoiceChannelID != nil {
		t.Fatalf("expected voice channel to be cleared, got %#v", got.VoiceChannelID)
	}
	if got.Mute {
		t.Fatalf("expected mute=false to be applied")
	}
	if !got.Deafen {
		t.Fatalf("expected deafen=true to be preserved when omitted")
	}
	if got.SelfVideo {
		t.Fatalf("expected self_video=false to be applied")
	}
}
