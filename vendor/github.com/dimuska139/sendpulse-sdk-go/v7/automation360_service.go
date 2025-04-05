package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
)

// Automation360Service is a service to interact with conversions
type Automation360Service struct {
	client *Client
}

// newAutomation360Service creates Automation360Service
func newAutomation360Service(cl *Client) *Automation360Service {
	return &Automation360Service{client: cl}
}

// Autoresponder represents statistics about an automation flow
type Autoresponder struct {
	Autoresponder struct {
		ID      int          `json:"id"`
		Name    string       `json:"name"`
		Status  int          `json:"status"`
		Created DateTimeType `json:"created"`
		Changed DateTimeType `json:"changed"`
	} `json:"autoresponder"`
	Flows []*struct {
		ID       int                    `json:"id"`
		MainID   int                    `json:"main_id"`
		AfType   string                 `json:"af_type"`
		Created  DateTimeType           `json:"created"`
		LastSend DateTimeType           `json:"last_send"`
		Task     map[string]interface{} `json:"task"`
	} `json:"flows"`
	Starts       int `json:"starts"`
	InQueue      int `json:"in_queue"`
	EndCount     int `json:"end_count"`
	SendMessages int `json:"send_messages"`
	Conversions  int `json:"conversions"`
}

// GetAutoresponderStatistics returns statistics about an automation flow
func (service *Automation360Service) GetAutoresponderStatistics(ctx context.Context, id int) (*Autoresponder, error) {
	path := fmt.Sprintf("/a360/autoresponders/%d", id)

	var respData *Autoresponder
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// StartEvent sends event to SendPulse
func (service *Automation360Service) StartEvent(ctx context.Context, eventName string, variables map[string]interface{}) error {
	path := fmt.Sprintf("/events/name/%s", eventName)

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, variables, &respData, true)
	return err
}

// MainTriggerBlockStat represents statistics about the "Start" element
type MainTriggerBlockStat struct {
	FlowID   int `json:"flow_id"`
	Executed int `json:"executed"`
	Deleted  int `json:"deleted"`
}

