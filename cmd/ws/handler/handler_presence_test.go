package handler

import (
	"testing"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/presence"
)

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
		CustomStatusText: "busy",
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
