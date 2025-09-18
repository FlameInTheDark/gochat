package pgdb

import (
	"context"
	"log/slog"

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
	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}
	l.logger.LogAttrs(ctx, slogLevel, msg, attrs...)
}
