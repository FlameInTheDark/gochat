package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
)

// WebhooksService is a service to interact with webhooks
type WebhooksService struct {
	client *Client
}

// newWebhooksService creates WebhooksService
func newWebhooksService(cl *Client) *WebhooksService {
	return &WebhooksService{client: cl}
}

// Webhook describes webhook information
type Webhook struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Url    string `json:"url"`
	Action string `json:"action"`
}

// GetWebhooks returns a list of webhooks
func (service *WebhooksService) GetWebhooks(ctx context.Context) ([]*Webhook, error) {
	path := "/v2/email-service/webhook"

	var respData struct {
		Success bool       `json:"success"`
		Data    []*Webhook `json:"data"`
	}

	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// GetWebhook returns specific webhook
func (service *WebhooksService) GetWebhook(ctx context.Context, id int) (*Webhook, error) {
	path := fmt.Sprintf("/v2/email-service/webhook/%d", id)

	var respData struct {
		Success bool     `json:"success"`
		Data    *Webhook `json:"data"`
	}

	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// CreateWebhook creates webhook
func (service *WebhooksService) CreateWebhook(ctx context.Context, actions []string, url string) ([]*Webhook, error) {
	path := "/v2/email-service/webhook/"

	type data struct {
		Actions []string `json:"actions"`
		Url     string   `json:"url"`
	}

	var respData struct {
		Success bool       `json:"success"`
		Data    []*Webhook `json:"data"`
	}
	params := data{Actions: actions, Url: url}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.Data, err
}

// UpdateWebhook updates a specific webhook
func (service *WebhooksService) UpdateWebhook(ctx context.Context, id int, url string) error {
	path := fmt.Sprintf("/v2/email-service/webhook/%d", id)

	type data struct {
		Url string `json:"url"`
	}

	var respData struct {
		Success bool   `json:"success"`
		Data    []bool `json:"data"`
	}
	params := data{Url: url}
	_, err := service.client.newRequest(ctx, http.MethodPut, path, params, &respData, true)
	return err
}

// DeleteWebhook deletes a specific webhook
func (service *WebhooksService) DeleteWebhook(ctx context.Context, id int) error {
	path := fmt.Sprintf("/v2/email-service/webhook/%d", id)

	var respData struct {
		Success bool   `json:"success"`
		Data    []bool `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, nil, &respData, true)
	return err
}
