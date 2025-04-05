package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
)

// MailingListsService is a service to interact with mailing lists
type MailingListsService struct {
	client *Client
}

// newMailingListsService creates MailingListsService
func newMailingListsService(cl *Client) *MailingListsService {
	return &MailingListsService{client: cl}
}

// CreateMailingList creates new mailing list
func (service *MailingListsService) CreateMailingList(ctx context.Context, name string) (int, error) {
	path := "/addressbooks"

	type data struct {
		Name string `json:"bookName"`
	}

	var response struct {
		ID int `json:"id"`
	}
	params := data{Name: name}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &response, true)
	return response.ID, err
}

// ChangeName changes a name of specific mailing list
func (service *MailingListsService) ChangeName(ctx context.Context, id int, name string) error {
	path := fmt.Sprintf("/addressbooks/%d", id)

	type data struct {
		Name string `json:"name"`
	}

	var response struct {
		Result bool
	}
	params := data{Name: name}
	_, err := service.client.newRequest(ctx, http.MethodPut, path, params, &response, true)
	return err
}

// MailingList represents detailed information of specific mailing list
type MailingList struct {
	ID               int          `json:"id"`
	Name             string       `json:"name"`
	AllEmailQty      int          `json:"all_email_qty"`
	ActiveEmailQty   int          `json:"active_email_qty"`
	InactiveEmailQty int          `json:"inactive_email_qty"`
	CreationDate     DateTimeType `json:"creationdate"`
	Status           int          `json:"status"`
	StatusExplain    string       `json:"status_explain"`
}

// GetMailingLists returns a list of mailing lists
func (service *MailingListsService) GetMailingLists(ctx context.Context, limit int, offset int) ([]*MailingList, error) {
	path := fmt.Sprintf("/addressbooks?limit=%d&offset=%d", limit, offset)
	var books []*MailingList
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &books, true)
	return books, err
}

// GetMailingList returns detailed information regarding a specific mailing list
func (service *MailingListsService) GetMailingList(ctx context.Context, mailingListID int) (*MailingList, error) {
	path := fmt.Sprintf("/addressbooks/%d", mailingListID)
	var books []*MailingList
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &books, true)
	var book *MailingList
	if len(books) != 0 {
		book = books[0]
	}
	return book, err
}

// VariableMeta method represents a variable of mailing list
type VariableMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetMailingListVariables method returns variables of specific mailing list
func (service *MailingListsService) GetMailingListVariables(ctx context.Context, mailingListID int) ([]*VariableMeta, error) {
	path := fmt.Sprintf("/addressbooks/%d/variables", mailingListID)
	var variables []*VariableMeta
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &variables, true)
	return variables, err
}

// Email describes email address
type Email struct {
	Email         string                 `json:"email"`
	Phone         int                    `json:"phone"`
	Status        int                    `json:"status"`
	StatusExplain string                 `json:"status_explain"`
	Variables     map[string]interface{} `json:"variables"`
}

// GetMailingListEmails returns a list of emails from a mailing list
func (service *MailingListsService) GetMailingListEmails(ctx context.Context, id, limit, offset int) ([]*Email, error) {
	path := fmt.Sprintf("/addressbooks/%d/emails?limit=%d&offset=%d", id, limit, offset)
	var emails []*Email
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &emails, true)
	return emails, err
}

// CountMailingListEmails returns a the total number of contacts in a mailing list
func (service *MailingListsService) CountMailingListEmails(ctx context.Context, mailingListID int) (int, error) {
	path := fmt.Sprintf("/addressbooks/%d/emails/total", mailingListID)
	var response struct {
		Total int
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response.Total, err
}

// GetMailingListEmailsByVariable returns all contacts in mailing list by value of variable
func (service *MailingListsService) GetMailingListEmailsByVariable(ctx context.Context, mailingListID int, variable string, value interface{}) ([]*Email, error) {
	path := fmt.Sprintf("/addressbooks/%d/variables/%s/%v", mailingListID, variable, value)
	var emails []*Email
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &emails, true)
	return emails, err
}

