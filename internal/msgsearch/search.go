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
	q, err := buildOSQuery(req)
	if err != nil {
		return nil, err
	}

	opts := []func(*opensearchapi.SearchRequest){
		s.osc.Search.WithContext(ctx),
		s.osc.Search.WithIndex("messages"),
		s.osc.Search.WithBody(bytes.NewReader(q)),
		s.osc.Search.WithTrackTotalHits(true),
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
		ids = append(ids, h.Source.MessageId)
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
			Source struct {
				MessageId int64 `json:"message_id"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type osSearchRequest struct {
	From   int      `json:"from,omitempty"`
	Size   int      `json:"size,omitempty"`
	Source []string `json:"_source,omitempty"`
	Query  struct {
		Bool struct {
			Filter []osSearchQuery `json:"filter,omitempty"`
			Must   []osSearchQuery `json:"must,omitempty"`
		} `json:"bool,omitempty"`
	} `json:"query,omitempty"`
}

type osSearchQuery struct {
	Match map[string]any `json:"match,omitempty"`
	Term  map[string]any `json:"term,omitempty"`
}

func buildOSQuery(req SearchRequest) ([]byte, error) {
	var osreq = osSearchRequest{
		From:   req.From,
		Size:   10,
		Source: []string{"message_id"},
	}

	// Required guild filter
	osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Match: map[string]any{"guild_id": req.GuildId}})

	// Optional filters
	if req.ChannelId != nil {
		osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Term: map[string]any{"channel_id": *req.ChannelId}})
	}
	if req.AuthorId != nil {
		osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Term: map[string]any{"user_id": *req.AuthorId}})
	}
	if len(req.Mentions) > 0 {
		osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Term: map[string]any{"mentions": req.Mentions}})
	}
	if len(req.Has) > 0 {
		osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Term: map[string]any{"has": req.Has}})
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if content != "" {
			osreq.Query.Bool.Must = append(osreq.Query.Bool.Must, osSearchQuery{Match: map[string]any{"content": content}})
		}
	}

	data, err := json.Marshal(osreq)
	if err != nil {
		return nil, err
	}

	return data, nil
}
