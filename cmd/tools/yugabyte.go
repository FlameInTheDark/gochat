package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	cli "github.com/urfave/cli/v3"
)

type yugabyteTableSnapshot struct {
	Table    string `json:"table"`
	Rows     int64  `json:"rows"`
	Checksum string `json:"checksum,omitempty"`
}

type yugabyteTableComparison struct {
	Table          string `json:"table"`
	SourceRows     int64  `json:"source_rows"`
	TargetRows     int64  `json:"target_rows"`
	SourceChecksum string `json:"source_checksum,omitempty"`
	TargetChecksum string `json:"target_checksum,omitempty"`
	Status         string `json:"status"`
	Reason         string `json:"reason,omitempty"`
}

type yugabyteVerifyResult struct {
	Schema   string                    `json:"schema"`
	Checksum bool                      `json:"checksum"`
	OK       bool                      `json:"ok"`
	Tables   []yugabyteTableComparison `json:"tables"`
}

func yugabyte() *cli.Command {
	return &cli.Command{
		Name:  "yugabyte",
		Usage: "YugabyteDB operational helpers",
		Commands: []*cli.Command{
			yugabyteVerifyCommand(),
		},
	}
}

func yugabyteVerifyCommand() *cli.Command {
	return &cli.Command{
		Name:  "verify",
		Usage: "Compare two YugabyteDB YSQL databases by table counts and checksums",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "source-dsn", Usage: "source YugabyteDB YSQL DSN", Required: true},
			&cli.StringFlag{Name: "target-dsn", Usage: "target YugabyteDB YSQL DSN", Required: true},
			&cli.StringFlag{Name: "schema", Value: "public", Usage: "schema to compare"},
			&cli.BoolFlag{Name: "skip-checksum", Usage: "compare row counts only"},
			&cli.DurationFlag{Name: "timeout", Value: 2 * time.Minute, Usage: "verification timeout"},
			&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Value: "text", Usage: "output: text|json"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			timeout := cmd.Duration("timeout")
			if timeout <= 0 {
				return fmt.Errorf("timeout must be greater than zero")
			}
			verifyCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			result, err := runYugabyteVerify(verifyCtx, yugabyteVerifyOptions{
				SourceDSN: cmd.String("source-dsn"),
				TargetDSN: cmd.String("target-dsn"),
				Schema:    cmd.String("schema"),
				Checksum:  !cmd.Bool("skip-checksum"),
			})
			if err != nil {
				return err
			}
			if err := printYugabyteVerifyResult(os.Stdout, cmd.String("format"), result); err != nil {
				return err
			}
			if !result.OK {
				return fmt.Errorf("yugabyte verification failed")
			}
			return nil
		},
	}
}

type yugabyteVerifyOptions struct {
	SourceDSN string
	TargetDSN string
	Schema    string
	Checksum  bool
}

func runYugabyteVerify(ctx context.Context, opts yugabyteVerifyOptions) (yugabyteVerifyResult, error) {
	if strings.TrimSpace(opts.SourceDSN) == "" {
		return yugabyteVerifyResult{}, fmt.Errorf("source-dsn is required")
	}
	if strings.TrimSpace(opts.TargetDSN) == "" {
		return yugabyteVerifyResult{}, fmt.Errorf("target-dsn is required")
	}
	if strings.TrimSpace(opts.Schema) == "" {
		return yugabyteVerifyResult{}, fmt.Errorf("schema is required")
	}

	source, err := sql.Open("postgres", opts.SourceDSN)
	if err != nil {
		return yugabyteVerifyResult{}, fmt.Errorf("open source database: %w", err)
	}
	defer source.Close()

	target, err := sql.Open("postgres", opts.TargetDSN)
	if err != nil {
		return yugabyteVerifyResult{}, fmt.Errorf("open target database: %w", err)
	}
	defer target.Close()

	if err := source.PingContext(ctx); err != nil {
		return yugabyteVerifyResult{}, fmt.Errorf("ping source database: %w", err)
	}
	if err := target.PingContext(ctx); err != nil {
		return yugabyteVerifyResult{}, fmt.Errorf("ping target database: %w", err)
	}

	sourceSnapshots, err := loadYugabyteTableSnapshots(ctx, source, opts.Schema, opts.Checksum)
	if err != nil {
		return yugabyteVerifyResult{}, fmt.Errorf("load source table snapshots: %w", err)
	}
	targetSnapshots, err := loadYugabyteTableSnapshots(ctx, target, opts.Schema, opts.Checksum)
	if err != nil {
		return yugabyteVerifyResult{}, fmt.Errorf("load target table snapshots: %w", err)
	}

	return compareYugabyteSnapshots(opts.Schema, opts.Checksum, sourceSnapshots, targetSnapshots), nil
}