// EmailToAdd represents structure for add email to mailing list
type EmailToAdd struct {
	Email     string                 `json:"email"`
	Variables map[string]interface{} `json:"variables"`
}

// SingleOptIn adds emails to mailing list using single-opt-in method
func (service *MailingListsService) SingleOptIn(ctx context.Context, mailingListID int, emails []*EmailToAdd) error {
	path := fmt.Sprintf("/addressbooks/%d/emails", mailingListID)
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Emails []*EmailToAdd `json:"emails"`
	}

	body := bodyFormat{Emails: emails}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &response, true)
	return err
}

// DoubleOptIn adds emails to mailing list using double-opt-in method
func (service *MailingListsService) DoubleOptIn(ctx context.Context, mailingListID int, emails []*EmailToAdd, senderEmail string, messageLang string, templateID string) error {
	path := fmt.Sprintf("/addressbooks/%d/emails", mailingListID)
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Emails       []*EmailToAdd `json:"emails"`
		Confirmation string        `json:"confirmation"`
		SenderEmail  string        `json:"sender_email"`
		MessageLang  string        `json:"message_lang"`
		TemplateID   string        `json:"template_id,omitempty"`
	}

	body := bodyFormat{
		Emails:       emails,
		Confirmation: "force",
		SenderEmail:  senderEmail,
		MessageLang:  messageLang,
	}

	if templateID != "" {
		body.TemplateID = templateID
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &response, true)
	return err
}

// DeleteMailingListEmails removes emails from specific mailing list
func (service *MailingListsService) DeleteMailingListEmails(ctx context.Context, mailingListID int, emails []string) error {
	path := fmt.Sprintf("/addressbooks/%d/emails", mailingListID)
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Emails []string `json:"emails"`
	}
	body := bodyFormat{Emails: emails}

	_, err := service.client.newRequest(ctx, http.MethodDelete, path, body, &response, true)
	return err
}

// DeleteMailingList removes specific mailing list
func (service *MailingListsService) DeleteMailingList(ctx context.Context, mailingListID int) error {
	path := fmt.Sprintf("/addressbooks/%d", mailingListID)
	var response struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, nil, &response, true)
	return err
}

// CampaignCost represents the cost of a campaign sent to a mailing list
type CampaignCost struct {
	Cur                       string `json:"email"`
	SentEmailsQty             int    `json:"sent_emails_qty"`
	OverdraftAllEmailsPrice   int    `json:"overdraft_all_emails_price"`
	AddressesDeltaFromBalance int    `json:"address_delta_from_balance"`
	AddressesDeltaFromTariff  int    `json:"address_delta_from_tariff"`
	MaxEmailsPerTask          int    `json:"max_emails_per_task"`
	Result                    bool   `json:"result"`
}

// CountCampaignCost calculates the cost of a campaign sent to a mailing list
func (service *MailingListsService) CountCampaignCost(ctx context.Context, mailingListID int) (*CampaignCost, error) {
	path := fmt.Sprintf("/addressbooks/%d/cost", mailingListID)
	var cost CampaignCost

	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &cost, true)
	return &cost, err
}

// UnsubscribeEmails unsubscribes emails from a specific mailing list
func (service *MailingListsService) UnsubscribeEmails(ctx context.Context, mailingListID int, emails []string) error {
	path := fmt.Sprintf("/addressbooks/%d/emails/unsubscribe", mailingListID)
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Emails []string `json:"emails"`
	}
	body := bodyFormat{Emails: emails}

	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &response, true)
	return err
}

// UpdateEmailVariables changes a variables for an email contact
func (service *MailingListsService) UpdateEmailVariables(ctx context.Context, mailingListID int, email string, variables []*Variable) error {
	path := fmt.Sprintf("/addressbooks/%d/emails/variable", mailingListID)
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Email     string      `json:"email"`
		Variables []*Variable `json:"variables"`
	}

	body := bodyFormat{Email: email, Variables: variables}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &response, true)
	return err
}
