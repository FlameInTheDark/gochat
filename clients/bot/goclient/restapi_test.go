package goclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type expectedRequest struct {
	method string
	path   string
	query  url.Values
	body   string
	reply  string
	status int
}

func TestBotRESTMethodsUseCurrentRuntimeRoutes(t *testing.T) {
	expected := []expectedRequest{
		{method: http.MethodGet, path: "/bot/api/v1/user/me", reply: `{"bot_user_id":1,"owner_user_id":2,"user":{"id":1,"name":"bot","discriminator":"0001","is_bot":true}}`},
		{method: http.MethodGet, path: "/bot/api/v1/guild", reply: `[{"id":2230469276416868352,"name":"guild","owner":1,"public":true,"granted_permissions":7}]`},
		{method: http.MethodGet, path: "/bot/api/v1/guild/2230469276416868352/channels", reply: `[]`},
		{method: http.MethodGet, path: "/bot/api/v1/message/channel/2230469276416868353", query: url.Values{"direction": {"after"}, "from": {"2230469276416868354"}, "limit": {"25"}}, reply: `[]`},
		{method: http.MethodPost, path: "/bot/api/v1/message/channel/2230469276416868353", body: `{"content":"hello"}`, reply: `{"id":2230469276416868355,"channel_id":2230469276416868353,"author":{"id":1,"name":"bot","discriminator":"0001","is_bot":true},"content":"hello","type":0}`},
		{method: http.MethodPatch, path: "/bot/api/v1/message/channel/2230469276416868353/2230469276416868355", body: `{"content":"edited"}`, reply: `{"id":2230469276416868355,"channel_id":2230469276416868353,"author":{"id":1,"name":"bot","discriminator":"0001","is_bot":true},"content":"edited","type":0}`},
		{method: http.MethodDelete, path: "/bot/api/v1/message/channel/2230469276416868353/2230469276416868355", status: http.StatusNoContent},
		{method: http.MethodPost, path: "/bot/api/v1/message/channel/2230469276416868353/2230469276416868355/ack", status: http.StatusNoContent},
		{method: http.MethodPost, path: "/bot/api/v1/message/channel/2230469276416868353/typing", status: http.StatusNoContent},
		{method: http.MethodPut, path: "/bot/api/v1/message/channel/2230469276416868353/2230469276416868355/reactions/smile:123", reply: `{"count":1,"me":true,"emoji":{"id":123,"name":"smile"}}`},
		{method: http.MethodDelete, path: "/bot/api/v1/message/channel/2230469276416868353/2230469276416868355/reactions/smile:123", status: http.StatusNoContent},
		{method: http.MethodGet, path: "/bot/api/v1/message/channel/2230469276416868353/2230469276416868355/reactions/smile:123", query: url.Values{"after": {"2230469276416868356"}, "limit": {"50"}}, reply: `{"items":[]}`},
	}
	index := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if index >= len(expected) {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
		}
		want := expected[index]
		index++
		if req.Method != want.method {
			t.Fatalf("request %d method = %s, want %s", index, req.Method, want.method)
		}
		if req.URL.EscapedPath() != want.path {
			t.Fatalf("request %d path = %s, want %s", index, req.URL.EscapedPath(), want.path)
		}
		if want.query != nil && req.URL.Query().Encode() != want.query.Encode() {
			t.Fatalf("request %d query = %s, want %s", index, req.URL.Query().Encode(), want.query.Encode())
		}
		if got := req.Header.Get("Authorization"); got != "Bot gcb_test" {
			t.Fatalf("authorization = %q", got)
		}
		if want.body != "" {
			raw, _ := io.ReadAll(req.Body)
			if strings.TrimSpace(string(raw)) != want.body {
				t.Fatalf("request %d body = %s, want %s", index, string(raw), want.body)
			}
		}
		status := want.status
		if status == 0 {
			status = http.StatusOK
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(want.reply)),
		}, nil
	})}

	s, err := New("gcb_test", WithAPIEndpoint("http://bot-api.test"), WithHTTPClient(httpClient))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := s.UserMe(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Guilds(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GuildChannels(ctx, 2230469276416868352); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChannelMessages(ctx, 2230469276416868353, 25, 0, 2230469276416868354, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChannelMessageSend(ctx, 2230469276416868353, "hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ChannelMessageEdit(ctx, 2230469276416868353, 2230469276416868355, "edited"); err != nil {
		t.Fatal(err)
	}
	if err := s.ChannelMessageDelete(ctx, 2230469276416868353, 2230469276416868355); err != nil {
		t.Fatal(err)
	}
	if err := s.ChannelMessageAck(ctx, 2230469276416868353, 2230469276416868355); err != nil {
		t.Fatal(err)
	}
	if err := s.ChannelTyping(ctx, 2230469276416868353); err != nil {
		t.Fatal(err)
	}
	if _, err := s.MessageReactionAdd(ctx, 2230469276416868353, 2230469276416868355, "smile:123"); err != nil {
		t.Fatal(err)
	}
	if err := s.MessageReactionRemove(ctx, 2230469276416868353, 2230469276416868355, "smile:123"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.MessageReactionUsers(ctx, 2230469276416868353, 2230469276416868355, "smile:123", 2230469276416868356, 50); err != nil {
		t.Fatal(err)
	}
	if index != len(expected) {
		t.Fatalf("performed %d requests, want %d", index, len(expected))
	}
}
