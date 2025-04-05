package sendpulse_sdk_go

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"strconv"
)

// CampaignsService is a service to interact with campaigns
type CampaignsService struct {
	client *Client
}

// newCampaignsService creates CampaignsService
func newCampaignsService(cl *Client) *CampaignsService {
	return &CampaignsService{client: cl}
}

// Campaign describes campaign params
type CampaignParams struct {
	SenderName    string            `json:"sender_name"`
	SenderEmail   string            `json:"sender_email"`
	Subject       string            `json:"subject"`
	Body          string            `json:"body,omitempty"`
	TemplateID    string            `json:"template_id,omitempty"`
	MailingListID int               `json:"list_id,omitempty"`
	SegmentID     int               `json:"segment_id,omitempty"`
	IsTest        bool              `json:"is_test,omitempty"`
	SendDate      DateTimeType      `json:"send_date,omitempty"`
	Name          string            `json:"name,omitempty"`
	Attachments   map[string]string `json:"attachments"`
	Type          string            `json:"type,omitempty"`
	BodyAMP       string            `json:"body_amp,omitempty"`
}

// Campaign describes a campaign
type Campaign struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message struct {
		SenderName    string `json:"sender_name"`
		SenderEmail   string `json:"sender_email"`
		Subject       string `json:"subject"`
		Body          string `json:"body"`
		Attachments   string `json:"attachments"`
		MailingListID int    `json:"list_id"`
	}
	Status            int          `json:"status"`
	AllEmailQty       int          `json:"all_email_qty"`
	TariffEmailQty    int          `json:"tariff_email_qty"`
	PaidEmailQty      int          `json:"paid_email_qty"`
	OverdraftPrice    float32      `json:"overdraft_price"`
	OverdraftCurrency string       `json:"overdraft_currency"`
	SendDate          DateTimeType `json:"send_date"`
}

// CreateCampaign creates a campaign. Please note that you can send a maximum of 4 campaigns per hour
func (service *CampaignsService) CreateCampaign(ctx context.Context, data CampaignParams) (*Campaign, error) {
	path := "/campaigns"
	var innerMailing struct {
		Campaign
		OverdraftPrice string `json:"overdraft_price"`
	}

	if data.Body != "" {
		data.Body = b64.StdEncoding.EncodeToString([]byte(data.Body))
	}

	if data.BodyAMP != "" {
		data.BodyAMP = b64.StdEncoding.EncodeToString([]byte(data.BodyAMP))
	}

	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &innerMailing, true)
	if err != nil {
		return nil, err
	}

	f64, _ := strconv.ParseFloat(innerMailing.OverdraftPrice, 32)

	innerMailing.Campaign.OverdraftPrice = float32(f64)

	return &innerMailing.Campaign, err
}

// UpdateCampaign updates a scheduled campaign
func (service *CampaignsService) UpdateCampaign(ctx context.Context, id int, data CampaignParams) error {
	path := fmt.Sprintf("/campaigns/%d", id)
	var respData struct {
		Result bool `json:"result"`
		Id     int  `json:"id"`
	}

	if data.Body != "" {
		data.Body = b64.StdEncoding.EncodeToString([]byte(data.Body))
	}

	if data.BodyAMP != "" {
		data.BodyAMP = b64.StdEncoding.EncodeToString([]byte(data.BodyAMP))
	}

	_, err := service.client.newRequest(ctx, http.MethodPatch, path, data, &respData, true)
	return err
}

// GetCampaign returns an information about specific campaign
func (service *CampaignsService) GetCampaign(ctx context.Context, id int) (*Campaign, error) {
	path := fmt.Sprintf("/campaigns/%d", id)
	var respData Campaign
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return &respData, err
}

// GetCampaigns returns a list of campaigns
func (service *CampaignsService) GetCampaigns(ctx context.Context, limit int, offset int) ([]*Campaign, error) {
	path := fmt.Sprintf("/campaigns?limit=%d&offset=%d", limit, offset)
	var items []*Campaign
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &items, true)
	return items, err
}

// Task represents a campaign
type Task struct {
	ID     int    `json:"task_id"`
	Name   string `json:"task_name"`
	Status int    `json:"task_status"`
}

// GetCampaignsByMailingList returns a list of campaigns by specific mailing list
func (service *CampaignsService) GetCampaignsByMailingList(ctx context.Context, mailingListID, limit, offset int) ([]*Task, error) {
	path := fmt.Sprintf("/addressbooks/%d/campaigns?limit=%d&offset=%d", mailingListID, limit, offset)
	var tasks []*Task
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &tasks, true)
	return tasks, err
}

// GetCampaignCountriesStatistics represents campaign statistics of countries
func (service *CampaignsService) GetCampaignCountriesStatistics(ctx context.Context, id int) (map[string]int, error) {
	path := fmt.Sprintf("/campaigns/%d/countries", id)
	response := make(map[string]int)
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response, err
}

// MailingRefStat represents campaign statistics of referrals
type MailingRefStat struct {
	Link  string `json:"link"`
	Count int    `json:"count"`
}

// GetCampaignReferralsStatistics returns campaign statistics of referrals
func (service *CampaignsService) GetCampaignReferralsStatistics(ctx context.Context, id int) ([]*MailingRefStat, error) {
	path := fmt.Sprintf("/campaigns/%d/referrals", id)
	var response []*MailingRefStat
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response, err
}

// CancelCampaign cancels a scheduled campaign
func (service *CampaignsService) CancelCampaign(ctx context.Context, id int) error {
	path := fmt.Sprintf("/campaigns/%d", id)
	var response struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, nil, &response, true)
	return err
}
