package mailer

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPasswordResetTemplateIncludesUserIDAndTokenInResetLink(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}

	root := filepath.Join(filepath.Dir(file), "..", "..")
	tmpl, err := NewEmailTemplate(
		filepath.Join(root, "email_notify.tmpl"),
		filepath.Join(root, "password_reset.tmpl"),
		"https://app.example.com",
		"GoChat",
		2026,
	)
	if err != nil {
		t.Fatalf("NewEmailTemplate() error = %v", err)
	}

	html, err := tmpl.Render(2230469276416868352, "just_a_random_text_from_email_012345678", EmailTypePasswordReset)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	expected := `href="https://app.example.com/reset/2230469276416868352/just_a_random_text_from_email_012345678"`
	if !strings.Contains(html, expected) {
		t.Fatalf("expected rendered template to contain %q", expected)
	}
}
