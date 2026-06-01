package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
	_ "github.com/yugabyte/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type DB struct {
	conn             *sqlx.DB
	logger           *slog.Logger
	pingFn           func(context.Context) error
	probeMetrics     *postgresProbeMetrics
	probeMetricsMu   sync.Mutex
	probeLoopStarted atomic.Bool
}

const (
	defaultProbeInterval = 30 * time.Second
	defaultProbeTimeout  = 3 * time.Second
)

type postgresProbeMetrics struct {
	attrs          []attribute.KeyValue
	statusGauge    metric.Int64ObservableGauge
	successCounter metric.Int64Counter
	failureCounter metric.Int64Counter
	duration       metric.Float64Histogram
	lastSuccess    metric.Int64ObservableGauge
	registration   metric.Registration
	statusValue    atomic.Int64
	lastSuccessAt  atomic.Int64
}

func NewDB(logger *slog.Logger) *DB {
	return &DB{logger: logger}
}

// ConnectOptions configures the PostgreSQL connection. Zero values use defaults.
type ConnectOptions struct {
	DriverName   string // "postgres" for lib/pq, "pgx" for yugabyte/pgx
	MaxRetries   int    // 0 → unlimited retries
	QueryLog     bool   // emit individual queries at debug level
	MaxOpenConns int    // default 50  — set lower when running many replicas
	MaxIdleConns int    // default 25
}

func (db *DB) Connect(dsn string, opts ConnectOptions) error {
	return db.ConnectContext(context.Background(), dsn, opts)
}

func (db *DB) ConnectContext(ctx context.Context, dsn string, opts ConnectOptions) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	connectCtx, finishConnect := observability.StartDependencySpan(
		ctx,
		"postgres",
		"connect",
		"primary",
	)
	var connectErr error
	defer func() {
		finishConnect(connectErr)
	}()

	driverName := opts.DriverName
	if driverName == "" {
		driverName = "postgres"
	}
	if strings.EqualFold(driverName, "pgx") {
		dsn = withPGXExecMode(dsn)
	}

	// Create base driver handle (does not actually establish a network connection).
	base, err := sql.Open(driverName, dsn)
	if err != nil {
		connectErr = err
		return fmt.Errorf("failed to open %s driver: %w", driverName, err)
	}

	// Wrap with query logger.
	// WithQueryerLevel/ExecerLevel/PreparerLevel set the level for SUCCESSFUL calls.
	// WithMinimumLevel is the actual filter — only events >= minLevel reach the logger.
	//   QueryLog=false (default): suppress successful queries, only surface SQL errors.
	//   QueryLog=true:            log everything at Debug for troubleshooting.
	minLevel := sqldblogger.LevelError
	if opts.QueryLog {
		minLevel = sqldblogger.LevelDebug
	}
	customLogger := &SlogLogger{logger: db.logger}
	wrapped := sqldblogger.OpenDriver(
		dsn,
		base.Driver(),
		customLogger,
		sqldblogger.WithMinimumLevel(minLevel),
		sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),
	)
	// base handle is no longer needed after wrapping
	_ = base.Close()

	db.conn = sqlx.NewDb(wrapped, "postgres")
	if db.conn == nil {
		connectErr = errors.New("failed to initialize DB handle")
		return connectErr
	}

	// Connection pool settings. Defaults are conservative for multi-replica deployments;
	// set PG_MAX_OPEN_CONNS = floor(pg_max_connections / replica_count).
	maxOpen := opts.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 50
	}
	maxIdle := opts.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 25
	}
	db.conn.SetMaxOpenConns(maxOpen)
	db.conn.SetMaxIdleConns(maxIdle)
	db.conn.SetConnMaxLifetime(30 * time.Minute)
	db.conn.SetConnMaxIdleTime(5 * time.Minute)

	// Ensure the database is reachable. Retry with a 5s delay to avoid
	// container restarts when DB is not yet ready. If maxRetries <= 0,
	// retry indefinitely until success.
	attempt := 0
	for {
		attempt++
		ctx, cancel := context.WithTimeout(connectCtx, 3*time.Second)
		pingCtx, finishPing := observability.StartDependencySpan(ctx, "postgres", "ping", "primary", attribute.Int("attempt", attempt))
		err = db.conn.PingContext(pingCtx)
		cancel()
		finishPing(err)
		if err == nil {
			break
		}

		db.logger.Warn(
			"Postgres ping failed; retrying",
			slog.Int("attempt", attempt),
			slog.String("error", err.Error()),
		)
		if opts.MaxRetries > 0 && attempt >= opts.MaxRetries {
			connectErr = err
			return fmt.Errorf("failed to connect to DB after %d attempts: %w", attempt, err)
		}
		select {
		case <-connectCtx.Done():
			connectErr = connectCtx.Err()
			return fmt.Errorf("connect to DB canceled: %w", connectErr)
		case <-time.After(5 * time.Second):
		}
	}

	db.registerPoolMetrics()
	db.logger.Info("Postgres DB connected")
	return nil
}

func withPGXExecMode(dsn string) string {
	if strings.Contains(dsn, "default_query_exec_mode=") {
		return dsn
	}

	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return dsn
	}

	if u, err := url.Parse(trimmed); err == nil && u.Scheme != "" && u.Host != "" {
		q := u.Query()
		q.Set("default_query_exec_mode", "exec")
		u.RawQuery = q.Encode()
		return u.String()
	}

	return trimmed + " default_query_exec_mode=exec"
}

