package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BotsIgService struct {
	client *Client
}

func newBotsIgService(cl *Client) *BotsIgService {
	return &BotsIgService{client: cl}
}

type IgAccount struct {
	Tariff struct {
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
	} `json:"tariff"`
	Statistics struct {
		Messages  int `json:"messages"`
		Bots      int `json:"bots"`
		Contacts  int `json:"contacts"`
		Variables int `json:"variables"`
	} `json:"statistics"`
}

func (service *BotsIgService) GetAccount(ctx context.Context) (*IgAccount, error) {
	path := "/instagram/account"

	var respData struct {
		Success bool       `json:"success"`
		Data    *IgAccount `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type IgBot struct {
	ID          string `json:"id"`
	ChannelData struct {
		ID         string `json:"id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Name       string `json:"name"`
		NameFormat string `json:"name_format"`
		ShortName  string `json:"short_name"`
		Picture    struct {
			Data struct {
				Height       int    `json:"height"`
				IsSilhouette bool   `json:"is_silhouette"`
				Url          string `json:"url"`
				Width        int    `json:"width"`
			} `json:"data"`
		} `json:"picture"`
	} `json:"channel_data"`
	IgUser struct {
		ID                int    `json:"id"`
		IgID              int    `json:"ig_id"`
		FollowersCount    int    `json:"followers_count"`
		FollowsCount      int    `json:"follows_count"`
		MediaCount        int    `json:"media_count"`
		ProfilePictureUrl string `json:"profile_picture_url"`
		Username          string `json:"username"`
		Website           string `json:"website"`
	} `json:"ig_user"`
	IgPage struct {
		InstagramBusinessAccount struct {
			ID                int    `json:"id"`
			IgID              int    `json:"ig_id"`
			Name              string `json:"name"`
			Biography         string `json:"biography"`
			FollowersCount    int    `json:"followers_count"`
			FollowsCount      int    `json:"follows_count"`
			MediaCount        int    `json:"media_count"`
			ProfilePictureUrl string `json:"profile_picture_url"`
			Username          string `json:"username"`
			Website           string `json:"website"`
		} `json:"instagram_business_account"`
		ID           int    `json:"id"`
		Category     string `json:"category"`
		CategoryList []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"category_list"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				Height       int    `json:"height"`
				IsSilhouette bool   `json:"is_silhouette"`
				Url          string `json:"url"`
				Width        int    `json:"width"`
			} `json:"data"`
		} `json:"picture"`
	} `json:"ig_page"`
	Inbox struct {
		Total  int `json:"total"`
		Unread int `json:"unread"`
	} `json:"inbox"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (service *BotsIgService) GetBots(ctx context.Context) ([]*IgBot, error) {
	path := "/instagram/bots"

	var respData struct {
		Success bool     `json:"success"`
		Data    []*IgBot `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type IgBotContact struct {
	ID          string `json:"id"`
	BotID       string `json:"bot_id"`
	Status      int    `json:"status"`
	ChannelData struct {
		ID         string `json:"id"`
		UserName   string `json:"user_name"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Name       string `json:"name"`
		ProfilePic string `json:"profile_pic"`
	} `json:"channel_data"`
	Tags                  []string               `json:"tags"`
	Variables             map[string]interface{} `json:"variables"`
	IsChatOpened          bool                   `json:"is_chat_opened"`
	LastActivityAt        time.Time              `json:"last_activity_at"`
	AutomationPausedUntil time.Time              `json:"automation_paused_until"`
	CreatedAt             time.Time              `json:"created_at"`
}

func (service *BotsIgService) GetContact(ctx context.Context, contactID string) (*IgBotContact, error) {
	path := fmt.Sprintf("/instagram/contacts/get?id=%s", contactID)

	var respData struct {
		Success bool          `json:"success"`
		Data    *IgBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsIgService) GetContactsByTag(ctx context.Context, tag, botID string) ([]*IgBotContact, error) {
	path := fmt.Sprintf("/instagram/contacts/getByTag?tag=%s&bot_id=%s", tag, botID)

	var respData struct {
		Success bool            `json:"success"`
		Data    []*IgBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsIgService) GetContactsByVariable(ctx context.Context, params BotContactsByVariableParams) ([]*IgBotContact, error) {
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
	path := "/instagram/contacts/getByVariable?" + urlParams.Encode()

	var respData struct {
		Success bool            `json:"success"`
		Data    []*IgBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type IgBotSendMessagesParams struct {
	ContactID string `json:"contact_id"`
	Messages  []struct {
		Type    string `json:"type"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	} `json:"messages"`
}

func (service *BotsIgService) SendTextByContact(ctx context.Context, params IgBotSendMessagesParams) error {
	path := "/instagram/contacts/sendText"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}

