package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type BotsWhatsAppService struct {
	client *Client
}

func newBotsWhatsAppService(cl *Client) *BotsWhatsAppService {
	return &BotsWhatsAppService{client: cl}
}

type WhatsAppAccount struct {
	Plan struct {
		Branding     bool      `json:"branding"`
		MaxBots      int       `json:"max_bots"`
		MaxContacts  int       `json:"max_contacts"`
		MaxMessages  int       `json:"max_messages"`
		MaxTags      int       `json:"max_tags"`
		MaxVariables int       `json:"max_variables"`
		Code         string    `json:"code"`
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

func (service *BotsWhatsAppService) GetAccount(ctx context.Context) (*WhatsAppAccount, error) {
	path := "/whatsapp/account"

	var respData struct {
		Success bool             `json:"success"`
		Data    *WhatsAppAccount `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type WhatsAppBot struct {
	ID          string `json:"id"`
	ChannelData struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	} `json:"channel_data"`
	Inbox struct {
		Total  int `json:"total"`
		Unread int `json:"unread"`
	} `json:"inbox"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (service *BotsWhatsAppService) GetBots(ctx context.Context) ([]*WhatsAppBot, error) {
	path := "/whatsapp/bots"

	var respData struct {
		Success bool           `json:"success"`
		Data    []*WhatsAppBot `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type WhatsAppBotContact struct {
	ID          string `json:"id"`
	BotID       string `json:"bot_id"`
	Status      int    `json:"status"`
	ChannelData struct {
		UserName     string `json:"username"`
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

type WhatsAppMessage struct {
	Type string `json:"type"`
	Text *struct {
		Body string `json:"body"`
	} `json:"text,omitempty"`
	Image *struct {
		Link    string `json:"link"`
		Caption string `json:"caption"`
	} `json:"image,omitempty"`
	Document *struct {
		Link    string `json:"link"`
		Caption string `json:"caption"`
	} `json:"document,omitempty"`
}

func (service *BotsWhatsAppService) CreateContact(ctx context.Context, botID, phone, name string) (*WhatsAppBotContact, error) {
	path := "/whatsapp/contacts"

	type bodyFormat struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
		BotID string `json:"bot_id"`
	}
	body := bodyFormat{
		Phone: phone,
		Name:  name,
		BotID: botID,
	}

	var respData struct {
		Success bool                `json:"success"`
		Data    *WhatsAppBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) GetContact(ctx context.Context, contactID string) (*WhatsAppBotContact, error) {
	path := fmt.Sprintf("/whatsapp/contacts/get?id=%s", contactID)

	var respData struct {
		Success bool                `json:"success"`
		Data    *WhatsAppBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) GetContactsByPhone(ctx context.Context, phone, botID string) ([]*WhatsAppBotContact, error) {
	path := fmt.Sprintf("/whatsapp/contacts/getByPhone?tag=%s&bot_id=%s", phone, botID)

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*WhatsAppBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) GetContactsByTag(ctx context.Context, tag, botID string) ([]*WhatsAppBotContact, error) {
	path := fmt.Sprintf("/whatsapp/contacts/getByTag?tag=%s&bot_id=%s", tag, botID)

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*WhatsAppBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) GetContactsByVariable(ctx context.Context, params BotContactsByVariableParams) ([]*WhatsAppBotContact, error) {
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
	path := "/whatsapp/contacts/getByVariable?" + urlParams.Encode()

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*WhatsAppBotContact `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) SendByContact(ctx context.Context, contactID string, message *WhatsAppMessage) error {
	path := "/whatsapp/contacts/send"

	type bodyFormat struct {
		ContactID string           `json:"contact_id"`
		Message   *WhatsAppMessage `json:"message"`
	}
	body := bodyFormat{
		ContactID: contactID,
		Message:   message,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendByPhone(ctx context.Context, botID, phone string, message *WhatsAppMessage) error {
	path := "/whatsapp/contacts/sendByPhone"

	type bodyFormat struct {
		BotID   string           `json:"bot_id"`
		Phone   string           `json:"phone"`
		Message *WhatsAppMessage `json:"message"`
	}
	body := bodyFormat{
		BotID:   botID,
		Phone:   phone,
		Message: message,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendTemplate(ctx context.Context, contactID, templateName, languageCode string) error {
	path := "/whatsapp/contacts/sendTemplate"

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
		Template  struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		} `json:"template"`
	}
	body := bodyFormat{
		ContactID: contactID,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		}{
			Name: templateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: languageCode,
			},
		},
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendTemplateWithVariables(ctx context.Context, contactID, templateName, languageCode string, variables []string) error {
	path := "/whatsapp/contacts/sendTemplate"

	type bodyComponentVariableFormat struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type bodyComponentFormat struct {
		Type       string                        `json:"type"`
		Parameters []bodyComponentVariableFormat `json:"parameters"`
	}

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
		Template  struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		} `json:"template"`
		Components []bodyComponentFormat `json:"components"`
	}
	bodyVariables := make([]bodyComponentVariableFormat, len(variables))
	for i, v := range variables {
		bodyVariables[i] = bodyComponentVariableFormat{
			Type: "text",
			Text: v,
		}
	}

	components := make([]bodyComponentFormat, 1)
	components[0] = bodyComponentFormat{
		Type:       "body",
		Parameters: bodyVariables,
	}

	body := bodyFormat{
		ContactID: contactID,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		}{
			Name: templateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: languageCode,
			},
		},
		Components: components,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendTemplateWithImage(ctx context.Context, contactID, templateName, languageCode, imageLink string) error {
	path := "/whatsapp/contacts/sendTemplate"

	type bodyComponentImageFormat struct {
		Type  string `json:"type"`
		Image struct {
			Link string `json:"link"`
		} `json:"image"`
	}

	type bodyComponentFormat struct {
		Type       string                     `json:"type"`
		Parameters []bodyComponentImageFormat `json:"parameters"`
	}

	type bodyFormat struct {
		ContactID string `json:"contact_id"`
		Template  struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		} `json:"template"`
		Components []bodyComponentFormat `json:"components"`
	}
	bodyImages := make([]bodyComponentImageFormat, 1)
	bodyImages[0] = bodyComponentImageFormat{
		Type: "image",
		Image: struct {
			Link string `json:"link"`
		}{
			Link: imageLink,
		},
	}

	components := make([]bodyComponentFormat, 1)
	components[0] = bodyComponentFormat{
		Type:       "header",
		Parameters: bodyImages,
	}

	body := bodyFormat{
		ContactID: contactID,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		}{
			Name: templateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: languageCode,
			},
		},
		Components: components,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendTemplateByPhone(ctx context.Context, botID, phone, templateName, languageCode string) error {
	path := "/whatsapp/contacts/sendTemplateByPhone"

	type bodyFormat struct {
		BotID    string `json:"bot_id"`
		Phone    string `json:"phone"`
		Template struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		} `json:"template"`
	}
	body := bodyFormat{
		BotID: botID,
		Phone: phone,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		}{
			Name: templateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: languageCode,
			},
		},
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendTemplateByPhoneWithVariables(ctx context.Context, botID, phone, templateName, languageCode string, variables []string) error {
	path := "/whatsapp/contacts/sendTemplateByPhone"

	type bodyComponentVariableFormat struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type bodyComponentFormat struct {
		Type       string                        `json:"type"`
		Parameters []bodyComponentVariableFormat `json:"parameters"`
	}

	type bodyFormat struct {
		BotID    string `json:"bot_id"`
		Phone    string `json:"phone"`
		Template struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		} `json:"template"`
		Components []bodyComponentFormat `json:"components"`
	}
	bodyVariables := make([]bodyComponentVariableFormat, len(variables))
	for i, v := range variables {
		bodyVariables[i] = bodyComponentVariableFormat{
			Type: "text",
			Text: v,
		}
	}

	components := make([]bodyComponentFormat, 1)
	components[0] = bodyComponentFormat{
		Type:       "body",
		Parameters: bodyVariables,
	}

	body := bodyFormat{
		BotID: botID,
		Phone: phone,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		}{
			Name: templateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: languageCode,
			},
		},
		Components: components,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SendTemplateByPhoneWithImage(ctx context.Context, botID, phone, templateName, languageCode, imageLink string) error {
	path := "/whatsapp/contacts/sendTemplateByPhone"

	type bodyComponentImageFormat struct {
		Type  string `json:"type"`
		Image struct {
			Link string `json:"link"`
		} `json:"image"`
	}

	type bodyComponentFormat struct {
		Type       string                     `json:"type"`
		Parameters []bodyComponentImageFormat `json:"parameters"`
	}

	type bodyFormat struct {
		BotID    string `json:"bot_id"`
		Phone    string `json:"phone"`
		Template struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		} `json:"template"`
		Components []bodyComponentFormat `json:"components"`
	}
	bodyImages := make([]bodyComponentImageFormat, 1)
	bodyImages[0] = bodyComponentImageFormat{
		Type: "image",
		Image: struct {
			Link string `json:"link"`
		}{
			Link: imageLink,
		},
	}

	components := make([]bodyComponentFormat, 1)
	components[0] = bodyComponentFormat{
		Type:       "header",
		Parameters: bodyImages,
	}

	body := bodyFormat{
		BotID: botID,
		Phone: phone,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
		}{
			Name: templateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: languageCode,
			},
		},
		Components: components,
	}

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

