package mailer

import (
	"context"
)

type MailNotification struct {
	From    User
	To      User
	Html    string
	Subject string
}

type Provider interface {
	Send(ctx context.Context, notify MailNotification) error
}

type User struct {
	Email string
	Name  string
}

type Mailer struct {
	provider Provider
	template *Template
	from     User
}

func NewMailer(provider Provider, template *Template, from User) *Mailer {
	return &Mailer{provider: provider, template: template, from: from}
}

// Send preserves legacy behavior for registration/password reset emails.
// For new email types, prefer SendTemplate which supports arbitrary templates and data.
func (m Mailer) Send(ctx context.Context, id int64, email, token string, emailType EmailType) error {
	html, err := m.template.Render(id, token, emailType)
	if err != nil {
		return err
	}
	// best-effort default subjects for legacy paths
	var subject string
	switch emailType {
	case EmailTypeRegistration:
		subject = "Email confirmation"
	case EmailTypePasswordReset:
		subject = "Password reset"
	}
	return m.provider.Send(ctx, MailNotification{
		From:    m.from,
		To:      User{Email: email},
		Html:    html,
		Subject: subject,
	})
}

// SendTemplate sends an email using a named template and arbitrary data
// and allows specifying any subject string.
func (m Mailer) SendTemplate(ctx context.Context, to User, subject, templateName string, data any) error {
	html, err := m.template.RenderNamed(templateName, data)
	if err != nil {
		return err
	}
	return m.provider.Send(ctx, MailNotification{
		From:    m.from,
		To:      to,
		Html:    html,
		Subject: subject,
	})
}
