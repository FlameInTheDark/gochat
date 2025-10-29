package dashamail

import (
	"context"
	"fmt"
	"net/http"

	"github.com/FlameInTheDark/gochat/internal/mailer"
	"resty.dev/v3"
)

const endpointFormat = "https://api.dashamail.com/?method=transactional.send&api_key=%s&format=JSON"

type Mailer struct {
	apiKey string
	client *resty.Client
}

func New(apiKey string) *Mailer {
	return &Mailer{apiKey: apiKey, client: resty.New().SetBaseURL(fmt.Sprintf(endpointFormat, apiKey))}
}

func (m *Mailer) Send(ctx context.Context, notify mailer.MailNotification) error {
	resp, err := m.client.R().SetBody(mailData{
		To:        notify.To.Email,
		Subject:   notify.Subject,
		FromEmail: notify.From.Email,
		Message:   notify.Html,
	}).
		WithContext(ctx).
		SetMethod(resty.MethodPost).
		Send()
	if err != nil {
		return fmt.Errorf("DashaMail send mail error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("DashaMail send mail status code: %d", resp.StatusCode())
	}
	return nil
}

type mailData struct {
	To        string `json:"to"`
	Subject   string `json:"subject"`
	FromEmail string `json:"from_email"`
	Message   string `json:"message"`
}
