package sendpulse_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SmsService struct {
	client *Client
}

func newSmsService(cl *Client) *SmsService {
	return &SmsService{client: cl}
}

type SmsVariable struct {
	Name  string      `json:"name"`
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value"`
}

type AddPhonesCounters struct {
	Added      int `json:"added"`
	Exceptions int `json:"exceptions"`
	Exists     int `json:"exists"`
}

func (service *SmsService) AddPhones(ctx context.Context, mailingListID int, phones []string) (*AddPhonesCounters, error) {
	path := "/sms/numbers"
	type paramsFormat struct {
		AddressBookID int      `json:"addressBookId"`
		Phones        []string `json:"phones"`
	}

	data := paramsFormat{AddressBookID: mailingListID, Phones: phones}

	var respData struct {
		Result   bool               `json:"result"`
		Counters *AddPhonesCounters `json:"counters"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &respData, true)
	return respData.Counters, err
}

type PhoneWithVariable struct {
	Phone     string
	Variables []SmsVariable
}

func (service *SmsService) AddPhonesWithVariables(ctx context.Context, mailingListID int, phones []*PhoneWithVariable) (*AddPhonesCounters, error) {
	path := "/sms/numbers/variables"
	type paramsFormat struct {
		AddressBookID int                        `json:"addressBookId"`
		Phones        map[string][][]SmsVariable `json:"phones"`
	}

	ph := make(map[string][][]SmsVariable)
	for _, item := range phones {
		ph[item.Phone] = append(ph[item.Phone], item.Variables)
	}

	data := paramsFormat{
		AddressBookID: mailingListID,
		Phones:        ph,
	}

	var respData struct {
		Result   bool               `json:"result"`
		Counters *AddPhonesCounters `json:"counters"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &respData, true)
	return respData.Counters, err
}

func (service *SmsService) UpdateVariablesSingle(ctx context.Context, addressBookID int, phone string, variables []SmsVariable) error {
	path := fmt.Sprintf("/addressbooks/%d/phones/variable", addressBookID)
	type paramsFormat struct {
		Phone     string        `json:"phone"`
		Variables []SmsVariable `json:"variables"`
	}

	data := paramsFormat{Phone: phone, Variables: variables}

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &respData, true)
	return err
}

func (service *SmsService) UpdateVariablesMultiple(ctx context.Context, addressBookID int, phones []string, variables []SmsVariable) error {
	path := "/sms/numbers"
	type paramsFormat struct {
		AddressBookID int           `json:"addressBookId"`
		Phones        []string      `json:"phones"`
		Variables     []SmsVariable `json:"variables"`
	}

	data := paramsFormat{
		AddressBookID: addressBookID,
		Phones:        phones,
		Variables:     variables,
	}

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPut, path, data, &respData, true)
	return err
}

func (service *SmsService) DeletePhones(ctx context.Context, addressBookID int, phones []string) error {
	path := "/sms/numbers"
	type paramsFormat struct {
		AddressBookID int      `json:"addressBookId"`
		Phones        []string `json:"phones"`
	}

	data := paramsFormat{
		AddressBookID: addressBookID,
		Phones:        phones,
	}

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, data, &respData, true)
	return err
}

type PhoneInfo struct {
	Status    int                    `json:"status"`
	Variables map[string]interface{} `json:"variables"`
	Added     DateTimeType           `json:"added"`
}

