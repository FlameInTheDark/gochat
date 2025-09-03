package sendpulse

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/FlameInTheDark/gochat/internal/mailer"

	sendpulse "github.com/dimuska139/sendpulse-sdk-go/v7"
)

type SendpulseMailer struct {
	c *sendpulse.Client
}

func New(userId, secret string) *SendpulseMailer {
	config := sendpulse.Config{
		UserID: userId,
		Secret: secret,
	}
	client := sendpulse.NewClient(http.DefaultClient, &config)
	return &SendpulseMailer{c: client}
}

func (m *SendpulseMailer) Send(ctx context.Context, notify mailer.MailNotification) error {
	params := sendpulse.SendEmailParams{
		Subject: notify.Subject,
		From: sendpulse.User{
			Name:  notify.From.Name,
			Email: notify.From.Email,
		},
		To: []sendpulse.User{
			{
				Name:  notify.To.Name,
				Email: notify.To.Email,
			},
		},
		Html: base64.StdEncoding.EncodeToString([]byte(notify.Html)),
	}
	_, err := m.c.SMTP.SendMessage(ctx, params)
	if err != nil {
		return err
	}
	return nil
}
