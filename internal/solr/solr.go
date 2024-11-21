package solr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type SolrClient struct {
	baseURL string
}

func New(baseURL string) *SolrClient {
	return &SolrClient{baseURL: strings.TrimRight(baseURL, "/")}
}

func (c *SolrClient) Query(ctx context.Context, collection string, q QueryBuilder) (*Response, error) {
	data, err := json.Marshal(&q)
	if err != nil {
		return nil, err
	}

	cl := resty.New()
	resp, err := cl.R().
		SetContext(ctx).
		SetBody(data).
		SetHeader("Content-Type", "application/json").
		Post(fmt.Sprintf("%s/solr/%s/query", c.baseURL, collection))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	var r Response
	if err := json.Unmarshal(resp.Body(), &r); err != nil {
		return nil, errors.Join(err, fmt.Errorf("unable to parse response"))
	}
	return &r, nil
}

func (c *SolrClient) Add(ctx context.Context, collection string, v interface{}) (*Response, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	cl := resty.New()
	resp, err := cl.R().
		SetContext(ctx).
		SetBody(data).
		SetHeader("Content-Type", "application/json").
		Post(fmt.Sprintf("%s/solr/%s/update?softCommit=true", c.baseURL, collection))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}
	var r Response
	if err := json.Unmarshal(resp.Body(), &r); err != nil {
		return nil, err
	}
	return &r, nil
}
