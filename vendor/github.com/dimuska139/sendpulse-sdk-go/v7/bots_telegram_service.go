package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BotsTelegramService struct {
	client *Client
}

func newBotsTelegramService(cl *Client) *BotsTelegramService {
	return &BotsTelegramService{client: cl}
}

type TelegramAccount struct {
	Plan struct {
		Code         string    `json:"code"`
		MaxBots      int       `json:"max_bots"`
		MaxContacts  int       `json:"max_contacts"`
		MaxMessages  int       `json:"max_messages"`
		MaxTags      int       `json:"max_tags"`
		MaxVariables int       `json:"max_variables"`
		Branding     bool      `json:"branding"`
		IsExceeded   bool      `json:"is_exceeded"`
		IsExpired    bool      `json:"is_expired"`
		ExpiredAt    time.Time `json:"expired_at"`
	} `json:"plan"`
	Statistics struct {
		Messages  int `json:"messages"`
		Bots      int `json:"bots"`
		Contacts  int `json:"contacts"`
		Variables int `json:"variables"`
	} `json:"statistics"`
}

func (service *BotsTelegramService) GetAccount(ctx context.Context) (*TelegramAccount, error) {
	path := "/telegram/account"

	var respData struct {
		Success bool             `json:"success"`
		Data    *TelegramAccount `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type TelegramBot struct {
	ID          string `json:"id"`
	ChannelData struct {
		ID          string `json:"id"`
		AccessToken string `json:"access_token"`
		Name        string `json:"name"`
		Username    string `json:"username"`
	} `json:"channel_data"`
	Inbox struct {
		Total  int `json:"total"`
		Unread int `json:"unread"`
	} `json:"inbox"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CommandsMenu struct {
		Status   int `json:"status"`
		Commands []struct {
			Description string `json:"description"`
			Command     string `json:"command"`
			FlowID      string `json:"flow_id"`
		}
	} `json:"commands_menu"`
}

func (service *BotsTelegramService) GetBots(ctx context.Context) ([]*TelegramBot, error) {
	path := "/telegram/bots"

	var respData struct {
		Success bool           `json:"success"`
		Data    []*TelegramBot `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type TelegramBotContact struct {
	ID          string `json:"id"`
	BotID       string `json:"bot_id"`
	Status      int    `json:"status"`
	ChannelData struct {
		Username     string `json:"username"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Name         string `json:"name"`
		LanguageCode string `json:"language_code"`
	} `json:"channel_data"`
	Tags                  []string               `json:"tags"`
	Variables             map[string]interface{} `json:"variables"`
	IsChatOpened          bool                   `json:"is_chat_opened"`
	LastActivityAt        time.Time              `json:"last_activity_at"`
	AutomationPausedUntil time.Time              `json:"automation_paused_until"`
	CreatedAt             time.Time              `json:"created_at"`
}

func (service *BotsTelegramService) GetContact(ctx context.Context, contactID string) (*TelegramBotContact, error) {
	path := fmt.Sprintf("/telegram/contacts/get?id=%s", contactID)

	var respData struct {
		Success bool                `json:"success"`
		Data    *TelegramBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsTelegramService) GetContactsByTag(ctx context.Context, tag, botID string) ([]*TelegramBotContact, error) {
	path := fmt.Sprintf("/telegram/contacts/getByTag?tag=%s&bot_id=%s", tag, botID)

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*TelegramBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsTelegramService) GetContactsByVariable(ctx context.Context, params BotContactsByVariableParams) ([]*TelegramBotContact, error) {
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
	path := "/telegram/contacts/getByVariable?" + urlParams.Encode()

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*TelegramBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsTelegramService) SendTextByContact(ctx context.Context, contactID string, text string) error {
	path := "/telegram/contacts/sendText"

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
		Text      string `json:"text"`
	}
	body := bodyFormat{
		ContactID: contactID,
		Text:      text,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsTelegramService) SetVariableToContact(ctx context.Context, contactID string, variableID string, variableName string, variableValue interface{}) error {
	path := "/telegram/contacts/setVariable"

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

func (service *BotsTelegramService) SetTagsToContact(ctx context.Context, contactID string, tags []string) error {
	path := "/telegram/contacts/setTag"

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

func (service *BotsTelegramService) DeleteTagFromContact(ctx context.Context, contactID string, tag string) error {
	path := "/telegram/contacts/deleteTag"

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

func (service *BotsTelegramService) DisableContact(ctx context.Context, contactID string) error {
	path := "/telegram/contacts/disable"

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

func (service *BotsTelegramService) EnableContact(ctx context.Context, contactID string) error {
	path := "/telegram/contacts/enable"

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

func (service *BotsTelegramService) DeleteContact(ctx context.Context, contactID string) error {
	path := "/telegram/contacts/delete"

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

func (service *BotsTelegramService) GetPauseAutomation(ctx context.Context, contactID string) (int, error) {
	path := fmt.Sprintf("/telegram/contacts/getPauseAutomation?contact_id=%s", contactID)

	var respData struct {
		Success bool `json:"success"`
		Data    struct {
			Minutes int `json:"minutes"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data.Minutes, err
}

func (service *BotsTelegramService) SetPauseAutomation(ctx context.Context, contactID string, minutes int) error {
	path := "/telegram/contacts/setPauseAutomation"
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

func (service *BotsTelegramService) DeletePauseAutomation(ctx context.Context, contactID string) error {
	path := "/telegram/contacts/deletePauseAutomation"
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

func (service *BotsTelegramService) GetBotVariables(ctx context.Context, botID string) ([]*BotVariable, error) {
	path := fmt.Sprintf("/telegram/variables?bot_id=%s", botID)

	var respData struct {
		Success bool           `json:"success"`
		Data    []*BotVariable `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsTelegramService) GetFlows(ctx context.Context, botID string) ([]*BotFlow, error) {
	path := fmt.Sprintf("/telegram/flows?bot_id=%s", botID)

	var respData struct {
		Success bool       `json:"success"`
		Data    []*BotFlow `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsTelegramService) RunFlow(ctx context.Context, contactID, flowID string, externalData map[string]interface{}) error {
	path := "/telegram/flows/run"

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

func (service *BotsTelegramService) RunFlowByTrigger(ctx context.Context, contactID, triggerKeyword string, externalData map[string]interface{}) error {
	path := "/telegram/flows/runByTrigger"

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

func (service *BotsTelegramService) GetBotTriggers(ctx context.Context, botID string) ([]*BotTrigger, error) {
	path := fmt.Sprintf("/telegram/triggers?bot_id=%s", botID)

	var respData struct {
		Success bool          `json:"success"`
		Data    []*BotTrigger `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type TelegramBotMessage struct {
	ID         string                 `json:"id"`
	ContactID  string                 `json:"contact_id"`
	BotID      string                 `json:"bot_id"`
	CampaignID string                 `json:"campaign_id"`
	Data       map[string]interface{} `json:"data"`
	Direction  int                    `json:"direction"`
	Status     int                    `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
}

type TelegramBotChat struct {
	Contact          *TelegramBotContact `json:"contact"`
	InboxLastMessage *TelegramBotMessage `json:"inbox_last_message"`
	InboxUnread      int                 `json:"inbox_unread"`
}

func (service *BotsTelegramService) GetBotChats(ctx context.Context, botID string) ([]*TelegramBotChat, error) {
	path := fmt.Sprintf("/telegram/chats?bot_id=%s", botID)

	var respData struct {
		Success bool               `json:"success"`
		Data    []*TelegramBotChat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsTelegramService) GetContactMessages(ctx context.Context, contactID string) ([]*TelegramBotMessage, error) {
	path := fmt.Sprintf("/telegram/chats/messages?contact_id=%s", contactID)

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*TelegramBotMessage `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type TelegramBotSendCampaignParams struct {
	Title    string                       `json:"title"`
	BotID    string                       `json:"bot_id"`
	SendAt   time.Time                    `json:"send_at"`
	Messages []TelegramBotCampaignMessage `json:"messages"`
}

type TelegramBotCampaignMessage struct {
	Type    string `json:"type"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

func (service *BotsTelegramService) SendCampaign(ctx context.Context, params TelegramBotSendCampaignParams) error {
	path := "/telegram/campaigns/send"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}
