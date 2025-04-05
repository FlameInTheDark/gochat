package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PushService is a service to interact with push notifications
type PushService struct {
	client *Client
}

// newPushService creates PushService
func newPushService(cl *Client) *PushService {
	return &PushService{client: cl}
}

// PushListParams describes params for GetMessages
type PushListParams struct {
	Limit     int
	Offset    int
	From      time.Time
	To        time.Time
	WebsiteID int
}

// Push represents information of push notification
type Push struct {
	ID        int          `json:"id"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	WebsiteID int          `json:"website_id"`
	From      DateTimeType `json:"from"`
	To        DateTimeType `json:"to"`
	Status    int          `json:"status"`
}

// GetMessages retrieves a list of sent web push campaigns
func (service *PushService) GetMessages(ctx context.Context, params PushListParams) ([]Push, error) {
	path := "/push/tasks/"
	var urlParts []string
	urlParts = append(urlParts, fmt.Sprintf("offset=%d", params.Offset))
	if params.Limit != 0 {
		urlParts = append(urlParts, fmt.Sprintf("limit=%d", params.Limit))
	}
	if !params.From.IsZero() {
		urlParts = append(urlParts, fmt.Sprintf("from=%s", params.From.Format("2006-01-02")))
	}
	if !params.To.IsZero() {
		urlParts = append(urlParts, fmt.Sprintf("to=%s", params.From.Format("2006-01-02")))
	}
	if params.WebsiteID != 0 {
		urlParts = append(urlParts, fmt.Sprintf("website_id=%d", params.WebsiteID))
	}

	if len(urlParts) != 0 {
		path += "?" + strings.Join(urlParts, "&")
	}

	var respData []Push
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// CountWebsites retrieves the total number of websites
func (service *PushService) CountWebsites(ctx context.Context) (int, error) {
	path := "/push/websites/total"
	var respData struct {
		Total int `json:"total"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Total, err
}

type PushWebsite struct {
	ID      int          `json:"id"`
	Url     string       `json:"url"`
	AddDate DateTimeType `json:"add_date"`
	Status  int          `json:"status"`
}

