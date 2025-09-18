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

// EmailTemplateData preserves the legacy data structure used by the
// built-in registration/password reset templates.
// New templates can use any custom data via RenderNamed.
type EmailTemplateData struct {
	AppName string
	BaseURL string
	UserId  int64
	Token   string
	Year    int
}

type Template struct {
	// generic named templates storage
	templates map[string]*template.Template

	// legacy fields to build default data
	appName string
	baseUrl string
	year    int
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

	t := &Template{
		templates: map[string]*template.Template{
			"registration":   regTmpl,
			"password_reset": pwdTmpl,
		},
		appName: appName,
		baseUrl: baseUrl,
		year:    year,
	}
	return t, nil
}

// AddTemplate parses and registers an additional template under a given name.
func (t *Template) AddTemplate(name, filePath string) error {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("error parsing template '%s': %w", name, err)
	}
	if t.templates == nil {
		t.templates = make(map[string]*template.Template)
	}
	t.templates[name] = tmpl
	return nil
}

// RenderNamed renders a template by its name with arbitrary data.
func (t *Template) RenderNamed(name string, data any) (string, error) {
	tmpl, ok := t.templates[name]
	if !ok {
		return "", fmt.Errorf("unknown email template: %s", name)
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email template render error: %w", err)
	}
	return buf.String(), nil
}

// Render keeps backward compatibility with existing callers that expect
// registration/password reset emails. It internally maps to named templates
// and constructs legacy data.
func (t *Template) Render(id int64, token string, emailType EmailType) (string, error) {
	data := EmailTemplateData{
		AppName: t.appName,
		BaseURL: t.baseUrl,
		UserId:  id,
		Token:   token,
		Year:    t.year,
	}

	var name string
	switch emailType {
	case EmailTypeRegistration:
		name = "registration"
	case EmailTypePasswordReset:
		name = "password_reset"
	default:
		return "", fmt.Errorf("unknown email type: %d", emailType)
	}
	return t.RenderNamed(name, data)
}
