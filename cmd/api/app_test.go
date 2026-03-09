package main

import "testing"

func TestBuildContentHosts(t *testing.T) {
	hosts, err := buildContentHosts([]string{
		"https://cdn.example.com/public/files",
		"https://cdn.example.com/emoji",
		" https://storage.example.com/bucket ",
	})
	if err != nil {
		t.Fatalf("buildContentHosts returned error: %v", err)
	}

	want := []string{
		"https://cdn.example.com",
		"https://storage.example.com",
	}

	if len(hosts) != len(want) {
		t.Fatalf("unexpected hosts length: got %v want %v", hosts, want)
	}
	for i := range want {
		if hosts[i] != want[i] {
			t.Fatalf("unexpected hosts: got %v want %v", hosts, want)
		}
	}
}

func TestBuildContentHostsRejectsInvalidURL(t *testing.T) {
	if _, err := buildContentHosts([]string{"cdn.example.com"}); err == nil {
		t.Fatal("expected invalid content host to fail")
	}
}