func (service *SmsService) GetPhoneInfo(ctx context.Context, addressBookID int, phone string) (*PhoneInfo, error) {
	path := fmt.Sprintf("/sms/numbers/info/%d/%s", addressBookID, phone)
	var respData struct {
		Result bool       `json:"result"`
		Data   *PhoneInfo `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *SmsService) AddToBlacklist(ctx context.Context, phones []string, description string) error {
	path := "/sms/black_list"

	type paramsFormat struct {
		Description string   `json:"description"`
		Phones      []string `json:"phones"`
	}

	data := paramsFormat{
		Description: description,
		Phones:      phones,
	}

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, data, &respData, true)
	return err
}

func (service *SmsService) RemoveFromBlacklist(ctx context.Context, phones []string) error {
	path := "/sms/black_list"

	type paramsFormat struct {
		Phones []string `json:"phones"`
	}

	data := paramsFormat{
		Phones: phones,
	}

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, data, &respData, true)
	return err
}

type BlacklistPhone struct {
	Phone       string
	Description string       `json:"description"`
	AddDate     DateTimeType `json:"add_date"`
}

func (service *SmsService) GetBlacklistedPhones(ctx context.Context, phones []string) ([]*BlacklistPhone, error) {
	path := "/sms/black_list/by_numbers"
	urlParams := url.Values{}
	urlParams.Add("phones", "["+strings.Join(phones, ",")+"]")
	path += "?" + urlParams.Encode()

	type BlacklistPhoneInternal struct {
		BlacklistPhone
		Phone int `json:"phone"`
	}

	var respData struct {
		Result bool                      `json:"result"`
		Data   []*BlacklistPhoneInternal `json:"data"`
	}

	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	if err != nil {
		return nil, err
	}

	data := make([]*BlacklistPhone, len(respData.Data))
	for i, item := range respData.Data {
		item.BlacklistPhone.Phone = strconv.Itoa(item.Phone)
		data[i] = &item.BlacklistPhone
	}

	return data, nil
}

type CreateSmsCampaignByAddressBookParams struct {
	Sender        string            `json:"sender"`
	MailingListID int               `json:"addressBookId"`
	Body          string            `json:"body"`
	Transliterate int               `json:"transliterate"`
	Route         map[string]string `json:"route,omitempty"`
	Date          DateTimeType      `json:"date"`
	Emulate       int               `json:"emulate"`
}

func (service *SmsService) CreateCampaignByMailingList(ctx context.Context, params CreateSmsCampaignByAddressBookParams) (int, error) {
	path := "/sms/campaigns"

	var respData struct {
		Result     bool `json:"result"`
		CampaignID int  `json:"campaign_id"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.CampaignID, err
}

type CreateSmsCampaignByPhonesParams struct {
	Sender        string            `json:"sender"`
	Phones        []string          `json:"phones"`
	Body          string            `json:"body"`
	Transliterate int               `json:"transliterate"`
	Route         map[string]string `json:"route,omitempty"`
	Date          DateTimeType      `json:"date"`
	Emulate       int               `json:"emulate"`
}

func (service *SmsService) CreateCampaignByPhones(ctx context.Context, params CreateSmsCampaignByPhonesParams) (int, error) {
	path := "/sms/send"

	var respData struct {
		Result     bool `json:"result"`
		CampaignID int  `json:"campaign_id"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPost, path, params, &respData, true)
	return respData.CampaignID, err
}

type SmsCampaign struct {
	ID                int          `json:"id"`
	AddressBookID     int          `json:"address_book_id"`
	CompanyPrice      float32      `json:"company_price"`
	CompanyCurrency   string       `json:"company_currency"`
	SendDate          DateTimeType `json:"send_date"`
	DateCreated       DateTimeType `json:"date_created"`
	SenderMailAddress string       `json:"sender_mail_address"`
	SenderMailName    string       `json:"sender_mail_name"`
}

func (service *SmsService) GetCampaigns(ctx context.Context, dateFrom, dateTo time.Time) ([]*SmsCampaign, error) {
	dtFormat := "2006-01-02 15:04:05"
	path := "/sms/campaigns/list"
	urlParams := url.Values{}
	urlParams.Add("dateFrom", dateFrom.Format(dtFormat))
	urlParams.Add("dateTo", dateTo.Format(dtFormat))
	path += "?" + urlParams.Encode()

	var respData struct {
		Result bool           `json:"result"`
		Data   []*SmsCampaign `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type SmsCampaignInfo struct {
	ID            int          `json:"id"`
	AddressBookID int          `json:"address_book_id"`
	Currency      string       `json:"currency"`
	CompanyPrice  string       `json:"company_price"`
	SendDate      DateTimeType `json:"send_date"`
	DateCreated   DateTimeType `json:"date_created"`
	SenderName    string       `json:"sender_name"`
	PhonesInfo    []struct {
		Phone         int     `json:"phone"`
		Status        int     `json:"status"`
		StatusExplain string  `json:"status_explain"`
		CountryCode   string  `json:"сountry_code"`
		MoneySpent    float32 `json:"money_spent"`
	} `json:"task_phones_info"`
}

func (service *SmsService) GetCampaignInfo(ctx context.Context, id int) (*SmsCampaignInfo, error) {
	path := fmt.Sprintf("/sms/campaigns/info/%d", id)

	var respData struct {
		Result bool             `json:"result"`
		Data   *SmsCampaignInfo `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

func (service *SmsService) CancelCampaign(ctx context.Context, id int) error {
	path := fmt.Sprintf("/sms/campaigns/cancel/%d", id)

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodPut, path, nil, &respData, true)
	return err
}

type SmsCampaignCostParams struct {
	AddressBookID int
	Phones        []string
	Body          string
	Sender        string
	Route         map[string]string
}

type SmsCampaignCampaignCost struct {
	Price    float32 `json:"price"`
	Currency string  `json:"currency"`
}

func (service *SmsService) GetCampaignCost(ctx context.Context, params SmsCampaignCostParams) (*SmsCampaignCampaignCost, error) {
	path := "/sms/campaigns/cost"
	urlParams := url.Values{}
	if params.AddressBookID != 0 {
		urlParams.Add("addressBookId", strconv.Itoa(params.AddressBookID))
	}
	if len(params.Phones) != 0 {
		urlParams.Add("phones", "["+strings.Join(params.Phones, ",")+"]")
	}
	if params.Body != "" {
		urlParams.Add("body", params.Body)
	}
	if params.Sender != "" {
		urlParams.Add("sender", params.Sender)
	}
	if len(params.Route) != 0 {
		route, _ := json.Marshal(params.Route)
		urlParams.Add("route", string(route))
	}
	if len(urlParams) != 0 {
		path += "?" + urlParams.Encode()
	}
	var respData struct {
		Result bool                     `json:"result"`
		Data   *SmsCampaignCampaignCost `json:"data"`
	}
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData.Data, err
}

type SmsSender struct {
	ID            int    `json:"id"`
	Sender        string `json:"sender"`
	Country       string `json:"country"`
	CountryCode   string `json:"country_code"`
	Status        int    `json:"status"`
	StatusExplain string `json:"status_explain"`
}

func (service *SmsService) GetSenders(ctx context.Context) ([]*SmsSender, error) {
	path := "/sms/senders"

	var respData []*SmsSender
	_, err := service.client.newRequest(ctx, http.MethodGet, path, nil, &respData, true)
	return respData, err
}

func (service *SmsService) DeleteCampaign(ctx context.Context, id int) error {
	path := "/sms/campaigns"

	type paramsFormat struct {
		ID int `json:"id"`
	}

	data := paramsFormat{
		ID: id,
	}

	var respData struct {
		Result bool `json:"result"`
	}
	_, err := service.client.newRequest(ctx, http.MethodDelete, path, data, &respData, true)
	return err
}
