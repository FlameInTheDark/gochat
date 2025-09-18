package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type DB struct {
	conn   *sqlx.DB
	logger *slog.Logger
}

func NewDB(logger *slog.Logger) *DB {
	return &DB{logger: logger}
}

func (db *DB) Connect(dsn string, maxRetries int) error {
	var err error
	var dbc *sql.DB

	for i := 0; i < maxRetries; i++ {
		dbc, err = sql.Open("postgres", dsn)
		if err == nil {
			break
		}
		db.logger.Warn("DB connect attempt failed", slog.Int("attempt", i+1), slog.String("error", err.Error()))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to DB after %d attempts: %w", maxRetries, err)
	}

	customLogger := &SlogLogger{logger: db.logger}

	dbc = sqldblogger.OpenDriver(
		dsn,
		dbc.Driver(),
		customLogger,
		sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),
	)

	db.conn = sqlx.NewDb(dbc, "postgres")
	if db.conn == nil {
		return errors.New("failed to connect to DB")
	}

	// Connection pool settings
	db.conn.SetMaxOpenConns(25)           // max open connections
	db.conn.SetMaxIdleConns(5)            // max idle connections
	db.conn.SetConnMaxLifetime(time.Hour) // recycle after 1 hour

	db.logger.Info("Postgres DB connected")

	return nil
}

func (db *DB) Conn() *sqlx.DB {
	return db.conn
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) Ping(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.conn.PingContext(ctx)
}
