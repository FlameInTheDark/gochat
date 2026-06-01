package goclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrNilContext is returned when a nil context is passed to a blocking call.
var ErrNilContext = errors.New("nil context")

// RESTError describes a non-2xx response from the bot REST API.
type RESTError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *RESTError) Error() string {
	body := strings.TrimSpace(string(e.Body))
	if body == "" {
		return e.Status
	}
	return fmt.Sprintf("%s: %s", e.Status, body)
}

// RequestOption mutates an outbound REST request.
type RequestOption func(*http.Request)

// WithHeader adds a header to an outbound REST request.
func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

// WithIdempotencyKey sets the idempotency key header used by the bot API.
func WithIdempotencyKey(key string) RequestOption {
	return WithHeader("Idempotency-Key", key)
}

// Request executes a raw bot REST API request.
func (s *Session) Request(ctx context.Context, method, path string, query url.Values, body any, out any, options ...RequestOption) error {
	if ctx == nil {
		return ErrNilContext
	}
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	attempts := s.MaxRestRetries + 1
	if attempts <= 0 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := s.doRequest(ctx, method, path, query, bodyBytes, body != nil, options...)
		if err != nil {
			if attempt+1 < attempts && isRetryableRequestError(err) {
				if err := sleepContext(ctx, retryDelay(attempt, nil)); err != nil {
					return err
				}
				continue
			}
			return err
		}
		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response body: %w", readErr)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if attempt+1 < attempts {
				if err := sleepContext(ctx, retryDelay(attempt, resp)); err != nil {
					return err
				}
				continue
			}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return &RESTError{StatusCode: resp.StatusCode, Status: resp.Status, Body: respBody}
		}
		if out != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("decode response body: %w", err)
			}
		}
		return nil
	}
	return nil
}