func (service *BotsWhatsAppService) SetVariableToContact(ctx context.Context, contactID string, variableID string, variableName string, variableValue interface{}) error {
	path := "/whatsapp/contacts/setVariable"

	type bodyFormat struct {
		ContactID     string      `json:"contact_id"`
		VariableID    string      `json:"variable_id,omitempty"`
		VariableName  string      `json:"variable_name,omitempty"`
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

func (service *BotsWhatsAppService) SetTagsToContact(ctx context.Context, contactID string, tags []string) error {
	path := "/whatsapp/contacts/setTag"

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

func (service *BotsWhatsAppService) DeleteTagFromContact(ctx context.Context, contactID string, tag string) error {
	path := "/whatsapp/contacts/deleteTag"

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

func (service *BotsWhatsAppService) DisableContact(ctx context.Context, contactID string) error {
	path := "/whatsapp/contacts/disable"

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

func (service *BotsWhatsAppService) EnableContact(ctx context.Context, contactID string) error {
	path := "/whatsapp/contacts/enable"

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

func (service *BotsWhatsAppService) DeleteContact(ctx context.Context, contactID string) error {
	path := "/whatsapp/contacts/delete"

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

func (service *BotsWhatsAppService) GetPauseAutomation(ctx context.Context, contactID string) (int, error) {
	path := fmt.Sprintf("/whatsapp/contacts/getPauseAutomation?contact_id=%s", contactID)

	var respData struct {
		Success bool `json:"success"`
		Data    struct {
			Minutes int `json:"minutes"`
		} `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data.Minutes, err
}

func (service *BotsWhatsAppService) SetPauseAutomation(ctx context.Context, contactID string, minutes int) error {
	path := "/whatsapp/contacts/setPauseAutomation"
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

func (service *BotsWhatsAppService) DeletePauseAutomation(ctx context.Context, contactID string) error {
	path := "/whatsapp/contacts/deletePauseAutomation"
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

func (service *BotsWhatsAppService) GetBotVariables(ctx context.Context, botID string) ([]*BotVariable, error) {
	path := fmt.Sprintf("/whatsapp/variables?bot_id=%s", botID)

	var respData struct {
		Success bool           `json:"success"`
		Data    []*BotVariable `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) GetFlows(ctx context.Context, botID string) ([]*BotFlow, error) {
	path := fmt.Sprintf("/whatsapp/flows?bot_id=%s", botID)

	var respData struct {
		Success bool       `json:"success"`
		Data    []*BotFlow `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) RunFlow(ctx context.Context, contactID, flowID string, externalData map[string]interface{}) error {
	path := "/whatsapp/flows/run"

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

func (service *BotsWhatsAppService) RunFlowByTrigger(ctx context.Context, contactID, triggerKeyword string, externalData map[string]interface{}) error {
	path := "/whatsapp/flows/runByTrigger"

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

func (service *BotsWhatsAppService) GetBotTriggers(ctx context.Context, botID string) ([]*BotTrigger, error) {
	path := fmt.Sprintf("/whatsapp/triggers?bot_id=%s", botID)

	var respData struct {
		Success bool          `json:"success"`
		Data    []*BotTrigger `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type WhatsAppBotMessage struct {
	ID         string                 `json:"id"`
	ContactID  string                 `json:"contact_id"`
	BotID      string                 `json:"bot_id"`
	CampaignID string                 `json:"campaign_id"`
	Data       map[string]interface{} `json:"data"`
	Direction  int                    `json:"direction"`
	Status     int                    `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
}

type WhatsAppBotChat struct {
	Contact          *WhatsAppBotContact `json:"contact"`
	InboxLastMessage *FbBotMessage       `json:"inbox_last_message"`
	InboxUnread      int                 `json:"inbox_unread"`
}

func (service *BotsWhatsAppService) GetBotChats(ctx context.Context, botID string) ([]*WhatsAppBotChat, error) {
	path := fmt.Sprintf("/whatsapp/chats?bot_id=%s", botID)

	var respData struct {
		Success bool               `json:"success"`
		Data    []*WhatsAppBotChat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *BotsWhatsAppService) GetContactMessages(ctx context.Context, contactID string) ([]*WhatsAppBotMessage, error) {
	path := fmt.Sprintf("/whatsapp/chats/messages?contact_id=%s", contactID)

	var respData struct {
		Success bool                  `json:"success"`
		Data    []*WhatsAppBotMessage `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type WhatsAppBotSendCampaignParams struct {
	Title    string            `json:"title"`
	BotID    string            `json:"bot_id"`
	SendAt   time.Time         `json:"send_at"`
	Messages []WhatsAppMessage `json:"messages"`
}

func (service *BotsWhatsAppService) SendCampaign(ctx context.Context, params WhatsAppBotSendCampaignParams) error {
	path := "/whatsapp/campaigns/send"

	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}

type WhatsAppBotSendCampaignByTemplateParams struct {
	Title        string
	BotID        string
	SendAt       time.Time
	Messages     []WhatsAppMessage
	TemplateName string
	LanguageCode string
}

func (service *BotsWhatsAppService) SendCampaignByTemplate(ctx context.Context, params WhatsAppBotSendCampaignByTemplateParams) error {
	path := "/whatsapp/campaigns/sendTemplate"
	type bodyFormat struct {
		Title    string       `json:"title"`
		BotID    string       `json:"bot_id"`
		SendAt   DateTimeType `json:"send_at"`
		Template struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
			Components struct {
				Type       string            `json:"type"`
				Parameters []WhatsAppMessage `json:"parameters"`
			} `json:"components"`
		}
	}
	body := bodyFormat{
		Title: params.Title,
		BotID: params.BotID,
		Template: struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
			Components struct {
				Type       string            `json:"type"`
				Parameters []WhatsAppMessage `json:"parameters"`
			} `json:"components"`
		}{
			Name: params.TemplateName,
			Language: struct {
				Code string `json:"code"`
			}{
				Code: params.LanguageCode,
			},
			Components: struct {
				Type       string            `json:"type"`
				Parameters []WhatsAppMessage `json:"parameters"`
			}{
				Type:       "header",
				Parameters: params.Messages,
			},
		},
	}
	var respData struct {
		Success bool `json:"success"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &respData, true)
	return err
}

type WhatsAppTemplate struct {
	ID             string            `json:"id"`
	BotID          string            `json:"bot_id"`
	Namespace      string            `json:"namespace"`
	Category       string            `json:"category"`
	Components     []WhatsAppMessage `json:"components"`
	Language       string            `json:"language"`
	Name           string            `json:"name"`
	RejectedReason string            `json:"rejected_reason"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
}

func (service *BotsWhatsAppService) GetTemplates(ctx context.Context) ([]*WhatsAppTemplate, error) {
	path := "/whatsapp/templates"

	var respData struct {
		Success bool                `json:"success"`
		Data    []*WhatsAppTemplate `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, nil, &respData, true)
	return respData.Data, err
}
