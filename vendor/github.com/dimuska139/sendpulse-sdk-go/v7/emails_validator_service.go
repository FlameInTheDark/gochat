package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ValidatorService is a service to validate email addresses
type ValidatorService struct {
	client *Client
}

// newValidatorService creates ValidatorService
func newValidatorService(cl *Client) *ValidatorService {
	return &ValidatorService{client: cl}
}

// ValidateMailingList sends a mailing list for review
func (service *ValidatorService) ValidateMailingList(ctx context.Context, mailingListID int) error {
	path := "/verifier-service/send-list-to-verify/"
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		ID int `json:"id"`
	}
	body := bodyFormat{ID: mailingListID}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &response, true)
	return err
}

// ValidationProgress is a results of mailing list validation progress
type ValidationProgress struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
}

// GetMailingListValidationProgress returns a progress of mailing list validation
func (service *ValidatorService) GetMailingListValidationProgress(ctx context.Context, mailingListID int) (*ValidationProgress, error) {
	path := fmt.Sprintf("/verifier-service/get-progress/?id=%d", mailingListID)
	var response struct {
		Result bool                `json:"result"`
		Data   *ValidationProgress `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response.Data, err
}

// MailingListValidationResult is a results of email validation progress
type MailingListValidationResult struct {
	ID                int          `json:"id"`
	Name              string       `json:"address_book_name"`
	AllEmailsQuantity int          `json:"all_emails_quantity"`
	Status            int          `json:"status"`
	CheckDate         DateTimeType `json:"check_date"`
	Data              struct {
		Unverified  int `json:"0"`
		Valid       int `json:"1"`
		Unconfirmed int `json:"2"`
		Invalid     int `json:"3"`
	} `json:"data"`
	IsUpdated       int    `json:"is_updated"`
	StatusText      string `json:"status_text"`
	IsGarbageInBook bool   `json:"is_garbage_in_book"`
}

// MailingListValidationResultDetailed is a detailed result of email validation progress
type MailingListValidationResultDetailed struct {
	MailingListValidationResult
	EmailAddresses []struct {
		ID           int          `json:"id"`
		EmailAddress string       `json:"email_address"`
		CheckDate    DateTimeType `json:"check_date"`
		Status       int          `json:"status"`
		StatusText   string       `json:"status_text"`
	} `json:"email_addresses"`
	EmailAddressesTotal int `json:"email_addresses_total"`
}

// GetMailingListValidationResult returns a list of email addresses from a mailing list with their verification results
func (service *ValidatorService) GetMailingListValidationResult(ctx context.Context, mailingListID int) (*MailingListValidationResultDetailed, error) {
	path := fmt.Sprintf("/verifier-service/check/?id=%d", mailingListID)
	var response *MailingListValidationResultDetailed
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response, err
}

// GetValidatedMailingLists returns a list of verified mailing lists
func (service *ValidatorService) GetValidatedMailingLists(ctx context.Context, limit, offset int) ([]*MailingListValidationResult, error) {
	path := fmt.Sprintf("/verifier-service/check-list?start=%d&count=%d", offset, limit)
	var response struct {
		Total int                            `json:"total"`
		List  []*MailingListValidationResult `json:"list"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response.List, err
}

// ValidateEmail verifies one email address
func (service *ValidatorService) ValidateEmail(ctx context.Context, email string) error {
	path := "/verifier-service/send-single-to-verify/"
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Email string `json:"email"`
	}
	body := bodyFormat{Email: email}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, body, &response, true)
	return err
}

// EmailValidationResult describes a result of a verification of specific email
type EmailValidationResult struct {
	Email  string `json:"email"`
	Checks struct {
		Status      int    `json:"status"`
		ValidFormat int    `json:"valid_format"`
		Disposable  int    `json:"disposable"`
		Webmail     int    `json:"webmail"`
		Gibberish   int    `json:"gibberish"`
		StatusText  string `json:"status_text"`
	} `json:"checks"`
}

// GetEmailValidationResult returns the results of a verification of specific email
func (service *ValidatorService) GetEmailValidationResult(ctx context.Context, email string) (*EmailValidationResult, error) {
	path := fmt.Sprintf("/verifier-service/get-single-result/?email=%s", email)
	var response struct {
		Result bool                   `json:"result"`
		Data   *EmailValidationResult `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response.Data, err
}

// DeleteEmailValidationResult removes the result of checking one address
func (service *ValidatorService) DeleteEmailValidationResult(ctx context.Context, email string) error {
	path := "/verifier-service/delete-single-result"
	var response struct {
		Result bool `json:"result"`
	}
	type bodyFormat struct {
		Email string `json:"email"`
	}
	body := bodyFormat{Email: email}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, body, &response, true)
	return err
}

// MailingListReportParams describes parameters to CreateMailingListValidationReport
type MailingListReportParams struct {
	ID       int    `json:"id"`
	Format   int    `json:"format,omitempty"`
	Statuses []int  `json:"status,omitempty"`
	Lang     string `json:"lang,omitempty"`
}

// CreateMailingListValidationReport creates a report with the verification results for a given mailing list
func (service *ValidatorService) CreateMailingListValidationReport(ctx context.Context, params MailingListReportParams) error {
	path := "/verifier-service/make-report"
	var response struct {
		Result bool `json:"result"`
	}

	type bodyFormat struct {
		ID       int    `json:"id"`
		Format   int    `json:"format,omitempty"`
		Statuses string `json:"status"`
		Lang     string `json:"lang,omitempty"`
	}

	body := bodyFormat{
		ID:     params.ID,
		Format: params.Format,
		Lang:   params.Lang,
	}

	if len(params.Statuses) != 0 {
		strStatuses := make([]string, 0)
		for _, status := range params.Statuses {
			strStatuses = append(strStatuses, strconv.Itoa(status))
		}
		body.Statuses = "[" + strings.Join(strStatuses, ",") + "]"
	}

	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &response, true)
	return err
}

// GetMailingListValidationReport returns a report with the results of a mailing list verification
func (service *ValidatorService) GetMailingListValidationReport(ctx context.Context, mailingListID int) (*MailingListValidationResultDetailed, error) {
	path := fmt.Sprintf("/verifier-service/check-report?id=%d", mailingListID)
	var response *MailingListValidationResultDetailed
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response, err
}
