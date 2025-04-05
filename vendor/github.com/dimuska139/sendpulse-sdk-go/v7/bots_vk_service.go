package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BotsVkService struct {
	client *Client
}

func newBotsVkService(cl *Client) *BotsVkService {
	return &BotsVkService{client: cl}
}

type VkAccount struct {
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

func (service *BotsVkService) GetAccount(ctx context.Context) (*VkAccount, error) {
	path := "/vk/account"

	var respData struct {
		Success bool       `json:"success"`
		Data    *VkAccount `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type VkBot struct {
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

func (service *BotsVkService) GetBots(ctx context.Context) ([]*VkBot, error) {
	path := "/vk/bots"

	var respData struct {
		Success bool     `json:"success"`
		Data    []*VkBot `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type VkBotContact struct {
	ID          string `json:"id"`
	BotID       string `json:"bot_id"`
	Status      int    `json:"status"`
	ChannelData struct {
		GroupID  int         `json:"group_id"`
		IsMember bool        `json:"is_member"`
		Name     string      `json:"name"`
		Data     interface{} `json:"data"`
	} `json:"channel_data"`
	Tags                  []string               `json:"tags"`
	Variables             map[string]interface{} `json:"variables"`
	IsChatOpened          bool                   `json:"is_chat_opened"`
	LastActivityAt        time.Time              `json:"last_activity_at"`
	AutomationPausedUntil time.Time              `json:"automation_paused_until"`
	CreatedAt             time.Time              `json:"created_at"`
}

func (service *BotsVkService) GetContact(ctx context.Context, contactID string) (*VkBotContact, error) {
	path := fmt.Sprintf("/vk/contacts/get?id=%s", contactID)

	var respData struct {
		Success bool          `json:"success"`
		Data    *VkBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsVkService) GetContactsByTag(ctx context.Context, tag, botID string) ([]*VkBotContact, error) {
	path := fmt.Sprintf("/vk/contacts/getByTag?tag=%s&bot_id=%s", tag, botID)

	var respData struct {
		Success bool            `json:"success"`
		Data    []*VkBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsVkService) GetContactsByVariable(ctx context.Context, params BotContactsByVariableParams) ([]*VkBotContact, error) {
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
	path := "/vk/contacts/getByVariable?" + urlParams.Encode()

	var respData struct {
		Success bool            `json:"success"`
		Data    []*VkBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsVkService) SendTextByContact(ctx context.Context, contactID string, text string) error {
	path := "/vk/contacts/sendText"

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

func (service *BotsVkService) SetVariableToContact(ctx context.Context, contactID string, variableID string, variableName string, variableValue interface{}) error {
	path := "/vk/contacts/setVariable"

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

func (service *BotsVkService) SetTagsToContact(ctx context.Context, contactID string, tags []string) error {
	path := "/vk/contacts/setTag"

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

func (service *BotsVkService) DeleteTagFromContact(ctx context.Context, contactID string, tag string) error {
	path := "/vk/contacts/deleteTag"

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

func (service *BotsVkService) DisableContact(ctx context.Context, contactID string) error {
	path := "/vk/contacts/disable"

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

func (service *BotsVkService) EnableContact(ctx context.Context, contactID string) error {
	path := "/vk/contacts/enable"

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

func (service *BotsVkService) DeleteContact(ctx context.Context, contactID string) error {
	path := "/vk/contacts/delete"

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

func (service *BotsVkService) GetPauseAutomation(ctx context.Context, contactID string) (int, error) {
	path := fmt.Sprintf("/vk/contacts/getPauseAutomation?contact_id=%s", contactID)

	var respData struct {
		Success bool `json:"success"`
		Data    struct {
			Minutes int `json:"minutes"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data.Minutes, err
}

func (service *BotsVkService) SetPauseAutomation(ctx context.Context, contactID string, minutes int) error {
	path := "/vk/contacts/setPauseAutomation"
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

func (service *BotsVkService) DeletePauseAutomation(ctx context.Context, contactID string) error {
	path := "/vk/contacts/deletePauseAutomation"
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

func (service *BotsVkService) GetBotVariables(ctx context.Context, botID string) ([]*BotVariable, error) {
	path := fmt.Sprintf("/vk/variables?bot_id=%s", botID)

	var respData struct {
		Success bool           `json:"success"`
		Data    []*BotVariable `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsVkService) GetFlows(ctx context.Context, botID string) ([]*BotFlow, error) {
	path := fmt.Sprintf("/vk/flows?bot_id=%s", botID)

	var respData struct {
		Success bool       `json:"success"`
		Data    []*BotFlow `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsVkService) RunFlow(ctx context.Context, contactID, flowID string, externalData map[string]interface{}) error {
	path := "/vk/flows/run"

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

func (service *BotsVkService) RunFlowByTrigger(ctx context.Context, contactID, triggerKeyword string, externalData map[string]interface{}) error {
	path := "/vk/flows/runByTrigger"

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

func (service *BotsVkService) GetBotTriggers(ctx context.Context, botID string) ([]*BotTrigger, error) {
	path := fmt.Sprintf("/vk/triggers?bot_id=%s", botID)

	var respData struct {
		Success bool          `json:"success"`
		Data    []*BotTrigger `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type VkBotMessage struct {
	ID         string                 `json:"id"`
	ContactID  string                 `json:"contact_id"`
	BotID      string                 `json:"bot_id"`
	CampaignID string                 `json:"campaign_id"`
	Data       map[string]interface{} `json:"data"`
	Direction  int                    `json:"direction"`
	Status     int                    `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
}

type VkBotChat struct {
	Contact          *VkBotContact `json:"contact"`
	InboxLastMessage *VkBotMessage `json:"inbox_last_message"`
	InboxUnread      int           `json:"inbox_unread"`
}

func (service *BotsVkService) GetBotChats(ctx context.Context, botID string) ([]*VkBotChat, error) {
	path := fmt.Sprintf("/vk/chats?bot_id=%s", botID)

	var respData struct {
		Success bool         `json:"success"`
		Data    []*VkBotChat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsVkService) GetContactMessages(ctx context.Context, contactID string) ([]*VkBotMessage, error) {
	path := fmt.Sprintf("/vk/chats/messages?contact_id=%s", contactID)

	var respData struct {
		Success bool            `json:"success"`
		Data    []*VkBotMessage `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type VkBotSendCampaignParams struct {
	Title    string                 `json:"title"`
	BotID    string                 `json:"bot_id"`
	SendAt   time.Time              `json:"send_at"`
	Messages []VkBotCampaignMessage `json:"messages"`
}

type VkBotCampaignMessage struct {
	Type    string `json:"type"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

func (service *BotsVkService) SendCampaign(ctx context.Context, params VkBotSendCampaignParams) error {
	path := "/vk/campaigns/send"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}
