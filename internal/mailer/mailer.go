package mailer

import (
	"context"
)

type Provider interface {
	Send(ctx context.Context, notify MailNotification) error
}

type Mailer struct {
	provider Provider
	template *Template
	from     User
}

func NewMailer(provider Provider, template *Template, from User) *Mailer {
	return &Mailer{provider: provider, template: template, from: from}
}

func (m Mailer) Send(ctx context.Context, id int64, email, token string, emailType EmailType) error {
	html, err := m.template.Render(id, token, emailType)
	if err != nil {
		return err
	}
	return m.provider.Send(ctx, MailNotification{
		From: m.from,
		To: User{
			Email: email,
		},
		Html: html,
	})
}

type MailNotification struct {
	From User
	To   User
	Html string
}

type User struct {
	Email string
	Name  string
}
