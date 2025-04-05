package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BotsFbService struct {
	client *Client
}

func newBotsFbService(cl *Client) *BotsFbService {
	return &BotsFbService{client: cl}
}

type FbAccount struct {
	Tariff struct {
		Branding     bool      `json:"branding"`
		MaxBots      int       `json:"max_bots"`
		MaxContacts  int       `json:"max_contacts"`
		MaxMessages  int       `json:"max_messages"`
		MaxTags      int       `json:"max_tags"`
		MaxVariables int       `json:"max_variables"`
		MaxRss       int       `json:"max_rss"`
		Code         string    `json:"code"`
		IsExceeded   bool      `json:"is_exceeded"`
		IsExpired    bool      `json:"is_expired"`
		ExpiredAt    time.Time `json:"expired_at"`
	} `json:"tariff"`
	Statistics struct {
		Messages  int `json:"messages"`
		Bots      int `json:"bots"`
		Contacts  int `json:"contacts"`
		Variables int `json:"variables"`
	} `json:"statistics"`
}

func (service *BotsFbService) GetAccount(ctx context.Context) (*FbAccount, error) {
	path := "/messenger/account"

	var respData struct {
		Success bool       `json:"success"`
		Data    *FbAccount `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type FbBot struct {
	ID          string `json:"id"`
	ChannelData struct {
		ID          string `json:"id"`
		AccessToken string `json:"access_token"`
		Name        string `json:"name"`
		Photo       string `json:"photo"`
	} `json:"channel_data"`
	Inbox struct {
		Total  int `json:"total"`
		Unread int `json:"unread"`
	} `json:"inbox"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (service *BotsFbService) GetBots(ctx context.Context) ([]*FbBot, error) {
	path := "/messenger/bots"

	var respData struct {
		Success bool     `json:"success"`
		Data    []*FbBot `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type FbBotContact struct {
	ID          string `json:"id"`
	BotID       string `json:"bot_id"`
	Status      int    `json:"status"`
	ChannelData struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		ProfilePic string `json:"profile_pic"`
		Locale     string `json:"locale"`
		Gender     string `json:"gender"`
	} `json:"channel_data"`
	Tags                  []string               `json:"tags"`
	Variables             map[string]interface{} `json:"variables"`
	IsChatOpened          bool                   `json:"is_chat_opened"`
	LastActivityAt        time.Time              `json:"last_activity_at"`
	AutomationPausedUntil time.Time              `json:"automation_paused_until"`
	UnsubscribedAt        time.Time              `json:"unsubscribed_at"`
	CreatedAt             time.Time              `json:"created_at"`
}

func (service *BotsFbService) GetContact(ctx context.Context, contactID string) (*FbBotContact, error) {
	path := fmt.Sprintf("/messenger/contacts/get?id=%s", contactID)

	var respData struct {
		Success bool          `json:"success"`
		Data    *FbBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsFbService) GetContactsByTag(ctx context.Context, tag, botID string) ([]*FbBotContact, error) {
	path := fmt.Sprintf("/messenger/contacts/getByTag?tag=%s&bot_id=%s", tag, botID)

	var respData struct {
		Success bool            `json:"success"`
		Data    []*FbBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type BotContactsByVariableParams struct {
	VariableID    string
	VariableName  string
	BotID         string
	VariableValue string
}

func (service *BotsFbService) GetContactsByVariable(ctx context.Context, params BotContactsByVariableParams) ([]*FbBotContact, error) {
	urlParams := url.Values{}
	urlParams.Add("variable_value", params.VariableValue)
	if params.VariableID != "" {
		urlParams.Add("variable_id", params.VariableID)
	}
	if params.VariableName != "" {
		urlParams.Add("variable_name", params.VariableName)
	}
	if params.BotID != "" {
		urlParams.Add("bot_id", params.BotID)
	}
	path := "/messenger/contacts/getByVariable?" + urlParams.Encode()

	var respData struct {
		Success bool            `json:"success"`
		Data    []*FbBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type FbBotSendTextParams struct {
	ContactID   string `json:"contact_id"`
	MessageType string `json:"message_type"`
	MessageTag  string `json:"message_tag"`
	Text        string `json:"text"`
}

func (service *BotsFbService) SendTextByContact(ctx context.Context, params FbBotSendTextParams) error {
	path := "/messenger/contacts/sendText"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}

func (service *BotsFbService) SetVariableToContact(ctx context.Context, contactID string, variableID string, variableName string, variableValue interface{}) error {
	path := "/messenger/contacts/setVariable"

	type bodyFormat struct {
		ContactID     string      `json:"contact_id"`
		VariableID    string      `json:"variable_id"`
		VariableName  string      `json:"variable_name"`
		VariableValue interface{} `json:"variable_value"`
	}
	body := bodyFormat{
		ContactID:     contactID,
		VariableID:    variableID,
		VariableName:  variableName,
		VariableValue: variableValue,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) SetTagsToContact(ctx context.Context, contactID string, tags []string) error {
	path := "/messenger/contacts/setTag"

	type bodyFormat struct {
		ContactID string   `json:"contact_id"`
		Tags      []string `json:"tags"`
	}
	body := bodyFormat{
		ContactID: contactID,
		Tags:      tags,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) DeleteTagFromContact(ctx context.Context, contactID string, tag string) error {
	path := "/messenger/contacts/deleteTag"

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
		Tag       string `json:"tag"`
	}
	body := bodyFormat{
		ContactID: contactID,
		Tag:       tag,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) DisableContact(ctx context.Context, contactID string) error {
	path := "/messenger/contacts/disable"

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
	}
	body := bodyFormat{
		ContactID: contactID,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) EnableContact(ctx context.Context, contactID string) error {
	path := "/messenger/contacts/enable"

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
	}
	body := bodyFormat{
		ContactID: contactID,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) DeleteContact(ctx context.Context, contactID string) error {
	path := "/messenger/contacts/delete"

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
	}
	body := bodyFormat{
		ContactID: contactID,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) GetPauseAutomation(ctx context.Context, contactID string) (int, error) {
	path := fmt.Sprintf("/messenger/contacts/getPauseAutomation?contact_id=%s", contactID)

	var respData struct {
		Success bool `json:"success"`
		Data    struct {
			Minutes int `json:"minutes"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data.Minutes, err
}

func (service *BotsFbService) SetPauseAutomation(ctx context.Context, contactID string, minutes int) error {
	path := "/messenger/contacts/setPauseAutomation"
	type bodyFormat struct {
		ContactID string `json:"contact_id"`
		Minutes   int    `json:"minutes"`
	}
	body := bodyFormat{
		ContactID: contactID,
		Minutes:   minutes,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) DeletePauseAutomation(ctx context.Context, contactID string) error {
	path := "/messenger/contacts/deletePauseAutomation"
	type bodyFormat struct {
		ContactID string `json:"contact_id"`
	}
	body := bodyFormat{
		ContactID: contactID,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

type BotVariable struct {
	ID          string    `json:"id"`
	BotID       string    `json:"bot_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        int       `json:"type"`
	ValueType   int       `json:"value_type"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (service *BotsFbService) GetBotVariables(ctx context.Context, botID string) ([]*BotVariable, error) {
	path := fmt.Sprintf("/messenger/variables?bot_id=%s", botID)

	var respData struct {
		Success bool           `json:"success"`
		Data    []*BotVariable `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type BotFlow struct {
	ID       string `json:"id"`
	BotID    string `json:"bot_id"`
	Name     string `json:"name"`
	Status   int    `json:"status"`
	Triggers []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"triggers"`
	CreatedAt time.Time `json:"created_at"`
}

func (service *BotsFbService) GetFlows(ctx context.Context, botID string) ([]*BotFlow, error) {
	path := fmt.Sprintf("/messenger/flows?bot_id=%s", botID)

	var respData struct {
		Success bool       `json:"success"`
		Data    []*BotFlow `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsFbService) RunFlow(ctx context.Context, contactID, flowID string, externalData map[string]interface{}) error {
	path := "/messenger/flows/run"

	type bodyFormat struct {
		ContactID    string                 `json:"contact_id"`
		FlowID       string                 `json:"flow_id"`
		ExternalData map[string]interface{} `json:"external_data,omitempty"`
	}
	body := bodyFormat{
		ContactID:    contactID,
		FlowID:       flowID,
		ExternalData: externalData,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsFbService) RunFlowByTrigger(ctx context.Context, contactID, triggerKeyword string, externalData map[string]interface{}) error {
	path := "/messenger/flows/runByTrigger"

	type bodyFormat struct {
		ContactID      string                 `json:"contact_id"`
		TriggerKeyword string                 `json:"trigger_keyword"`
		ExternalData   map[string]interface{} `json:"external_data,omitempty"`
	}
	body := bodyFormat{
		ContactID:      contactID,
		TriggerKeyword: triggerKeyword,
		ExternalData:   externalData,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

type BotTrigger struct {
	ID        string   `json:"id"`
	BotID     string   `json:"bot_id"`
	FlowID    string   `json:"flow_id"`
	Name      string   `json:"name"`
	Type      int      `json:"type"`
	Status    int      `json:"status"`
	Keywords  []string `json:"keywords"`
	Execution struct {
		Interval int `json:"interval"`
		Units    int `json:"units"`
	} `json:"execution"`
	CreatedAt time.Time `json:"created_at"`
}

func (service *BotsFbService) GetBotTriggers(ctx context.Context, botID string) ([]*BotTrigger, error) {
	path := fmt.Sprintf("/messenger/triggers?bot_id=%s", botID)

	var respData struct {
		Success bool          `json:"success"`
		Data    []*BotTrigger `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type FbBotMessage struct {
	ID           string                 `json:"id"`
	ContactID    string                 `json:"contact_id"`
	BotID        string                 `json:"bot_id"`
	CampaignID   string                 `json:"campaign_id"`
	Data         map[string]interface{} `json:"data"`
	Direction    int                    `json:"direction"`
	Status       int                    `json:"status"`
	DeliveredAt  time.Time              `json:"delivered_at"`
	OpenedAt     time.Time              `json:"opened_at"`
	RedirectedAt time.Time              `json:"redirected_at"`
	CreatedAt    time.Time              `json:"created_at"`
}

type FbBotChat struct {
	Contact          *FbBotContact `json:"contact"`
	InboxLastMessage *FbBotMessage `json:"inbox_last_message"`
	InboxUnread      int           `json:"inbox_unread"`
}

func (service *BotsFbService) GetBotChats(ctx context.Context, botID string) ([]*FbBotChat, error) {
	path := fmt.Sprintf("/messenger/chats?bot_id=%s", botID)

	var respData struct {
		Success bool         `json:"success"`
		Data    []*FbBotChat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsFbService) GetContactMessages(ctx context.Context, contactID string) ([]*FbBotMessage, error) {
	path := fmt.Sprintf("/messenger/chats/messages?contact_id=%s", contactID)

	var respData struct {
		Success bool            `json:"success"`
		Data    []*FbBotMessage `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type FbBotSendCampaignParams struct {
	Title                   string                 `json:"title"`
	BotID                   string                 `json:"bot_id"`
	MessageTag              string                 `json:"message_tag"`
	MessageNotificationType string                 `json:"message_notification_type"`
	SendAt                  time.Time              `json:"send_at"`
	Messages                []FbBotCampaignMessage `json:"messages"`
}

type FbBotCampaignMessage struct {
	Type string `json:"type"`
	Data struct {
		Text string `json:"text"`
	} `json:"data"`
}

func (service *BotsFbService) SendCampaign(ctx context.Context, params FbBotSendCampaignParams) error {
	path := "/messenger/campaigns/send"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}
