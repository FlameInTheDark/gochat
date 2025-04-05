package sendpulse_sdk_go

import (
	"context"
	"net/http"
	"strings"
)

// Automation360Service is a service to interact with user balance
type BalanceService struct {
	client *Client
}

// newBalanceService creates Automation360Service
func newBalanceService(cl *Client) *BalanceService {
	return &BalanceService{client: cl}
}

// Balance represents main information about user's balance
type Balance struct {
	Currency        string  `json:"currency"`
	BalanceCurrency float32 `json:"balance_currency"`
}

// Balance represents detailed information about user's balance
type BalanceDetailed struct {
	Balance struct {
		Main     float32 `json:"main,string"`
		Bonus    float32 `json:"bonus,string"`
		Currency string  `json:"currency"`
	} `json:"balance"`
	Email struct {
		TariffName         string       `json:"tariff_name"`
		FinishedTime       DateTimeType `json:"finished_time"`
		EmailsLeft         int          `json:"emails_left"`
		MaximumSubscribers int          `json:"maximum_subscribers"`
		CurrentSubscribers int          `json:"current_subscribers"`
	} `json:"email"`
	Smtp struct {
		TariffName string       `json:"tariff_name"`
		EndDate    DateTimeType `json:"end_date"`
		AutoRenew  int          `json:"auto_renew"`
	} `json:"smtp"`
	Push struct {
		TariffName string       `json:"tariff_name"`
		EndDate    DateTimeType `json:"end_date"`
		AutoRenew  int          `json:"auto_renew"`
	} `json:"push"`
}

// GetBalance returns main information about users's balance
func (service *BalanceService) GetBalance(ctx context.Context, currency string) (*Balance, error) {
	path := "/balance"
	if currency != "" {
		path += "/" + strings.ToLower(currency)
	}

	var respData Balance
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return &respData, err
}

// GetDetailedBalance returns detailed information about users's balance
func (service *BalanceService) GetDetailedBalance(ctx context.Context) (*BalanceDetailed, error) {
	path := "/user/balance/detail"

	var respData BalanceDetailed
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return &respData, err
}
