package guild

import (
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
