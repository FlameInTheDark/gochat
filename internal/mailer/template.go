package mailer

import (
	"bytes"
	"fmt"
	"html/template"
)

// EmailType defines the type of email to send
type EmailType int

const (
	// EmailTypeRegistration is for registration confirmation emails
	EmailTypeRegistration EmailType = iota
	// EmailTypePasswordReset is for password reset emails
	EmailTypePasswordReset
)

type EmailTemplateData struct {
	AppName          string
	ConfirmationLink string
	Year             int
}

type Template struct {
	registrationTmpl  *template.Template
	passwordResetTmpl *template.Template
	appName           string
	baseUrl           string
	year              int
}

func NewEmailTemplate(registrationFile, passwordResetFile, baseUrl, appName string, year int) (*Template, error) {
	regTmpl, err := template.ParseFiles(registrationFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing registration template: %w", err)
	}

	pwdTmpl, err := template.ParseFiles(passwordResetFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing password reset template: %w", err)
	}

	return &Template{
		registrationTmpl:  regTmpl,
		passwordResetTmpl: pwdTmpl,
		appName:           appName,
		baseUrl:           baseUrl,
		year:              year,
	}, nil
}

func (t *Template) Render(id int64, token string, emailType EmailType) (string, error) {
	buf := bytes.Buffer{}
	data := EmailTemplateData{
		t.appName,
		fmt.Sprintf("%s/%d/%s", t.baseUrl, id, token),
		t.year,
	}

	var tmpl *template.Template
	switch emailType {
	case EmailTypeRegistration:
		tmpl = t.registrationTmpl
	case EmailTypePasswordReset:
		tmpl = t.passwordResetTmpl
	default:
		return "", fmt.Errorf("unknown email type: %d", emailType)
	}

	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("email template render error: %w", err)
	}
	return buf.String(), nil
}