func (s *Session) doRequest(ctx context.Context, method, path string, query url.Values, body []byte, jsonBody bool, options ...RequestOption) (*http.Response, error) {
	endpoint := strings.TrimRight(s.APIEndpoint, "/") + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", s.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.UserAgent)
	if jsonBody {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, option := range options {
		option(req)
	}
	return s.Client.Do(req)
}

func isRetryableRequestError(err error) bool {
	return err != nil
}

func retryDelay(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if raw := resp.Header.Get("Retry-After"); raw != "" {
			if seconds, err := strconv.ParseFloat(raw, 64); err == nil && seconds > 0 {
				return time.Duration(seconds * float64(time.Second))
			}
		}
	}
	delay := time.Duration(1<<attempt) * 200 * time.Millisecond
	if delay > 3*time.Second {
		return 3 * time.Second
	}
	return delay
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// UserMe returns the current bot account.
func (s *Session) UserMe(ctx context.Context, options ...RequestOption) (*MeResponse, error) {
	var out MeResponse
	err := s.Request(ctx, http.MethodGet, EndpointUserMe, nil, nil, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Me returns the current bot account.
func (s *Session) Me(ctx context.Context, options ...RequestOption) (*MeResponse, error) {
	return s.UserMe(ctx, options...)
}

// Guilds lists guilds where the bot is installed.
func (s *Session) Guilds(ctx context.Context, options ...RequestOption) ([]*GuildResponse, error) {
	var out []*GuildResponse
	err := s.Request(ctx, http.MethodGet, EndpointGuilds, nil, nil, &out, options...)
	return out, err
}

// GuildChannels lists channels visible to the bot in a guild.
func (s *Session) GuildChannels(ctx context.Context, guildID int64, options ...RequestOption) ([]*Channel, error) {
	var out []*Channel
	err := s.Request(ctx, http.MethodGet, EndpointGuildChannels(formatID(guildID)), nil, nil, &out, options...)
	return out, err
}

// ChannelMessages lists messages visible to the bot.
func (s *Session) ChannelMessages(ctx context.Context, channelID int64, limit int, beforeID, afterID, aroundID int64, options ...RequestOption) ([]*Message, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	switch {
	case aroundID > 0:
		query.Set("from", formatID(aroundID))
		query.Set("direction", "around")
	case afterID > 0:
		query.Set("from", formatID(afterID))
		query.Set("direction", "after")
	case beforeID > 0:
		query.Set("from", formatID(beforeID))
		query.Set("direction", "before")
	}
	var out []*Message
	err := s.Request(ctx, http.MethodGet, EndpointChannelMessages(formatID(channelID)), query, nil, &out, options...)
	return out, err
}

// ChannelMessageSend sends a simple text message as the bot.
func (s *Session) ChannelMessageSend(ctx context.Context, channelID int64, content string, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageSendComplex(ctx, channelID, &MessageSend{Content: content}, options...)
}

// ChannelMessageSendComplex sends a message as the bot.
func (s *Session) ChannelMessageSendComplex(ctx context.Context, channelID int64, data *MessageSend, options ...RequestOption) (*Message, error) {
	if data == nil {
		data = &MessageSend{}
	}
	if err := ValidateEmbeds(data.Embeds); err != nil {
		return nil, err
	}
	var out Message
	err := s.Request(ctx, http.MethodPost, EndpointChannelMessages(formatID(channelID)), nil, data, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ChannelMessageEdit edits message content as the bot.
func (s *Session) ChannelMessageEdit(ctx context.Context, channelID, messageID int64, content string, options ...RequestOption) (*Message, error) {
	return s.ChannelMessageEditComplex(ctx, channelID, messageID, &MessageEdit{Content: &content}, options...)
}

// ChannelMessageEditComplex edits a message as the bot.
func (s *Session) ChannelMessageEditComplex(ctx context.Context, channelID, messageID int64, data *MessageEdit, options ...RequestOption) (*Message, error) {
	if data == nil {
		data = &MessageEdit{}
	}
	if data.Embeds != nil {
		if err := ValidateEmbeds(*data.Embeds); err != nil {
			return nil, err
		}
	}
	var out Message
	path := EndpointChannelMessage(formatID(channelID), formatID(messageID))
	err := s.Request(ctx, http.MethodPatch, path, nil, data, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ChannelMessageDelete deletes a message as the bot.
func (s *Session) ChannelMessageDelete(ctx context.Context, channelID, messageID int64, options ...RequestOption) error {
	path := EndpointChannelMessage(formatID(channelID), formatID(messageID))
	return s.Request(ctx, http.MethodDelete, path, nil, nil, nil, options...)
}

// ChannelMessageAck marks a channel read as the bot.
func (s *Session) ChannelMessageAck(ctx context.Context, channelID, messageID int64, options ...RequestOption) error {
	path := EndpointChannelMessageAck(formatID(channelID), formatID(messageID))
	return s.Request(ctx, http.MethodPost, path, nil, nil, nil, options...)
}

// ChannelTyping sends a typing indicator as the bot.
func (s *Session) ChannelTyping(ctx context.Context, channelID int64, options ...RequestOption) error {
	return s.Request(ctx, http.MethodPost, EndpointChannelTyping(formatID(channelID)), nil, nil, nil, options...)
}

// MessageReactionAdd adds the bot's reaction to a message.
func (s *Session) MessageReactionAdd(ctx context.Context, channelID, messageID int64, reactionName string, options ...RequestOption) (*MessageReaction, error) {
	var out MessageReaction
	path := EndpointMessageReactions(formatID(channelID), formatID(messageID), reactionName)
	err := s.Request(ctx, http.MethodPut, path, nil, nil, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// MessageReactionRemove removes the bot's reaction from a message.
func (s *Session) MessageReactionRemove(ctx context.Context, channelID, messageID int64, reactionName string, options ...RequestOption) error {
	return s.Request(ctx, http.MethodDelete, EndpointMessageReactions(formatID(channelID), formatID(messageID), reactionName), nil, nil, nil, options...)
}

// MessageReactionUsers lists users who reacted with a specific reaction.
func (s *Session) MessageReactionUsers(ctx context.Context, channelID, messageID int64, reactionName string, after int64, limit int, options ...RequestOption) (*MessageReactionUsersPage, error) {
	query := url.Values{}
	if after > 0 {
		query.Set("after", formatID(after))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	var out MessageReactionUsersPage
	err := s.Request(ctx, http.MethodGet, EndpointMessageReactions(formatID(channelID), formatID(messageID), reactionName), query, nil, &out, options...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
