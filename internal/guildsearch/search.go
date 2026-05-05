package guildsearch

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
	"go.opentelemetry.io/otel/attribute"
)

const indexName = "guilds"

type Search struct {
	osc *opensearch.Client
}

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

	existsCtx, finishExists := observability.StartDependencySpan(context.Background(), "opensearch", "indices.exists", target, attribute.String("index", indexName))
	res, err := c.Indices.Exists([]string{indexName}, c.Indices.Exists.WithContext(existsCtx))
	if err != nil {
		finishExists(err)
		return nil, fmt.Errorf("failed to check if guilds index exists: %w", err)
	}
	finishExists(nil)
	defer closeQuietly(res.Body)

	if res.StatusCode != http.StatusOK {
		ctx, finishCreate := observability.StartDependencySpan(context.Background(), "opensearch", "indices.create", target, attribute.String("index", indexName))
		data, err := json.Marshal(defaultGuildsIndex)
		if err != nil {
			finishCreate(err)
			return nil, err
		}
		create, err := c.Indices.Create(indexName, c.Indices.Create.WithBody(bytes.NewReader(data)), c.Indices.Create.WithContext(ctx))
		if err == nil && create != nil {
			defer closeQuietly(create.Body)
			if create.IsError() {
				body, _ := io.ReadAll(create.Body)
				err = fmt.Errorf("create guilds index response error: %s", string(body))
			}
		}
		finishCreate(err)
		if err != nil {
			return nil, err
		}
	}

	return &Search{osc: c}, nil
}

func (s *Search) IndexGuild(ctx context.Context, guild Guild) error {
	if s.osc == nil {
		return fmt.Errorf("opensearch client is not initialized")
	}
	if !guild.Public {
		return s.DeleteGuild(ctx, guild.GuildID)
	}

	data, err := json.Marshal(guild)
	if err != nil {
		return err
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "index", indexName, attribute.Int64("guild.id", guild.GuildID))
	res, err := s.osc.Index(
		indexName,
		bytes.NewReader(data),
		s.osc.Index.WithDocumentID(fmt.Sprintf("%d", guild.GuildID)),
		s.osc.Index.WithContext(ctx),
		s.osc.Index.WithRefresh("wait_for"),
	)
	if err != nil {
		end(err)
		return err
	}
	defer closeQuietly(res.Body)
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("index guild response error: %s", string(body))
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (s *Search) DeleteGuild(ctx context.Context, guildID int64) error {
	if s.osc == nil {
		return fmt.Errorf("opensearch client is not initialized")
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "delete", indexName, attribute.Int64("guild.id", guildID))
	res, err := s.osc.Delete(indexName, fmt.Sprintf("%d", guildID), s.osc.Delete.WithContext(ctx), s.osc.Delete.WithRefresh("wait_for"))
	if err != nil {
		end(err)
		return err
	}
	defer closeQuietly(res.Body)
	if res.IsError() && res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("delete guild response error: %s", string(body))
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (s *Search) SearchGuilds(ctx context.Context, req SearchRequest) (*Results, error) {
	q, err := buildOSQuery(req)
	if err != nil {
		return nil, err
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "search", indexName)
	res, err := s.osc.Search(
		s.osc.Search.WithContext(ctx),
		s.osc.Search.WithIndex(indexName),
		s.osc.Search.WithBody(bytes.NewReader(q)),
		s.osc.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		end(err)
		return nil, err
	}
	defer closeQuietly(res.Body)
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			end(nil)
			return &Results{}, nil
		}
		body, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("search guilds response error: %s", string(body))
		end(err)
		return nil, err
	}

	var sr osSearchResponse
	if err = json.NewDecoder(res.Body).Decode(&sr); err != nil {
		end(err)
		return nil, err
	}

	ids := make([]int64, 0, len(sr.Hits.Hits))
	for _, hit := range sr.Hits.Hits {
		ids = append(ids, hit.Source.GuildID)
	}
	end(nil)
	return &Results{Ids: ids, Total: sr.Hits.Total.Value}, nil
}

