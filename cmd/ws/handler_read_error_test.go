package main

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	fws "github.com/fasthttp/websocket"
)

func TestIsExpectedWSReadError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "normal close frame",
			err:  &fws.CloseError{Code: fws.CloseNormalClosure},
			want: true,
		},
		{
			name: "abnormal close unexpected eof",
			err:  &fws.CloseError{Code: fws.CloseAbnormalClosure, Text: io.ErrUnexpectedEOF.Error()},
			want: true,
		},
		{
			name: "plain unexpected eof",
			err:  io.ErrUnexpectedEOF,
			want: true,
		},
		{
			name: "closed network connection",
			err:  net.ErrClosed,
			want: true,
		},
		{
			name: "real read failure",
			err:  errors.New("boom"),
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isExpectedWSReadError(tt.err); got != tt.want {
				t.Fatalf("isExpectedWSReadError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestWSReadDeadlineUsesHeartbeatGrace(t *testing.T) {
	t.Parallel()

	if got, want := wsReadDeadline(35_000), 120*time.Second; got != want {
		t.Fatalf("wsReadDeadline() = %s, want %s", got, want)
	}
}