// GetWebsites retrieves a list of websites
func (service *PushService) GetWebsites(ctx context.Context, limit, offset int) ([]*PushWebsite, error) {
	path := fmt.Sprintf("/push/websites/?limit=%d&offset=%d", limit, offset)
	var respData []*PushWebsite
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// PushWebsiteVariable describes variable of push notification
type PushWebsiteVariable struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetWebsiteVariables returns a list of variables for specific website
func (service *PushService) GetWebsiteVariables(ctx context.Context, websiteID int) ([]*PushWebsiteVariable, error) {
	path := fmt.Sprintf("/push/websites/%d/variables", websiteID)
	var respData []*PushWebsiteVariable
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// WebsiteSubscriptionsParams describes params for GetWebsiteSubscriptions
type WebsiteSubscriptionsParams struct {
	Limit  int
	Offset int
	From   time.Time
	To     time.Time
}

// WebsiteSubscription represents subscriber
type WebsiteSubscription struct {
	ID               int                   `json:"id"`
	Browser          string                `json:"browser"`
	Lang             string                `json:"lang"`
	Os               string                `json:"os"`
	CountryCode      string                `json:"country_code"`
	City             string                `json:"city"`
	Variables        []PushWebsiteVariable `json:"variables"`
	SubscriptionDate DateTimeType          `json:"subscription_date"`
	Status           int                   `json:"status"`
}

// GetWebsiteSubscriptions returns a list subscribers for a certain website
func (service *PushService) GetWebsiteSubscriptions(ctx context.Context, websiteID int, params WebsiteSubscriptionsParams) ([]*WebsiteSubscription, error) {
	path := fmt.Sprintf("/push/websites/%d/subscriptions", websiteID)

	var urlParts []string
	urlParts = append(urlParts, fmt.Sprintf("offset=%d", params.Offset))
	if params.Limit != 0 {
		urlParts = append(urlParts, fmt.Sprintf("limit=%d", params.Limit))
	}
	if !params.From.IsZero() {
		urlParts = append(urlParts, fmt.Sprintf("subscription_date_from=%s", params.From.Format("2006-01-02")))
	}
	if !params.To.IsZero() {
		urlParts = append(urlParts, fmt.Sprintf("subscription_date_to=%s", params.From.Format("2006-01-02")))
	}

	if len(urlParts) != 0 {
		path += "?" + strings.Join(urlParts, "&")
	}

	var respData []*WebsiteSubscription
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// CountWebsiteSubscriptions returns the total number of website subscribers
func (service *PushService) CountWebsiteSubscriptions(ctx context.Context, websiteID int) (int, error) {
	path := fmt.Sprintf("/push/websites/%d/subscriptions/total", websiteID)
	var respData struct {
		Total int `json:"total"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Total, err
}

// WebsiteInfo describes information about website
type WebsiteInfo struct {
	ID                int          `json:"id"`
	Url               string       `json:"url"`
	Status            string       `json:"status"`
	Icon              string       `json:"icon"`
	AddDate           DateTimeType `json:"add_date"`
	TotalSubscribers  int          `json:"total_subscribers"`
	Unsubscribed      int          `json:"unsubscribed"`
	SubscribersToday  int          `json:"subscribers_today"`
	ActiveSubscribers int          `json:"active_subscribers"`
}

// GetWebsiteInfo returns information about specific website
func (service *PushService) GetWebsiteInfo(ctx context.Context, websiteID int) (*WebsiteInfo, error) {
	path := fmt.Sprintf("/push/websites/info/%d", websiteID)
	var respData *WebsiteInfo
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// ActivateSubscription activates a subscriber
func (service *PushService) ActivateSubscription(ctx context.Context, subscriptionID int) error {
	path := "/push/subscriptions/state"
	type paramsFormat struct {
		ID    int `json:"id"`
		State int `json:"state"`
	}

	data := paramsFormat{ID: subscriptionID, State: 1}

	var respData struct {
		Result bool `json:"true"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &respData, true)
	return err
}

// DeactivateSubscription deactivates a subscriber
func (service *PushService) DeactivateSubscription(ctx context.Context, subscriptionID int) error {
	path := "/push/subscriptions/state"
	type paramsFormat struct {
		ID    int `json:"id"`
		State int `json:"state"`
	}

	data := paramsFormat{ID: subscriptionID, State: 0}

	var respData struct {
		Result bool `json:"true"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &respData, true)
	return err
}

// PushMessageParams describes parameters to CreatePushCampaign
type PushMessageParams struct {
	Title                string    `json:"title"`
	WebsiteID            int       `json:"website_id"`
	Body                 string    `json:"body"`
	TtlSec               int       `json:"ttl"`
	Link                 string    `json:"link,omitempty"`
	FilterLang           string    `json:"filter_lang,omitempty"`
	FilterBrowser        string    `json:"filter_browser,omitempty"`
	FilterRegion         string    `json:"filter_region,omitempty"`
	FilterUrl            string    `json:"filter_url,omitempty"`
	SubscriptionDateFrom time.Time `json:"filter_subscription_date_from,omitempty"`
	SubscriptionDateTo   time.Time `json:"filter_subscription_date_to,omitempty"`
	Filter               *struct {
		VariableName string `json:"variable_name"`
		Operator     string `json:"operator"`
		Conditions   []struct {
			Condition string      `json:"condition"`
			Value     interface{} `json:"value"`
		} `json:"conditions"`
	} `json:"filter,omitempty"`
	StretchTimeSec int          `json:"stretch_time"`
	SendDate       DateTimeType `json:"send_date"`
	Buttons        *struct {
		Text string `json:"text"`
		Link string `json:"link"`
	} `json:"buttons,omitempty"`
	Image *struct {
		Name       string `json:"name"`
		DataBase64 string `json:"data"`
	} `json:"image,omitempty"`
	Icon *struct {
		Name       string `json:"name"`
		DataBase64 string `json:"data"`
	} `json:"icon,omitempty"`
}

// CreatePushCampaign creates new push campaign
func (service *PushService) CreatePushCampaign(ctx context.Context, params PushMessageParams) (int, error) {
	path := "/push/tasks"

	var respData struct {
		ID     int  `json:"id"`
		Result bool `json:"true"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.ID, err
}

// PushMessagesStatistics describes statistics on sent campaign
type PushMessagesStatistics struct {
	ID      int `json:"id"`
	Message struct {
		Title string `json:"title"`
		Text  string `json:"text"`
		Link  string `json:"link"`
	}
	Website   string `json:"website"`
	WebsiteID int    `json:"website_id"`
	Status    int    `json:"status"`
	Send      int    `json:"send,string"`
	Delivered int    `json:"delivered"`
	Redirect  int    `json:"redirect"`
}

// GetPushMessagesStatistics returns statistics on sent campaigns
func (service *PushService) GetPushMessagesStatistics(ctx context.Context, taskID int) (*PushMessagesStatistics, error) {
	path := fmt.Sprintf("/push/tasks/%d", taskID)

	var respData *PushMessagesStatistics
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}
