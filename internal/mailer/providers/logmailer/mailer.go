package logmailer

import (
	"context"
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/observability"
)

type LogMailer struct {
	logger *slog.Logger
}

func New(log *slog.Logger) *LogMailer {
	if log == nil {
		return &LogMailer{logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}
	}
	return &LogMailer{logger: log}
}

func (m *LogMailer) Send(ctx context.Context, notify mailer.MailNotification) error {
	m.logger.Info(
		"mail sent",
		slog.String("from_name", notify.From.Name),
		slog.String("from_email", observability.RedactEmail(notify.From.Email)),
		slog.String("to_name", notify.To.Name),
		slog.String("to_email", observability.RedactEmail(notify.To.Email)),
		slog.String("subject", notify.Subject),
		slog.Int("html_bytes", len(notify.Html)))
	return nil
}
