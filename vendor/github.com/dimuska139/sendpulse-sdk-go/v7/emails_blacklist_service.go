package sendpulse_sdk_go

import (
	"context"
	b64 "encoding/base64"
	"net/http"
	"strings"
)

// BlacklistService is a service to interact with blacklist
type BlacklistService struct {
	client *Client
}

// newBlacklistService creates BlacklistService
func newBlacklistService(cl *Client) *BlacklistService {
	return &BlacklistService{client: cl}
}

// AddToBlacklist appends an email addresses to a blacklist
func (service *BlacklistService) AddToBlacklist(ctx context.Context, emails []string, comment string) error {
	path := "/blacklist"

	type paramsFormat struct {
		Emails  string `json:"emails"`
		Comment string `json:"comment,omitempty"`
	}

	params := paramsFormat{
		Emails: b64.StdEncoding.EncodeToString([]byte(strings.Join(emails, ","))),
	}

	if comment != "" {
		params.Comment = comment
	}

	type response struct {
		Result bool
	}

	var respData response
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}

// RemoveFromBlacklist removes an email addresses from a blacklist
func (service *BlacklistService) RemoveFromBlacklist(ctx context.Context, emails []string) error {
	path := "/blacklist"

	type paramsFormat struct {
		Emails string `json:"emails"`
	}

	params := paramsFormat{
		Emails: b64.StdEncoding.EncodeToString([]byte(strings.Join(emails, ","))),
	}

	type response struct {
		Result bool
	}

	var respData response
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, params, &respData, true)
	return err
}

// GetEmails returns a list of emails added to blacklist
func (service *BlacklistService) GetEmails(ctx context.Context) ([]string, error) {
	path := "/blacklist"

	var respData []string
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}