func (s *Search) SearchTags(ctx context.Context, query string, limit int) ([]string, error) {
	data, err := buildOSTagsQuery(query, limit)
	if err != nil {
		return nil, err
	}

	ctx, end := observability.StartDependencySpan(ctx, "opensearch", "search_tags", indexName)
	res, err := s.osc.Search(
		s.osc.Search.WithContext(ctx),
		s.osc.Search.WithIndex(indexName),
		s.osc.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		end(err)
		return nil, err
	}
	defer closeQuietly(res.Body)
	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			end(nil)
			return []string{}, nil
		}
		body, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("search guild tags response error: %s", string(body))
		end(err)
		return nil, err
	}

	var response struct {
		Aggregations struct {
			Tags struct {
				Buckets []struct {
					Key string `json:"key"`
				} `json:"buckets"`
			} `json:"tags"`
		} `json:"aggregations"`
	}
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		end(err)
		return nil, err
	}
	tags := make([]string, 0, len(response.Aggregations.Tags.Buckets))
	for _, bucket := range response.Aggregations.Tags.Buckets {
		tags = append(tags, bucket.Key)
	}
	end(nil)
	return tags, nil
}

func buildOSQuery(req SearchRequest) ([]byte, error) {
	if req.Size <= 0 {
		req.Size = 16
	}
	if req.Size > 16 {
		req.Size = 16
	}

	osreq := osSearchRequest{
		From:   req.From,
		Size:   req.Size,
		Source: []string{"guild_id"},
	}
	osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Term: map[string]any{"public": true}})
	for _, tag := range req.Tags {
		if tag == "" {
			continue
		}
		osreq.Query.Bool.Filter = append(osreq.Query.Bool.Filter, osSearchQuery{Term: map[string]any{"tags": tag}})
	}

	query := strings.TrimSpace(req.Query)
	if query != "" {
		osreq.Query.Bool.Must = append(osreq.Query.Bool.Must, osSearchQuery{MultiMatch: map[string]any{
			"query":  query,
			"fields": []string{"name^3", "tags^2", "description"},
		}})
	}

	switch req.Sort {
	case SortPopularity:
		osreq.Sort = []map[string]any{
			{"members_count": SortOrder{Order: "desc", Missing: "_last"}},
			{"_score": SortOrder{Order: "desc"}},
			{"guild_id": SortOrder{Order: "asc"}},
		}
	case SortAlphabetical:
		osreq.Sort = []map[string]any{
			{"name.keyword": SortOrder{Order: "asc", Missing: "_last"}},
			{"guild_id": SortOrder{Order: "asc"}},
		}
	default:
		osreq.Sort = []map[string]any{
			{"_score": SortOrder{Order: "desc"}},
			{"guild_id": SortOrder{Order: "asc"}},
		}
	}

	return json.Marshal(osreq)
}

func buildOSTagsQuery(query string, limit int) ([]byte, error) {
	if limit <= 0 {
		limit = 16
	}
	if limit > 16 {
		limit = 16
	}

	req := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{"term": map[string]any{"public": true}},
				},
			},
		},
		"aggs": map[string]any{
			"tags": map[string]any{
				"terms": map[string]any{
					"field": "tags",
					"size":  limit,
				},
			},
		},
	}
	query = strings.TrimSpace(query)
	if query != "" {
		req["query"].(map[string]any)["bool"].(map[string]any)["filter"] = append(
			req["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]map[string]any),
			map[string]any{"prefix": map[string]any{"tags": query}},
		)
	}

	return json.Marshal(req)
}

func closeQuietly(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}

func searchTarget(addresses []string) string {
	for _, address := range addresses {
		address = strings.TrimSpace(address)
		if address != "" {
			return address
		}
	}
	return indexName
}