func loadYugabyteTableSnapshots(ctx context.Context, db *sql.DB, schema string, checksum bool) (map[string]yugabyteTableSnapshot, error) {
	rows, err := db.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1
  AND table_type = 'BASE TABLE'
ORDER BY table_name`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	snapshots := make(map[string]yugabyteTableSnapshot, len(tables))
	for _, table := range tables {
		rowCount, err := loadYugabyteTableCount(ctx, db, schema, table)
		if err != nil {
			return nil, fmt.Errorf("count %s.%s: %w", schema, table, err)
		}

		snapshot := yugabyteTableSnapshot{Table: table, Rows: rowCount}
		if checksum {
			tableChecksum, err := loadYugabyteTableChecksum(ctx, db, schema, table)
			if err != nil {
				return nil, fmt.Errorf("checksum %s.%s: %w", schema, table, err)
			}
			snapshot.Checksum = tableChecksum
		}
		snapshots[table] = snapshot
	}

	return snapshots, nil
}

func loadYugabyteTableCount(ctx context.Context, db *sql.DB, schema, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", quoteYSQLIdent(schema), quoteYSQLIdent(table))
	var count int64
	if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func loadYugabyteTableChecksum(ctx context.Context, db *sql.DB, schema, table string) (string, error) {
	tableRef := fmt.Sprintf("%s.%s", quoteYSQLIdent(schema), quoteYSQLIdent(table))
	query := fmt.Sprintf(`
SELECT COALESCE(md5(string_agg(row_hash, '' ORDER BY row_hash)), md5(''))
FROM (
    SELECT md5(row_to_json(src)::text) AS row_hash
    FROM %s AS src
) AS row_hashes`, tableRef)

	var checksum string
	if err := db.QueryRowContext(ctx, query).Scan(&checksum); err != nil {
		return "", err
	}
	return checksum, nil
}

func compareYugabyteSnapshots(schema string, checksum bool, source, target map[string]yugabyteTableSnapshot) yugabyteVerifyResult {
	tableSet := make(map[string]struct{}, len(source)+len(target))
	for table := range source {
		tableSet[table] = struct{}{}
	}
	for table := range target {
		tableSet[table] = struct{}{}
	}

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	result := yugabyteVerifyResult{
		Schema:   schema,
		Checksum: checksum,
		OK:       true,
		Tables:   make([]yugabyteTableComparison, 0, len(tables)),
	}

	for _, table := range tables {
		sourceSnapshot, sourceOK := source[table]
		targetSnapshot, targetOK := target[table]
		comparison := yugabyteTableComparison{
			Table:          table,
			SourceRows:     sourceSnapshot.Rows,
			TargetRows:     targetSnapshot.Rows,
			SourceChecksum: sourceSnapshot.Checksum,
			TargetChecksum: targetSnapshot.Checksum,
			Status:         "ok",
		}

		switch {
		case !sourceOK:
			comparison.Status = "mismatch"
			comparison.Reason = "missing from source"
		case !targetOK:
			comparison.Status = "mismatch"
			comparison.Reason = "missing from target"
		case sourceSnapshot.Rows != targetSnapshot.Rows:
			comparison.Status = "mismatch"
			comparison.Reason = "row count differs"
		case checksum && sourceSnapshot.Checksum != targetSnapshot.Checksum:
			comparison.Status = "mismatch"
			comparison.Reason = "checksum differs"
		}

		if comparison.Status != "ok" {
			result.OK = false
		}
		result.Tables = append(result.Tables, comparison)
	}

	return result
}

func printYugabyteVerifyResult(w io.Writer, format string, result yugabyteVerifyResult) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "text":
		status := "FAILED"
		if result.OK {
			status = "OK"
		}
		if _, err := fmt.Fprintf(w, "verification=%s schema=%s checksum=%t tables=%d\n", status, result.Schema, result.Checksum, len(result.Tables)); err != nil {
			return err
		}
		for _, table := range result.Tables {
			if table.Status == "ok" {
				if _, err := fmt.Fprintf(w, "%s rows=%d status=ok\n", table.Table, table.SourceRows); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(
				w,
				"%s source_rows=%d target_rows=%d status=%s reason=%s\n",
				table.Table,
				table.SourceRows,
				table.TargetRows,
				table.Status,
				table.Reason,
			); err != nil {
				return err
			}
		}
		return nil
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func quoteYSQLIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
