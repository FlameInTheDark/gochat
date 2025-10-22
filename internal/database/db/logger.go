package db

import (
	"fmt"
	"log/slog"
)

type DBLogger struct {
	logger *slog.Logger
}

func NewDBLogger(logger *slog.Logger) *DBLogger {
	return &DBLogger{logger: logger.With(slog.String("module", "database"))}
}

func (l *DBLogger) Print(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}

func (l *DBLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}

func (l *DBLogger) Println(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}
