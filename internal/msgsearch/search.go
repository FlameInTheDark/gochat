package msgsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/solr"
)

const collection = "gochat"

type Search struct {
	s *solr.SolrClient
}

func NewSearch(s *solr.SolrClient) *Search {
	return &Search{s: s}
}

func (s *Search) IndexMessage(ctx context.Context, m ...AddMessage) error {
	resp, err := s.s.Add(ctx, collection, m)
	if err != nil {
		return fmt.Errorf("error adding message to index: %w", err)
	}

	if resp.Header.Status != 0 {
		return fmt.Errorf("error adding message to index, bad status code: %d", resp.Header.Status)
	}

	return nil
}

func (s *Search) Search(ctx context.Context, req SearchRequest) (ids []int64, err error) {
	and := solr.And{solr.M{"guild_id": req.GuildId}}
	if req.Has != nil {
		and = append(and, solr.M{"has": *req.Has})
	}
	if req.Content != nil {
		and = append(and, solr.M{"content": *req.Content})
	}
	if req.Mentions != nil {
		and = append(and, solr.M{"mentions": *req.Mentions})
	}
	query := solr.Builder.Query(and)
	if req.ChannelId != nil {
		query = query.Filter(solr.F{"channel_id": *req.ChannelId})
	}
	if req.AuthorId != nil {
		query = query.Filter(solr.F{"author_id": *req.AuthorId})
	}
	res, err := s.s.Query(ctx, collection, query)
	if err != nil {
		return nil, err
	}
	var v []SearchMessageResponse
	err = json.Unmarshal(res.Response.Docs, &v)
	if err != nil {
		return nil, err
	}
	for _, m := range v {
		ids = append(ids, m.MessageId...)
	}
	return
}
