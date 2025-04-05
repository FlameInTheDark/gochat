package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
)

type ViberService struct {
	client *Client
}

func newViberService(cl *Client) *ViberService {
	return &ViberService{client: cl}
}

type CreateViberCampaignParams struct {
	TaskName        string       `json:"task_name"`
	MessageType     int          `json:"message_type,omitempty"`
	SenderID        int          `json:"sender_id"`
	MessageLiveTime int          `json:"message_live_time"`
	SendDate        DateTimeType `json:"send_date"`
	MailingListID   int          `json:"address_book"`
	Recipients      []int        `json:"recipients"`
	Message         string       `json:"message"`
	Additional      *struct {
		Button *struct {
			Text string `json:"text"`
			Link string `json:"link"`
		} `json:"button,omitempty"`
		Image *struct {
			Link string `json:"link"`
		} `json:"image,omitempty"`
		ResendSms *struct {
			Status        bool   `json:"status"`
			SmsText       string `json:"sms_text"`
			SmsSenderName string `json:"sms_sender_name"`
		} `json:"resend_sms,omitempty"`
	} `json:"additional,omitempty"`
}

func (service *ViberService) CreateCampaign(ctx context.Context, params CreateViberCampaignParams) (int, error) {
	path := "/viber"

	var respData struct {
		Result bool `json:"result"`
		Data   struct {
			TaskID int `json:"task_id"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.Data.TaskID, err
}

type UpdateViberCampaignParams struct {
	TaskID          int          `json:"main_task_id"`
	TaskName        string       `json:"task_name"`
	Message         string       `json:"message"`
	MessageType     int          `json:"message_type"`
	ButtonText      string       `json:"button_text,omitempty"`
	ButtonLink      string       `json:"button_link,omitempty"`
	ImageLink       string       `json:"image_link,omitempty"`
	AddressBookID   int          `json:"address_book,omitempty"`
	SenderID        int          `json:"sender_id"`
	MessageLiveTime int          `json:"message_live_time"`
	SendDate        DateTimeType `json:"send_date"`
}

func (service *ViberService) UpdateCampaign(ctx context.Context, params UpdateViberCampaignParams) error {
	path := "/viber/update"

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}

type ViberCampaign struct {
	ID              int          `json:"id"`
	Name            string       `json:"name"`
	Message         string       `json:"message"`
	ButtonText      string       `json:"button_text"`
	ButtonLink      string       `json:"button_link"`
	ImageLink       string       `json:"image_link"`
	AddressBookID   int          `json:"address_book"`
	SenderName      string       `json:"sender_name"`
	SenderID        int          `json:"sender_id"`
	MessageLiveTime int          `json:"message_live_time"`
	SendDate        DateTimeType `json:"send_date"`
	Status          string       `json:"status"`
	Created         DateTimeType `json:"created"`
}

func (service *ViberService) GetCampaigns(ctx context.Context, limit, offset int) ([]*ViberCampaign, error) {
	path := fmt.Sprintf("/viber/task?limit=%d&offset=%d", limit, offset)

	var respData []*ViberCampaign
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

type ViberCampaignStatistics struct {
	ID            int          `json:"id"`
	Name          string       `json:"name"`
	Message       string       `json:"message"`
	ButtonText    string       `json:"button_text"`
	ButtonLink    string       `json:"button_link"`
	ImageLink     string       `json:"image_link"`
	AddressBookID int          `json:"address_book"`
	SenderName    string       `json:"sender_name"`
	SendDate      DateTimeType `json:"send_date"`
	Status        string       `json:"status"`
	Statistics    struct {
		Sent        int `json:"sent"`
		Delivered   int `json:"delivered"`
		Read        int `json:"read"`
		Redirected  int `json:"redirected"`
		Undelivered int `json:"undelivered"`
		Errors      int `json:"errors"`
	} `json:"statistic"`
	Created DateTimeType `json:"created"`
}

func (service *ViberService) GetStatistics(ctx context.Context, campaignID int) (*ViberCampaignStatistics, error) {
	path := fmt.Sprintf("/viber/task/%d", campaignID)

	var respData *ViberCampaignStatistics
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

type ViberSender struct {
	ID           int      `json:"id"`
	Status       string   `json:"status"`
	Name         string   `json:"name"`
	ServiceType  string   `json:"service_type"`
	WebSite      string   `json:"web_site"`
	Description  string   `json:"description"`
	Countries    []string `json:"countries"`
	TrafficType  string   `json:"traffic_type"`
	AdminComment string   `json:"admin_comment"`
	Owner        string   `json:"owner"`
}

func (service *ViberService) GetSenders(ctx context.Context) ([]*ViberSender, error) {
	path := "/viber/senders"

	var respData []*ViberSender
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

func (service *ViberService) GetSender(ctx context.Context, senderID int) (*ViberSender, error) {
	path := fmt.Sprintf("/viber/senders/%d", senderID)

	var respData *ViberSender
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

type ViberRecipient struct {
	Phone         int          `json:"phone"`
	AddressBookID int          `json:"address_book_id"`
	Status        string       `json:"status"`
	SendDate      DateTimeType `json:"send_date"`
	Price         float32      `json:"price"`
	Currency      string       `json:"currency"`
	LastUpdate    DateTimeType `json:"last_update"`
}

func (service *ViberService) GetRecipients(ctx context.Context, taskID int) ([]*ViberRecipient, error) {
	path := fmt.Sprintf("/viber/task/%d/recipients", taskID)

	var respData struct {
		TaskID     int               `json:"task_id"`
		Recipients []*ViberRecipient `json:"recipients"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Recipients, err
}
