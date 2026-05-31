package guild

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

func TestCreateGuildChannelCategoryRequestValidate(t *testing.T) {
	t.Run("accepts valid name", func(t *testing.T) {
		req := CreateGuildChannelCategoryRequest{Name: "voice"}

		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	t.Run("requires name", func(t *testing.T) {
		req := CreateGuildChannelCategoryRequest{}

		err := req.Validate()
		if err == nil {
			t.Fatal("Validate() error = nil")
		}
		if !strings.Contains(err.Error(), ErrChannelNameRequired) {
			t.Fatalf("Validate() error = %v, want message containing %q", err, ErrChannelNameRequired)
		}
	})
}

func TestInstallBotRequestUnmarshalBotUserID(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int64
	}{
		{
			name: "string snowflake",
			body: `{"bot_user_id":"2321114293732376576","granted_permissions":559105}`,
			want: 2321114293732376576,
		},
		{
			name: "number snowflake",
			body: `{"bot_user_id":2321114293732376576,"granted_permissions":559105}`,
			want: 2321114293732376576,
		},
		{
			name: "missing snowflake",
			body: `{"granted_permissions":559105}`,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req InstallBotRequest

			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if got := req.BotUserId; got != tt.want {
				t.Fatalf("BotUserId = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInstallBotRequestUnmarshalBotUserIDRejectsInvalidString(t *testing.T) {
	var req InstallBotRequest

	if err := json.Unmarshal([]byte(`{"bot_user_id":"not-a-snowflake"}`), &req); err == nil {
		t.Fatal("Unmarshal() error = nil")
	}
}

func TestCreateGuildChannelRequestValidate(t *testing.T) {
	t.Run("accepts valid name", func(t *testing.T) {
		req := CreateGuildChannelRequest{
			Name: "voice",
			Type: model.ChannelTypeGuild,
		}

		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	t.Run("requires name", func(t *testing.T) {
		req := CreateGuildChannelRequest{
			Type: model.ChannelTypeGuild,
		}

		err := req.Validate()
		if err == nil {
			t.Fatal("Validate() error = nil")
		}
		if !strings.Contains(err.Error(), ErrChannelNameRequired) {
			t.Fatalf("Validate() error = %v, want message containing %q", err, ErrChannelNameRequired)
		}
	})
}
