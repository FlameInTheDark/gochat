package pgdb

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/helper"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type SlogLogger struct {
	logger *slog.Logger
}

func (l *SlogLogger) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	var slogLevel slog.Level
	switch level {
	case sqldblogger.LevelDebug:
		slogLevel = slog.LevelDebug
	case sqldblogger.LevelInfo:
		slogLevel = slog.LevelInfo
	case sqldblogger.LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	// Convert driver-provided data to slog attrs
	attrs := make([]slog.Attr, 0, len(data)+2)
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}
	// Enrich with context attributes (e.g., request_id)
	if extra := helper.AttrsFromContext(ctx); len(extra) > 0 {
		attrs = append(attrs, extra...)
	}
	l.logger.LogAttrs(ctx, slogLevel, msg, attrs...)
}
