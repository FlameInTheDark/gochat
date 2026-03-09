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
	// Create base driver handle (does not actually establish a network connection).
	base, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open postgres driver: %w", err)
	}

	// Wrap with query logger
	customLogger := &SlogLogger{logger: db.logger}
	wrapped := sqldblogger.OpenDriver(
		dsn,
		base.Driver(),
		customLogger,
		sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),
	)
	// base handle is no longer needed after wrapping
	_ = base.Close()

	db.conn = sqlx.NewDb(wrapped, "postgres")
	if db.conn == nil {
		return errors.New("failed to initialize DB handle")
	}

	// Connection pool settings – tuned for Citus fan-out queries
	db.conn.SetMaxOpenConns(50)
	db.conn.SetMaxIdleConns(25)
	db.conn.SetConnMaxLifetime(30 * time.Minute)
	db.conn.SetConnMaxIdleTime(5 * time.Minute)

	// Ensure the database is reachable. Retry with a 5s delay to avoid
	// container restarts when DB is not yet ready. If maxRetries <= 0,
	// retry indefinitely until success.
	attempt := 0
	for {
		attempt++
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = db.conn.PingContext(ctx)
		cancel()
		if err == nil {
			break
		}

		db.logger.Warn(
			"Postgres ping failed; retrying",
			slog.Int("attempt", attempt),
			slog.String("error", err.Error()),
		)
		if maxRetries > 0 && attempt >= maxRetries {
			return fmt.Errorf("failed to connect to DB after %d attempts: %w", attempt, err)
		}
		time.Sleep(5 * time.Second)
	}

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
