package sendpulse_sdk_go

type EmailsService struct {
	client       *Client
	MailingLists *MailingListsService
	Templates    *TemplatesService
	Senders      *SendersService
	Blacklist    *BlacklistService
	Webhooks     *WebhooksService
	Address      *AddressService
	Campaigns    *CampaignsService
	Validator    *ValidatorService
}

func newEmailsService(cl *Client) *EmailsService {
	return &EmailsService{
		client:       cl,
		MailingLists: newMailingListsService(cl),
		Campaigns:    newCampaignsService(cl),
		Templates:    newTemplatesService(cl),
		Senders:      newSendersService(cl),
		Address:      newAddressService(cl),
		Blacklist:    newBlacklistService(cl),
		Webhooks:     newWebhooksService(cl),
		Validator:    newValidatorService(cl),
	}
}
