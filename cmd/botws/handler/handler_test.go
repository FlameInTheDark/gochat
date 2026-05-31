package handler

import "testing"

func TestContainsSession(t *testing.T) {
	if !containsSession([]string{"a", "b"}, "b") {
		t.Fatalf("expected session to be found")
	}
	if containsSession([]string{"a", "b"}, "c") {
		t.Fatalf("did not expect session to be found")
	}
}
