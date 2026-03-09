package kvs

import "testing"

func TestNormalizeAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "localhost:6379"},
		{name: "ipv4 without port", in: "127.0.0.1", want: "127.0.0.1:6379"},
		{name: "hostname without port", in: "keydb", want: "keydb:6379"},
		{name: "existing port", in: "keydb:6380", want: "keydb:6380"},
		{name: "ipv6 without port", in: "::1", want: "[::1]:6379"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeAddr(tt.in); got != tt.want {
				t.Fatalf("normalizeAddr(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
