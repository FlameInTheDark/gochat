package logmailer

import (
	"context"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"log/slog"
	"os"
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
		"Mail sent",
		slog.String("from", notify.From.Name+": "+notify.From.Email),
		slog.String("to", notify.To.Name+": "+notify.To.Email),
		slog.String("html", notify.Html))
	return nil
}
