package resendp

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"

	"github.com/FlameInTheDark/gochat/internal/mailer"
)

type ResendMailer struct {
	client *resend.Client
}

func New(key string) *ResendMailer {
	if key == "" {
		panic("Resend key is empty")
	}
	client := resend.NewClient(key)
	return &ResendMailer{client: client}
}

func (m *ResendMailer) Send(ctx context.Context, notify mailer.MailNotification) error {
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", notify.From.Name, notify.From.Email),
		To:      []string{notify.To.Email},
		Subject: notify.Subject,
		Html:    notify.Html,
	}
	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return err
	}
	return nil
}
