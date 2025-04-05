package sendpulse_sdk_go

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type TemplatesService struct {
	client *Client
}

func newTemplatesService(cl *Client) *TemplatesService {
	return &TemplatesService{client: cl}
}

func (service *TemplatesService) CreateTemplate(ctx context.Context, name string, body string, lang string) (int, error) {
	path := "/template"

	type paramsFormat struct {
		Name string `json:"name,omitempty"`
		Lang string `json:"lang"`
		Body string `json:"body"`
	}

	params := paramsFormat{
		Body: b64.StdEncoding.EncodeToString([]byte(body)),
		Lang: lang,
	}

	if name != "" {
		params.Name = name
	}

	var response struct {
		Result bool `json:"result"`
		RealID int  `json:"real_id"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &response, true)
	return response.RealID, err
}

func (service *TemplatesService) UpdateTemplate(ctx context.Context, templateID int, body string, lang string) error {
	path := fmt.Sprintf("/template/edit/%d", templateID)

	type paramsFormat struct {
		Lang string `json:"lang"`
		Body string `json:"body"`
	}

	params := paramsFormat{
		Body: b64.StdEncoding.EncodeToString([]byte(body)),
		Lang: lang,
	}

	var response struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &response, true)
	return err
}

type TemplateCategory struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	MetaDescription string `json:"meta_description"`
	FullDescription string `json:"full_description"`
	Code            string `json:"code"`
	Sort            int    `json:"sort"`
}

type Template struct {
	ID              string            `json:"id"`
	RealID          int               `json:"real_id"`
	Name            string            `json:"name"`
	NameSlug        string            `json:"name_slug"`
	Lang            string            `json:"lang"`
	MetaDescription string            `json:"meta_description"`
	FullDescription string            `json:"full_description"`
	Category        string            `json:"category"`
	CategoryInfo    *TemplateCategory `json:"category_info"`
	Tags            map[string]string `json:"tags"`
	Mark            string            `json:"mark"`       //
	MarkCount       int               `json:"mark_count"` //
	Body            string            `json:"body"`       //
	Owner           string            `json:"owner"`
	Created         DateTimeType      `json:"created"`
	Preview         string            `json:"preview"`
	IsStructure     bool              `json:"is_structure"`
}

func (t *Template) UnmarshalJSON(data []byte) error {
	raw := string(data)
	data = []byte(strings.ReplaceAll(raw, `"category_info": []`, `"category_info": null`))

	type template Template
	if err := json.Unmarshal(data, (*template)(t)); err != nil {
		return err
	}
	return nil
}

func (service *TemplatesService) GetTemplate(ctx context.Context, templateID int) (*Template, error) {
	path := fmt.Sprintf("/template/%d", templateID)
	var respData Template
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return &respData, err
}

func (service *TemplatesService) GetTemplates(ctx context.Context, limit, offset int, owner string) ([]*Template, error) {
	path := fmt.Sprintf("/templates?limit=%d&offset=%d", limit, offset)
	if owner != "" {
		path += fmt.Sprintf("&owner=%s", owner)
	}

	var respData []*Template
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}
