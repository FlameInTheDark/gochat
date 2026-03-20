package msgsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"go.opentelemetry.io/otel/attribute"
)

type Search struct {
	osc *opensearch.Client
}

func closeQuietly(c io.Closer) {
	_ = c.Close()
}

// NewSearch creates a Search service.
func NewSearch(addresses []string, tlsSkip bool, username, password string) (*Search, error) {
	target := searchTarget(addresses)
	conf := opensearch.Config{
		Addresses: addresses,
		Transport: observability.NewHTTPTransport("opensearch", &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkip}}),
		Username:  username,
		Password:  password,
	}
	c, err := opensearch.NewClient(conf)
	if err != nil {
		return nil, err
	}

	// Init indices if not exist
	existsCtx, finishExists := observability.StartDependencySpan(context.Background(), "opensearch", "indices.exists", target, attribute.String("index", "messages"))
	res, err := c.Indices.Exists([]string{"messages"}, c.Indices.Exists.WithContext(existsCtx))
	if err != nil {
		finishExists(err)
		return nil, fmt.Errorf("failed to check if index exists: %w", err)
	}
	finishExists(nil)
	defer closeQuietly(res.Body)

	if res.StatusCode != 200 {
		ctx, finishCreate := observability.StartDependencySpan(context.Background(), "opensearch", "indices.create", target, attribute.String("index", "messages"))

		data, err := json.Marshal(defaultMessagesIndex)
		if err != nil {
			finishCreate(err)
			return nil, err
		}

		_, err = c.Indices.Create(
			"messages",
			c.Indices.Create.WithBody(bytes.NewReader(data)),
			c.Indices.Create.WithContext(ctx),
		)
		finishCreate(err)
		if err != nil {
			return nil, err
		}
	}

	return &Search{osc: c}, nil
}

func (s *Search) IndexMessage(ctx context.Context, m Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "index", "messages", attribute.Int64("channel.id", m.ChannelId))
	index, err := s.osc.Index(
		"messages",
		bytes.NewReader(data),
		s.osc.Index.WithDocumentID(fmt.Sprintf("%d", m.MessageId)),
		s.osc.Index.WithRouting(fmt.Sprintf("%d", m.ChannelId)),
		s.osc.Index.WithContext(ctx),
	)
	if err != nil {
		end(err)
		return err
	}
	if index.IsError() {
		err = fmt.Errorf("error indexing message: %s", index.String())
		end(err)
		return err
	}
	defer closeQuietly(index.Body)
	end(nil)
	return nil
}

type Results struct {
	Ids   []int64
	Total int
}

func (s *Search) Search(ctx context.Context, req SearchRequest) (results *Results, err error) {
	if s.osc == nil {
		return nil, fmt.Errorf("opensearch client is not initialized")
	}
	// Build query body with from/size using map-based builder
	q, err := buildOSQuery(req)
	if err != nil {
		return nil, err
	}

	opts := []func(*opensearchapi.SearchRequest){
		s.osc.Search.WithContext(ctx),
		s.osc.Search.WithIndex("messages"),
		s.osc.Search.WithBody(bytes.NewReader(q)),
		s.osc.Search.WithTrackTotalHits(true),
		s.osc.Search.WithRouting(fmt.Sprintf("%d", req.ChannelId)),
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "search", "messages", attribute.Int64("channel.id", req.ChannelId))
	opts[0] = s.osc.Search.WithContext(ctx)
	res, err := s.osc.Search(opts...)
	if err != nil {
		end(err)
		return nil, err
	}
	defer closeQuietly(res.Body)
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			end(nil)
			return
		}
		b, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("search response error: %s", string(b))
		end(err)
		return nil, err
	}

	var sr osSearchResponse
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&sr); err != nil {
		end(err)
		return nil, err
	}

	var ids = make([]int64, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		ids = append(ids, h.Source.MessageId)
	}
	end(nil)
	return &Results{Ids: ids, Total: sr.Hits.Total.Value}, nil
}

func buildOSQuery(req SearchRequest) ([]byte, error) {
	var osreq = osSearchRequest{
		From:   req.From,
		Size:   10,
		Source: []string{"message_id"},
	}
	osreq.Sort = []map[string]any{
		{"message_id": SortOrder{Order: "desc", Missing: "_last"}},
		{"_id": "desc"},
	}

	// Required channel filter
	osreq.Query.Bool.Filter = append(
		osreq.Query.Bool.Filter,
		osSearchQuery{Term: map[string]any{"channel_id": req.ChannelId}},
	)

	// Guild filter is optional so DM/private channel documents without guild_id remain searchable.
	if req.GuildId != nil {
		osreq.Query.Bool.Filter = append(
			osreq.Query.Bool.Filter,
			osSearchQuery{Term: map[string]any{"guild_id": *req.GuildId}},
		)
	}

	// Optional filters
	if req.UserId != nil {
		osreq.Query.Bool.Filter = append(
			osreq.Query.Bool.Filter,
			osSearchQuery{Term: map[string]any{"user_id": *req.UserId}},
		)
	}

	for _, v := range req.Mentions {
		osreq.Query.Bool.Filter = append(
			osreq.Query.Bool.Filter,
			osSearchQuery{Term: map[string]any{"mentions": v}},
		)
	}

	for _, v := range req.Has {
		osreq.Query.Bool.Filter = append(
			osreq.Query.Bool.Filter,
			osSearchQuery{Term: map[string]any{"has": v}},
		)
	}

	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if content != "" {
			osreq.Query.Bool.Must = append(
				osreq.Query.Bool.Must,
				osSearchQuery{Match: map[string]any{"content": content}},
			)
		}
	}

	data, err := json.Marshal(osreq)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Search) DeleteMessage(ctx context.Context, m DeleteMessage) error {
	if s.osc == nil {
		return fmt.Errorf("opensearch client is not initialized")
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "delete", "messages", attribute.Int64("channel.id", m.ChannelId))
	res, err := s.osc.Delete(
		"messages",
		fmt.Sprintf("%d", m.MessageId),
		s.osc.Delete.WithRouting(fmt.Sprintf("%d", m.ChannelId)),
		s.osc.Delete.WithContext(ctx),
	)
	if err != nil {
		end(err)
		return err
	}
	defer closeQuietly(res.Body)

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			end(nil)
			return nil
		}
		b, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("delete response error: %s", string(b))
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (s *Search) UpdateMessage(ctx context.Context, m Message) error {
	if s.osc == nil {
		return fmt.Errorf("opensearch client is not initialized")
	}

	payload := map[string]any{"doc": m}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "update", "messages", attribute.Int64("channel.id", m.ChannelId))
	res, err := opensearchapi.UpdateRequest{
		Index:      "messages",
		DocumentID: fmt.Sprintf("%d", m.MessageId),
		Routing:    fmt.Sprintf("%d", m.ChannelId),
		Body:       bytes.NewReader(data),
		Refresh:    "wait_for",
	}.Do(ctx, s.osc)
	if err != nil {
		end(err)
		return err
	}
	defer closeQuietly(res.Body)

	if res.StatusCode >= 300 {
		if res.StatusCode == http.StatusNotFound {
			end(nil)
			return nil
		}
		b, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("update response error: %s", string(b))
		end(err)
		return err
	}
	end(nil)
	return nil
}

func searchTarget(addresses []string) string {
	for _, address := range addresses {
		address = strings.TrimSpace(address)
		if address != "" {
			return address
		}
	}
	return "messages"
}
