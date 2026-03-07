package s3

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		useSSL   bool
		want     string
	}{
		{name: "empty", endpoint: "", useSSL: true, want: ""},
		{name: "http preserved", endpoint: "http://localhost:9000", useSSL: true, want: "http://localhost:9000"},
		{name: "https preserved", endpoint: "https://localhost:9000", useSSL: false, want: "https://localhost:9000"},
		{name: "ssl added", endpoint: "localhost:9000", useSSL: true, want: "https://localhost:9000"},
		{name: "plain added", endpoint: "localhost:9000", useSSL: false, want: "http://localhost:9000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeEndpoint(tc.endpoint, tc.useSSL); got != tc.want {
				t.Fatalf("normalizeEndpoint(%q, %t) = %q, want %q", tc.endpoint, tc.useSSL, got, tc.want)
			}
		})
	}
}

func TestReadNextUploadChunk(t *testing.T) {
	t.Run("empty reader", func(t *testing.T) {
		got, err := readNextUploadChunk(bytes.NewReader(nil), 8)
		if err != nil {
			t.Fatalf("readNextUploadChunk returned error: %v", err)
		}
		if got != nil {
			t.Fatalf("readNextUploadChunk returned %v, want nil", got)
		}
	})

	t.Run("partial final chunk", func(t *testing.T) {
		got, err := readNextUploadChunk(bytes.NewReader([]byte("hello")), 8)
		if err != nil {
			t.Fatalf("readNextUploadChunk returned error: %v", err)
		}
		if string(got) != "hello" {
			t.Fatalf("readNextUploadChunk returned %q, want %q", string(got), "hello")
		}
	})

	t.Run("full chunk", func(t *testing.T) {
		got, err := readNextUploadChunk(bytes.NewReader([]byte("12345678rest")), 8)
		if err != nil {
			t.Fatalf("readNextUploadChunk returned error: %v", err)
		}
		if string(got) != "12345678" {
			t.Fatalf("readNextUploadChunk returned %q, want %q", string(got), "12345678")
		}
	})

	t.Run("read error", func(t *testing.T) {
		wantErr := errors.New("boom")
		_, err := readNextUploadChunk(errReader{err: wantErr}, 8)
		if !errors.Is(err, wantErr) {
			t.Fatalf("readNextUploadChunk error = %v, want %v", err, wantErr)
		}
	})
}

type errReader struct {
	err error
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

var _ io.Reader = errReader{}
