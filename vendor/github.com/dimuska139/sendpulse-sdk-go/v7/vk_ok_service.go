package sendpulse_sdk_go

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type VkOkService struct {
	client *Client
}

func newVkOkService(cl *Client) *VkOkService {
	return &VkOkService{client: cl}
}

type CreateVkOkSenderParams struct {
	Name        string
	VkUrl       string
	OkUrl       string
	CoverLetter *os.File
}

func (service *VkOkService) CreateSender(ctx context.Context, params CreateVkOkSenderParams) (int, error) {
	path := "/vk-ok/senders"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if params.CoverLetter != nil {
		part, err := writer.CreateFormFile("cover_letter", filepath.Base(path))
		if err != nil {
			return 0, err
		}
		_, err = io.Copy(part, params.CoverLetter)
		if err != nil {
			return 0, err
		}
	}
	_ = writer.WriteField("name", params.Name)
	if params.VkUrl != "" {
		_ = writer.WriteField("vk_url", params.VkUrl)
	}

	if params.OkUrl != "" {
		_ = writer.WriteField("ok_url", params.OkUrl)
	}

	if err := writer.Close(); err != nil {
		return 0, err
	}

	var respData struct {
		ID int `json:"id"`
	}
	_, err := service.client.newFormDataRequest(ctx, path, body, writer.FormDataContentType(), &respData, true)
	return respData.ID, err
}

type CreateVkOkTemplateParams struct {
	Name      string `json:"name"`
	VkMessage string `json:"vk_message,omitempty"`
	OkMessage string `json:"ok_message,omitempty"`
	SenderID  int    `json:"sender_id"`
}

func (service *VkOkService) CreateTemplate(ctx context.Context, params CreateVkOkTemplateParams) (int, error) {
	path := "/vk-ok/templates"

	var respData struct {
		Total int `json:"total"`
		Data  struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.Data.ID, err
}

type VkOkTemplate struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	SenderID  int    `json:"sender_id"`
	Name      string `json:"name"`
	VkMessage string `json:"vk_message"`
	OkMessage string `json:"ok_message"`
	Sender    struct {
		ID        int    `json:"id"`
		UserID    int    `json:"user_id"`
		Name      string `json:"name"`
		VkUrl     string `json:"vk_url"`
		OkUrl     string `json:"ok_url"`
		CreatedAt string `json:"created_at"`
		UpdateAt  string `json:"update_at"`
	} `json:"sender"`
	Status       int `json:"status"`
	StatusDetail struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"status_detail"`
}

func (service *VkOkService) GetTemplates(ctx context.Context) ([]*VkOkTemplate, error) {
	path := "/vk-ok/templates"

	var respData struct {
		Total int             `json:"total"`
		Data  []*VkOkTemplate `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *VkOkService) GetTemplate(ctx context.Context, templateID int) (*VkOkTemplate, error) {
	path := fmt.Sprintf("/vk-ok/templates/%d", templateID)

	var respData struct {
		Total int           `json:"total"`
		Data  *VkOkTemplate `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type SendVkOkTemplateParams struct {
	AddressBooks []int `json:"address_book"`
	Recipients   []struct {
		Phone     string                 `json:"phone"`
		Variables map[string]interface{} `json:"variables"`
	} `json:"recipients"`
	LifeTime   int             `json:"life_time"`
	LifeType   string          `json:"life_type"`
	Name       string          `json:"name"`
	Routes     map[string]bool `json:"routes"`
	SendDate   DateTimeType    `json:"send_date"`
	TemplateID int             `json:"template_id"`
}

func (service *VkOkService) Send(ctx context.Context, params SendVkOkTemplateParams) (int, error) {
	path := "/vk-ok/campaigns"

	var respData struct {
		Total int `json:"total"`
		Data  struct {
			ID int `json:"id"`
		}
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.Data.ID, err
}

type VkOkCurrency struct {
	ID   int    `json:"id"`
	Name string `json:"currency_name"`
	Abbr string `json:"currency_abbr"`
	Sign string `json:"currency_sign"`
}

type VkOkCampaignStatistics struct {
	ID           int          `json:"id"`
	UserID       int          `json:"user_id"`
	Name         string       `json:"name"`
	TotalPrice   int          `json:"total_price"`
	PriceRate    int          `json:"price_rate"`
	Currency     VkOkCurrency `json:"currency"`
	LifeTime     int          `json:"life_time"`
	LifeType     string       `json:"life_type"`
	SendDate     string       `json:"send_date"`
	CreatedAt    string       `json:"created_at"`
	Template     VkOkTemplate `json:"template"`
	Status       int          `json:"status"`
	StatusDetail struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"status_detail"`
	GroupStat []struct {
		ID           int `json:"id"`
		UserID       int `json:"user_id"`
		Sent         int `json:"sent"`
		Delivered    int `json:"delivered"`
		NotDelivered int `json:"not_delivered"`
		Opened       int `json:"opened"`
	} `json:"group_stat"`
}

func (service *VkOkService) GetCampaignsStatistics(ctx context.Context) ([]*VkOkCampaignStatistics, error) {
	path := "/vk-ok/campaigns"

	var respData struct {
		Total int                       `json:"total"`
		Data  []*VkOkCampaignStatistics `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *VkOkService) GetCampaignStatistics(ctx context.Context, campaignID int) (*VkOkCampaignStatistics, error) {
	path := fmt.Sprintf("/vk-ok/campaigns/%d", campaignID)

	var respData *VkOkCampaignStatistics

	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

type VkOkCampaignPhone struct {
	ID           int          `json:"id"`
	UserID       int          `json:"user_id"`
	CampaignID   int          `json:"campaign_id"`
	TemplateID   int          `json:"template_id"`
	Phone        int          `json:"phone"`
	PhoneCost    int          `json:"phone_cost"`
	CurrencyID   int          `json:"currency_id"`
	PriceRate    int          `json:"price_rate"`
	Currency     VkOkCurrency `json:"currency"`
	CreatedAt    string       `json:"created_at"`
	Status       int          `json:"status"`
	StatusDetail struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"status_detail"`
}

func (service *VkOkService) GetCampaignPhones(ctx context.Context, campaignID int) ([]*VkOkCampaignPhone, error) {
	path := fmt.Sprintf("/vk-ok/campaigns/%d/phones", campaignID)

	var respData struct {
		Total int                  `json:"total"`
		Data  []*VkOkCampaignPhone `json:"data"`
	}

	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}