// GetStartBlockStatistics returns statistics about the "Start" element
func (service *Automation360Service) GetStartBlockStatistics(ctx context.Context, id int) (*MainTriggerBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/main-trigger/%d/group-stat", id)

	var respData struct {
		Data *MainTriggerBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// EmailBlockStat represents statistics about the "Email" element
type EmailBlockStat struct {
	FlowID int `json:"flow_id"`
	Task   struct {
		ID                int          `json:"id"`
		MailingListID     int          `json:"address_book_id"`
		MessageTitle      string       `json:"message_title"`
		SenderMailAddress string       `json:"sender_mail_address"`
		SenderMailName    string       `json:"sender_mail_name"`
		Created           DateTimeType `json:"created"`
	} `json:"task"`
	Sent         int          `json:"sent"`
	Delivered    int          `json:"delivered"`
	Opened       int          `json:"opened"`
	Clicked      int          `json:"clicked"`
	Errors       int          `json:"errors"`
	Unsubscribed int          `json:"unsubscribed"`
	MarkedAsSpam int          `json:"marked_as_spam"`
	LastSend     DateTimeType `json:"last_send"`
}

// GetEmailBlockStatistics returns statistics about the "Email" element
func (service *Automation360Service) GetEmailBlockStatistics(ctx context.Context, id int) (*EmailBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/email/%d/group-stat", id)

	var respData struct {
		Data *EmailBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// PushBlockStat represents statistics about the "Push" element
type PushBlockStat struct {
	FlowID    int          `json:"flow_id"`
	Sent      int          `json:"sent"`
	Delivered int          `json:"delivered"`
	Clicked   int          `json:"clicked"`
	Errors    int          `json:"errors"`
	LastSend  DateTimeType `json:"last_send"`
}

// GetPushBlockStatistics returns statistics about the "Push" element
func (service *Automation360Service) GetPushBlockStatistics(ctx context.Context, id int) (*PushBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/push/%d/group-stat", id)

	var respData struct {
		Data *PushBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// SmsBlockStat represents statistics about the "SMS" element
type SmsBlockStat struct {
	FlowID    int          `json:"flow_id"`
	Executed  int          `json:"executed"`
	Sent      int          `json:"sent"`
	Delivered int          `json:"delivered"`
	Opened    int          `json:"opened"`
	Clicked   int          `json:"clicked"`
	Errors    int          `json:"errors"`
	LastSend  DateTimeType `json:"last_send"`
}

// GetSmsBlockStatistics returns statistics about the "SMS" element
func (service *Automation360Service) GetSmsBlockStatistics(ctx context.Context, id int) (*SmsBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/sms/%d/group-stat", id)

	var respData struct {
		Data *SmsBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// MessengerBlockStat represents statistics about the "Messenger" element
type MessengerBlockStat struct {
	FlowID   int          `json:"flow_id"`
	Executed int          `json:"executed"`
	Sent     int          `json:"sent"`
	LastSend DateTimeType `json:"last_send"`
}

// GetMessengerBlockStatistics returns statistics about the "Messenger" element
func (service *Automation360Service) GetMessengerBlockStatistics(ctx context.Context, id int) (*MessengerBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/messenger/%d/group-stat", id)

	var respData struct {
		Data *MessengerBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// FilterBlockStat represents statistics about the "Filter" element
type FilterBlockStat struct {
	FlowID   int          `json:"flow_id"`
	Executed int          `json:"executed"`
	LastSend DateTimeType `json:"last_send"`
}

// GetFilterBlockStatistics returns statistics about the "Filter" element
func (service *Automation360Service) GetFilterBlockStatistics(ctx context.Context, id int) (*FilterBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/filter/%d/group-stat", id)

	var respData struct {
		Data *FilterBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// TriggerBlockStat represents statistics about the "Condition" element
type TriggerBlockStat struct {
	FlowID   int          `json:"flow_id"`
	Executed int          `json:"executed"`
	LastSend DateTimeType `json:"last_send"`
}

// GetTriggerBlockStatistics returns statistics about the "Condition" element
func (service *Automation360Service) GetTriggerBlockStatistics(ctx context.Context, id int) (*TriggerBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/trigger/%d/group-stat", id)

	var respData struct {
		Data *TriggerBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// GoalBlockStat represents statistics about the "Goal" element
type GoalBlockStat struct {
	FlowID    int          `json:"flow_id"`
	Executed  int          `json:"executed"`
	Sent      int          `json:"sent"`
	Delivered int          `json:"delivered"`
	Opened    int          `json:"opened"`
	Clicked   int          `json:"clicked"`
	Errors    int          `json:"errors"`
	LastSend  DateTimeType `json:"last_send"`
}

// GetGoalBlockStatistics returns statistics about the "Goal" element
func (service *Automation360Service) GetGoalBlockStatistics(ctx context.Context, id int) (*GoalBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/goal/%d/group-stat", id)

	var respData struct {
		Data *GoalBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// ActionBlockStat represents statistics about the "Action" element
type ActionBlockStat struct {
	FlowID   int          `json:"flow_id"`
	Executed int          `json:"executed"`
	LastSend DateTimeType `json:"last_send"`
}

// GetActionBlockStatistics returns statistics about the "Action" element
func (service *Automation360Service) GetActionBlockStatistics(ctx context.Context, id int) (*ActionBlockStat, error) {
	path := fmt.Sprintf("/a360/stats/action/%d/group-stat", id)

	var respData struct {
		Data *ActionBlockStat `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// AutoresponderConversion represents the flow conversion
type AutoresponderConversion struct {
	TotalConversions       int `json:"total_conversions"`
	MaintriggerConversions int `json:"maintrigger_conversions"`
	GoalConversions        int `json:"goal_conversions"`
	Maintrigger            struct {
		ID          int          `json:"id"`
		MainID      int          `json:"main_id"`
		AfType      string       `json:"af_type"`
		Created     DateTimeType `json:"created"`
		LastSend    DateTimeType `json:"last_send"`
		Conversions int          `json:"conversions"`
	} `json:"maintrigger"`
	Goals []struct {
		ID          int          `json:"id"`
		Name        string       `json:"name"`
		MainID      int          `json:"main_id"`
		AfType      string       `json:"af_type"`
		Created     DateTimeType `json:"created"`
		Conversions int          `json:"conversions"`
	} `json:"goals"`
}

// GetAutoresponderConversions returns the flow conversions list
func (service *Automation360Service) GetAutoresponderConversions(ctx context.Context, id int) (*AutoresponderConversion, error) {
	path := fmt.Sprintf("/a360/autoresponders/%d/conversions", id)

	var respData struct {
		Data *AutoresponderConversion `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

// AutoresponderContact represents the contact that converted
type AutoresponderContact struct {
	ID             int          `json:"id"`
	ConversionType string       `json:"conversion_type"`
	FlowID         int          `json:"flow_id"`
	Email          string       `json:"email"`
	Phone          string       `json:"phone"`
	ConversionDate DateTimeType `json:"conversion_date"`
	StartDate      DateTimeType `json:"start_date"`
}

// GetAutoresponderContacts returns a list of the contacts that converted
func (service *Automation360Service) GetAutoresponderContacts(ctx context.Context, id int) ([]*AutoresponderContact, error) {
	path := fmt.Sprintf("/a360/autoresponders/%d/conversions/list/all", id)

	var respData struct {
		Items []*AutoresponderContact `json:"items"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Items, err
}