func (service *BotsIgService) SetVariableToContact(ctx context.Context, contactID string, variableID string, variableName string, variableValue interface{}) error {
	path := "/instagram/contacts/setVariable"

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

func (service *BotsIgService) SetTagsToContact(ctx context.Context, contactID string, tags []string) error {
	path := "/instagram/contacts/setTag"

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

func (service *BotsIgService) DeleteTagFromContact(ctx context.Context, contactID string, tag string) error {
	path := "/instagram/contacts/deleteTag"

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

func (service *BotsIgService) DisableContact(ctx context.Context, contactID string) error {
	path := "/instagram/contacts/disable"

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

func (service *BotsIgService) EnableContact(ctx context.Context, contactID string) error {
	path := "/instagram/contacts/enable"

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

func (service *BotsIgService) DeleteContact(ctx context.Context, contactID string) error {
	path := "/instagram/contacts/delete"

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

func (service *BotsIgService) GetPauseAutomation(ctx context.Context, contactID string) (int, error) {
	path := fmt.Sprintf("/instagram/contacts/getPauseAutomation?contact_id=%s", contactID)

	var respData struct {
		Success bool `json:"success"`
		Data    struct {
			Minutes int `json:"minutes"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data.Minutes, err
}

func (service *BotsIgService) SetPauseAutomation(ctx context.Context, contactID string, minutes int) error {
	path := "/instagram/contacts/setPauseAutomation"
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

func (service *BotsIgService) DeletePauseAutomation(ctx context.Context, contactID string) error {
	path := "/instagram/contacts/deletePauseAutomation"
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

func (service *BotsIgService) GetBotVariables(ctx context.Context, botID string) ([]*BotVariable, error) {
	path := fmt.Sprintf("/instagram/variables?bot_id=%s", botID)

	var respData struct {
		Success bool           `json:"success"`
		Data    []*BotVariable `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type BotIgFlow struct {
	ID     string `json:"id"`
	BotID  string `json:"bot_id"`
	Name   string `json:"name"`
	Status struct {
		Active   int `json:"ACTIVE"`
		Inactive int `json:"INACTIVE"`
		Draft    int `json:"DRAFT"`
	} `json:"status"`
	Triggers []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type int    `json:"type"`
	} `json:"triggers"`
	CreatedAt time.Time `json:"created_at"`
}

func (service *BotsIgService) GetFlows(ctx context.Context, botID string) ([]*BotIgFlow, error) {
	path := fmt.Sprintf("/instagram/flows?bot_id=%s", botID)

	var respData struct {
		Success bool         `json:"success"`
		Data    []*BotIgFlow `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsIgService) RunFlow(ctx context.Context, contactID, flowID string, externalData map[string]interface{}) error {
	path := "/instagram/flows/run"

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

func (service *BotsIgService) RunFlowByTrigger(ctx context.Context, contactID, triggerKeyword string, externalData map[string]interface{}) error {
	path := "/instagram/flows/runByTrigger"

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

func (service *BotsIgService) GetBotTriggers(ctx context.Context, botID string) ([]*BotTrigger, error) {
	path := fmt.Sprintf("/instagram/triggers?bot_id=%s", botID)

	var respData struct {
		Success bool          `json:"success"`
		Data    []*BotTrigger `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type IgBotMessage struct {
	ID         string                 `json:"id"`
	ContactID  string                 `json:"contact_id"`
	BotID      string                 `json:"bot_id"`
	CampaignID string                 `json:"campaign_id"`
	Data       map[string]interface{} `json:"data"`
	Direction  int                    `json:"direction"`
	Status     int                    `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
}

type IgBotChat struct {
	Contact          *IgBotContact `json:"contact"`
	InboxLastMessage *IgBotMessage `json:"inbox_last_message"`
	InboxUnread      int           `json:"inbox_unread"`
}

func (service *BotsIgService) GetBotChats(ctx context.Context, botID string) ([]*IgBotChat, error) {
	path := fmt.Sprintf("/instagram/chats?bot_id=%s", botID)

	var respData struct {
		Success bool         `json:"success"`
		Data    []*IgBotChat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsIgService) GetContactMessages(ctx context.Context, contactID string) ([]*IgBotMessage, error) {
	path := fmt.Sprintf("/instagram/chats/messages?contact_id=%s", contactID)

	var respData struct {
		Success bool            `json:"success"`
		Data    []*IgBotMessage `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type IgBotSendCampaignParams struct {
	Title    string                 `json:"title"`
	BotID    string                 `json:"bot_id"`
	SendAt   time.Time              `json:"send_at"`
	Messages []IgBotCampaignMessage `json:"messages"`
}

type IgBotCampaignMessage struct {
	Type    string `json:"type"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

func (service *BotsIgService) SendCampaign(ctx context.Context, params IgBotSendCampaignParams) error {
	path := "/instagram/campaigns/send"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}
