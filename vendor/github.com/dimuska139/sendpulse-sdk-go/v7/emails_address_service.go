package sendpulse_sdk_go

import (
	"context"
	"fmt"
	"net/http"
)

// AddressService is a service to work with emails addresses
type AddressService struct {
	client *Client
}

// newAddressService creates AddressService
func newAddressService(cl *Client) *AddressService {
	return &AddressService{client: cl}
}

// Variable represents a variable of email address
type Variable struct {
	Name  string      `json:"name"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value"`
}

// EmailInfo represents a general information of email address
type EmailInfo struct {
	BookID    int         `json:"book_id"`
	Status    int         `json:"status"`
	Variables []*Variable `json:"variables"`
}

// GetEmailInfo returns general information about specific email address
func (service *AddressService) GetEmailInfo(ctx context.Context, email string) ([]*EmailInfo, error) {
	path := fmt.Sprintf("/emails/%s", email)
	var response []*EmailInfo
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response, err
}

// GetEmailsInfo retrieves general informaion for a List of Email Addresses
func (service *AddressService) GetEmailsInfo(ctx context.Context, emails []string) (map[string][]*EmailInfo, error) {
	path := "/emails"
	type data struct {
		Emails []string `json:"emails"`
	}

	params := data{Emails: emails}
	respData := make(map[string][]*EmailInfo)
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData, err
}

// EmailInfoList represents a detailed information of email address
type EmailInfoList struct {
	ListName string       `json:"list_name"`
	ListID   int          `json:"list_id"`
	AddDate  DateTimeType `json:"add_date"`
	Source   string       `json:"source"`
}

// GetDetails retrieves detailed information about specific email address
func (service *AddressService) GetDetails(ctx context.Context, email string) ([]*EmailInfoList, error) {
	path := fmt.Sprintf("/emails/%s/details", email)
	var response []*EmailInfoList
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &response, true)
	return response, err
}

// GetStatisticsByCampaign returns information for a specific email address from a specific campaign
func (service *AddressService) GetStatisticsByCampaign(ctx context.Context, campaignID int, email string) (*CampaignEmailStatistics, error) {
	path := fmt.Sprintf("/campaigns/%d/email/%s", campaignID, email)
	var respData *CampaignEmailStatistics
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// AddressBookEmailStatistics represents statistics by specific address book
type AddressBookEmailStatistics struct {
	Email         string      `json:"email"`
	AddressBookID int         `json:"abook_id,string"`
	Status        int         `json:"status"`
	StatusExplain string      `json:"status_explain"`
	Variables     []*Variable `json:"variables"`
}

// CampaignEmailStatistics represents statistics of specific campaign
type CampaignEmailStatistics struct {
	SendDate            DateTimeType `json:"send_date"`
	GlobalStatus        int          `json:"global_status"`
	GlobalStatusExplain string       `json:"global_status_explain"`
	DetailStatus        int          `json:"detail_status"`
	DetailStatusExplain string       `json:"detail_status_explain"`
}

// GetStatisticsByAddressBook returns information for a specific email address from a specific address book
func (service *AddressService) GetStatisticsByAddressBook(ctx context.Context, addressBookID int, email string) (*AddressBookEmailStatistics, error) {
	path := fmt.Sprintf("/addressbooks/%d/emails/%s", addressBookID, email)
	var respData AddressBookEmailStatistics
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return &respData, err
}

// DeleteFromAllAddressBooks removes specific email address from all address books
func (service *AddressService) DeleteFromAllAddressBooks(ctx context.Context, email string) error {
	path := fmt.Sprintf("/emails/%s", email)
	var respData struct {
		Result bool
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, nil, &respData, true)
	return err
}

// CampaignsEmailStatistics represents statistics for an email address and campaigns it is in
type CampaignsEmailStatistics struct {
	Statistic *struct {
		Sent int `json:"sent"`
		Open int `json:"open"`
		Link int `json:"link"`
	} `json:"statistic"`
	Addressbooks []*struct {
		Id   int    `json:"id"`
		Name string `json:"address_book_name"`
	}
	Blacklist bool
}

// GetEmailStatisticsByCampaignsAndAddressBooks returns statistics for an email address and campaigns it is in
func (service *AddressService) GetEmailStatisticsByCampaignsAndAddressBooks(ctx context.Context, email string) (*CampaignsEmailStatistics, error) {
	path := fmt.Sprintf("/emails/%s/campaigns", email)
	var respData *CampaignsEmailStatistics
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

// CampaignsAndAddressBooksEmailStatistics represents statistics for multiple email addresses and campaigns they are in
type CampaignsAndAddressBooksEmailStatistics struct {
	Sent         int `json:"sent"`
	Open         int `json:"open"`
	Link         int `json:"link"`
	Addressbooks []*struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"adressbooks"`
	Blacklist bool
}

// GetEmailsStatisticsByCampaignsAndAddressBooks returns statistics for multiple email addresses and campaigns they are in
func (service *AddressService) GetEmailsStatisticsByCampaignsAndAddressBooks(ctx context.Context, emails []string) (map[string]*CampaignsAndAddressBooksEmailStatistics, error) {
	path := "/emails/campaigns"
	respData := make(map[string]*CampaignsAndAddressBooksEmailStatistics)

	type data struct {
		Emails []string `json:"emails"`
	}

	params := data{Emails: emails}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData, err
}

// ChangeVariables is a method for change a variable for an email contact
func (service *AddressService) ChangeVariables(ctx context.Context, addressBookID int, email string, variables []*Variable) error {
	path := fmt.Sprintf("/addressbooks/%d/emails/variable", addressBookID)

	type data struct {
		Email     string      `json:"email"`
		Variables []*Variable `json:"variables"`
	}

	params := data{Email: email, Variables: variables}
	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return err
}
