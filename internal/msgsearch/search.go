package msgsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type Search struct{ osc *opensearch.Client }

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
	return &Search{osc: c}, nil
}

func (s *Search) IndexMessage(ctx context.Context, m ...AddMessage) error { return nil }

type Results struct {
	Ids   []int64
	Total int
}

func (s *Search) Search(ctx context.Context, req SearchRequest) (results *Results, err error) {
	if s.osc == nil {
		return nil, fmt.Errorf("opensearch client is not initialized")
	}
	// Build query body with from/size using map-based builder
	q := buildOSQuery(req)
	body := map[string]any{
		"from": req.From,
		"size": 10,
	}
	if q != nil {
		body["query"] = q
	} else {
		body["query"] = map[string]any{"match_all": map[string]any{}}
	}
	queryBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	opts := []func(*opensearchapi.SearchRequest){
		s.osc.Search.WithContext(ctx),
		s.osc.Search.WithIndex("messages"),
		s.osc.Search.WithBody(bytes.NewReader(queryBytes)),
		s.osc.Search.WithTrackTotalHits(true),
		s.osc.Search.WithSource("false"),
	}
	if req.ChannelId != nil {
		opts = append(opts, s.osc.Search.WithRouting(fmt.Sprintf("%d", *req.ChannelId)))
	}
	res, err := s.osc.Search(opts...)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
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
		id64, err := strconv.ParseInt(h.ID, 10, 64)
		if err == nil {
			ids = append(ids, id64)
		}
	}
	return &Results{Ids: ids, Total: sr.Hits.Total.Value}, nil
}

type osSearchResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			ID string `json:"_id"`
		} `json:"hits"`
	} `json:"hits"`
}

func buildOSQuery(req SearchRequest) map[string]any {
	must := make([]any, 0, 4)
	filter := make([]any, 0, 4)

	// Required guild filter
	filter = append(filter, map[string]any{"term": map[string]any{"guild_id": req.GuildId}})
	// Optional filters
	if req.ChannelId != nil {
		filter = append(filter, map[string]any{"term": map[string]any{"channel_id": *req.ChannelId}})
	}
	if req.AuthorId != nil {
		filter = append(filter, map[string]any{"term": map[string]any{"user_id": *req.AuthorId}})
	}
	if len(req.Mentions) > 0 {
		filter = append(filter, map[string]any{"terms": map[string]any{"mentions": req.Mentions}})
	}
	if len(req.Has) > 0 {
		filter = append(filter, map[string]any{"terms": map[string]any{"has": req.Has}})
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if content != "" {
			must = append(must, map[string]any{"match": map[string]any{"content": content}})
		}
	}

	boolNode := map[string]any{}
	if len(must) > 0 {
		boolNode["must"] = must
	}
	if len(filter) > 0 {
		boolNode["filter"] = filter
	}
	if len(boolNode) == 0 {
		return nil
	}
	return map[string]any{"bool": boolNode}
}
