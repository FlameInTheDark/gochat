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

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type Search struct {
	osc *opensearch.Client
}

// NewSearch creates a Search service.
func NewSearch(addresses []string, tlsSkip bool, username, password string) (*Search, error) {
	conf := opensearch.Config{
		Addresses: addresses,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkip}},
		Username:  username,
		Password:  password,
	}
	c, err := opensearch.NewClient(conf)
	if err != nil {
		return nil, err
	}

	// Init indices if not exist
	res, err := c.Indices.Exists([]string{"messages"})
	if err != nil {
		return nil, fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		ctx := context.Background()

		data, err := json.Marshal(defaultMessagesIndex)
		if err != nil {
			return nil, err
		}

		_, err = c.Indices.Create(
			"messages",
			c.Indices.Create.WithBody(bytes.NewReader(data)),
			c.Indices.Create.WithContext(ctx),
		)
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
	index, err := s.osc.Index(
		"messages",
		bytes.NewReader(data),
		s.osc.Index.WithDocumentID(fmt.Sprintf("%d", m.MessageId)),
		s.osc.Index.WithRouting(fmt.Sprintf("%d", m.ChannelId)),
		s.osc.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	if index.IsError() {
		return fmt.Errorf("error indexing message: %s", index.String())
	}
	defer index.Body.Close()
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

	res, err := s.osc.Search(opts...)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return
		}
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search response error: %s", string(b))
	}

	var sr osSearchResponse
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&sr); err != nil {
		return nil, err
	}

	var ids = make([]int64, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		ids = append(ids, h.Source.MessageId)
	}
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

	// Required guild filter
	osreq.Query.Bool.Filter = append(
		osreq.Query.Bool.Filter,
		osSearchQuery{Term: map[string]any{"guild_id": req.GuildId}},
	)

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

	res, err := s.osc.Delete(
		"messages",
		fmt.Sprintf("%d", m.MessageId),
		s.osc.Delete.WithRouting(fmt.Sprintf("%d", m.ChannelId)),
		s.osc.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil
		}
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete response error: %s", string(b))
	}
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

	res, err := opensearchapi.UpdateRequest{
		Index:      "messages",
		DocumentID: fmt.Sprintf("%d", m.MessageId),
		Routing:    fmt.Sprintf("%d", m.ChannelId),
		Body:       bytes.NewReader(data),
		Refresh:    "wait_for",
	}.Do(ctx, s.osc)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		if res.StatusCode == http.StatusNotFound {
			return nil
		}
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("update response error: %s", string(b))
	}
	return nil
}
