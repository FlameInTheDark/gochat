package mailer

import (
	"bytes"
	"fmt"
	"html/template"
)

type EmailTemplateData struct {
	AppName          string
	ConfirmationLink string
	Year             int
}

type Template struct {
	tmpl    *template.Template
	appName string
	baseUrl string
	year    int
}

func NewEmailTemplate(file, baseUrl, appName string, year int) (*Template, error) {
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}
	return &Template{
		tmpl:    tmpl,
		appName: appName,
		baseUrl: baseUrl,
		year:    year,
	}, nil
}

func (t *Template) Render(id int64, token string) (string, error) {
	buf := bytes.Buffer{}
	data := EmailTemplateData{
		t.appName,
		fmt.Sprintf("%s/%d/%s", t.baseUrl, id, token),
		t.year,
	}
	err := t.tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("email template render error: %w", err)
	}
	return buf.String(), nil
}