func (db *DB) Conn() *sqlx.DB {
	return db.conn
}

func (db *DB) Close() error {
	if db != nil {
		db.closeProbeMetrics()
	}
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) Ping(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ctx, end := observability.StartDependencySpan(ctx, "postgres", "ping", "primary")
	err := db.pingContext(ctx)
	end(err)
	return err
}

func (db *DB) StartProbeLoop(ctx context.Context, interval time.Duration) {
	if db == nil || !db.probeLoopStarted.CompareAndSwap(false, true) {
		return
	}
	if interval <= 0 {
		interval = defaultProbeInterval
	}
	if ctx == nil {
		return
	}
	baseCtx := ctx

	_ = db.runProbe(baseCtx, defaultProbeTimeout)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = db.runProbe(baseCtx, defaultProbeTimeout)
		}
	}
}

func (db *DB) registerPoolMetrics() {
	if db == nil || db.conn == nil {
		return
	}
	meter := observability.Meter("gochat/postgres")
	pool := db.conn.DB
	if pool == nil {
		return
	}

	openGauge, _ := meter.Int64ObservableGauge("gochat.postgres.pool.open_connections")
	inUseGauge, _ := meter.Int64ObservableGauge("gochat.postgres.pool.in_use")
	idleGauge, _ := meter.Int64ObservableGauge("gochat.postgres.pool.idle")
	waitCountGauge, _ := meter.Int64ObservableGauge("gochat.postgres.pool.wait_count")
	waitDurationGauge, _ := meter.Float64ObservableGauge("gochat.postgres.pool.wait_duration")

	_, _ = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		stats := pool.Stats()
		observer.ObserveInt64(openGauge, int64(stats.OpenConnections))
		observer.ObserveInt64(inUseGauge, int64(stats.InUse))
		observer.ObserveInt64(idleGauge, int64(stats.Idle))
		observer.ObserveInt64(waitCountGauge, stats.WaitCount)
		observer.ObserveFloat64(waitDurationGauge, stats.WaitDuration.Seconds())
		return nil
	}, openGauge, inUseGauge, idleGauge, waitCountGauge, waitDurationGauge)
}

func (db *DB) pingContext(ctx context.Context) error {
	if db == nil {
		return errors.New("postgres db is nil")
	}
	if db.pingFn != nil {
		return db.pingFn(ctx)
	}
	if db.conn == nil {
		return errors.New("postgres db is not connected")
	}
	return db.conn.PingContext(ctx)
}

func (db *DB) runProbe(ctx context.Context, timeout time.Duration) error {
	if db == nil {
		return errors.New("postgres db is nil")
	}
	if timeout <= 0 {
		timeout = defaultProbeTimeout
	}

	recordCtx := observability.BackgroundFromContext(ctx)
	probeCtx, cancel := context.WithTimeout(recordCtx, timeout)
	started := time.Now()
	err := db.pingContext(probeCtx)
	cancel()

	db.ensureProbeMetrics().record(recordCtx, time.Since(started), err)
	return err
}

func (db *DB) ensureProbeMetrics() *postgresProbeMetrics {
	db.probeMetricsMu.Lock()
	defer db.probeMetricsMu.Unlock()

	if db.probeMetrics != nil {
		return db.probeMetrics
	}

	meter := observability.Meter("gochat/postgres")
	attrs := []attribute.KeyValue{
		attribute.String("postgres.target", "primary"),
		attribute.String("probe", "ping"),
	}

	statusGauge, _ := meter.Int64ObservableGauge("gochat.postgres.probe.status")
	successCounter, _ := meter.Int64Counter("gochat.postgres.probe.success")
	failureCounter, _ := meter.Int64Counter("gochat.postgres.probe.failure")
	duration, _ := meter.Float64Histogram("gochat.postgres.probe.duration")
	lastSuccess, _ := meter.Int64ObservableGauge("gochat.postgres.probe.last_success_unix")

	probe := &postgresProbeMetrics{
		attrs:          attrs,
		statusGauge:    statusGauge,
		successCounter: successCounter,
		failureCounter: failureCounter,
		duration:       duration,
		lastSuccess:    lastSuccess,
	}
	probe.statusValue.Store(0)
	registration, _ := meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(probe.statusGauge, probe.statusValue.Load(), metric.WithAttributes(probe.attrs...))
		observer.ObserveInt64(probe.lastSuccess, probe.lastSuccessAt.Load(), metric.WithAttributes(probe.attrs...))
		return nil
	}, statusGauge, lastSuccess)
	probe.registration = registration
	db.probeMetrics = probe
	return db.probeMetrics
}

func (db *DB) closeProbeMetrics() {
	db.probeMetricsMu.Lock()
	defer db.probeMetricsMu.Unlock()

	if db.probeMetrics == nil || db.probeMetrics.registration == nil {
		return
	}
	_ = db.probeMetrics.registration.Unregister()
	db.probeMetrics.registration = nil
}

func (m *postgresProbeMetrics) record(ctx context.Context, duration time.Duration, err error) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(m.attrs...)
	m.duration.Record(ctx, duration.Seconds(), attrs)
	if err != nil {
		m.statusValue.Store(0)
		m.failureCounter.Add(ctx, 1, attrs)
		return
	}
	m.statusValue.Store(1)
	m.lastSuccessAt.Store(time.Now().Unix())
	m.successCounter.Add(ctx, 1, attrs)
}
