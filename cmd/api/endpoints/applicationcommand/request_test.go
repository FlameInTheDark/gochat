package applicationcommand

import (
	"encoding/json"
	"testing"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
)

func TestParseInvokeRequestAcceptsStringSnowflakes(t *testing.T) {
	req, err := parseInvokeRequest([]byte(`{"command_id":"2323620025379848192","channel_id":"2299526295437967361","options":[]}`), appcmd.InteractionTypeApplicationCommand)
	if err != nil {
		t.Fatalf("parseInvokeRequest returned error: %v", err)
	}
	if req.CommandID != 2323620025379848192 {
		t.Fatalf("unexpected command id: %d", req.CommandID)
	}
	if req.ChannelID != 2299526295437967361 {
		t.Fatalf("unexpected channel id: %d", req.ChannelID)
	}
	if req.Type != appcmd.InteractionTypeApplicationCommand {
		t.Fatalf("unexpected interaction type: %d", req.Type)
	}
}

func TestParseInvokeRequestAcceptsDiscordInteractionShape(t *testing.T) {
	body := []byte(`{
		"type": 2,
		"application_id": "562525284348329986",
		"guild_id": "514687794086412289",
		"channel_id": "565062979255795712",
		"data": {
			"id": "1222140942451085363",
			"name": "day",
			"type": 1,
			"options": [{"type": 3, "name": "location", "value": "test"}],
			"application_command": {"id": "1222140942451085363"}
		},
		"nonce": "1511131987748847616",
		"analytics_location": "slash_ui"
	}`)
	req, err := parseInvokeRequest(body, appcmd.InteractionTypeApplicationCommand)
	if err != nil {
		t.Fatalf("parseInvokeRequest returned error: %v", err)
	}
	if req.CommandID != 1222140942451085363 {
		t.Fatalf("unexpected command id: %d", req.CommandID)
	}
	if req.GuildID == nil || *req.GuildID != 514687794086412289 {
		t.Fatalf("unexpected guild id: %#v", req.GuildID)
	}
	if req.ChannelID != 565062979255795712 {
		t.Fatalf("unexpected channel id: %d", req.ChannelID)
	}
	if len(req.Options) != 1 || req.Options[0].Name != "location" || req.Options[0].Value != "test" {
		t.Fatalf("unexpected options: %#v", req.Options)
	}
}

func TestParseInvokeRequestPreservesNumericOptionValues(t *testing.T) {
	req, err := parseInvokeRequest([]byte(`{"command_id":"1","channel_id":"2","data":{"options":[{"type":4,"name":"count","value":3}]}}`), appcmd.InteractionTypeApplicationCommand)
	if err != nil {
		t.Fatalf("parseInvokeRequest returned error: %v", err)
	}
	value, ok := req.Options[0].Value.(json.Number)
	if !ok {
		t.Fatalf("expected json.Number option value, got %T", req.Options[0].Value)
	}
	if value.String() != "3" {
		t.Fatalf("unexpected numeric value: %s", value.String())
	}
}
