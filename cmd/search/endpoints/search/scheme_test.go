package search

import (
	"strings"
	"testing"
)

func TestMessageSearchRequestValidateRejectsNegativePage(t *testing.T) {
	req := MessageSearchRequest{
		ChannelId: 1,
		Page:      -1,
	}

	err := req.Validate()
	if err == nil || !strings.Contains(err.Error(), ErrPageInvalid) {
		t.Fatalf("expected %q validation error, got %v", ErrPageInvalid, err)
	}
}

func TestMessageSearchRequestValidateRejectsTooLongContent(t *testing.T) {
	content := strings.Repeat("a", maxSearchContentLength+1)
	req := MessageSearchRequest{
		ChannelId: 1,
		Content:   &content,
	}

	err := req.Validate()
	if err == nil || !strings.Contains(err.Error(), ErrContentTooLong) {
		t.Fatalf("expected %q validation error, got %v", ErrContentTooLong, err)
	}
}

func TestMessageSearchRequestValidateRejectsUnsupportedHasValue(t *testing.T) {
	req := MessageSearchRequest{
		ChannelId: 1,
		Has:       []string{"gif"},
	}

	err := req.Validate()
	if err == nil || !strings.Contains(err.Error(), ErrHasInvalid) {
		t.Fatalf("expected %q validation error, got %v", ErrHasInvalid, err)
	}
}
