package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	apiconfig "github.com/FlameInTheDark/gochat/cmd/api/config"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gocql/gocql"
	"github.com/jmoiron/sqlx"
	cli "github.com/urfave/cli/v3"
)

const (
	selectChannelsForPositionBackfill = `
SELECT id, last_message, message_position
FROM channels
WHERE ($1::bigint = 0 OR id = $1::bigint)
  AND id > $2
ORDER BY id ASC
LIMIT $3::int`
	updateChannelMessagePosition = `
UPDATE channels
SET message_position = GREATEST(message_position, $2)
WHERE id = $1`
	selectBucketMessagePositions = `
SELECT id, position
FROM gochat.messages
WHERE channel_id = ? AND bucket = ?`
	updateMessagePosition = `
UPDATE gochat.messages
SET position = ?
WHERE channel_id = ? AND bucket = ? AND id = ?`
)

type channelPositionRow struct {
	ID              int64 `db:"id"`
	LastMessage     int64 `db:"last_message"`
	MessagePosition int64 `db:"message_position"`
}

type messagePositionRow struct {
	ID       int64
	Bucket   int64
	Position int64
}

type messagePositionUpdate struct {
	ID       int64
	Bucket   int64
	Position int64
}

type backfillPlan struct {
	Updates []messagePositionUpdate
	Cursor  int64
}

func messages() *cli.Command {
	return &cli.Command{
		Name:  "messages",
		Usage: "Tools to operate with messages",
		Commands: []*cli.Command{
			messagesBackfillPositions(),
		},
	}
}

func messagesBackfillPositions() *cli.Command {
	return &cli.Command{
		Name:  "backfill-positions",
		Usage: "Fill missing message.position values channel-by-channel and raise channels.message_position",
		Flags: []cli.Flag{
			&cli.Int64Flag{Name: "channel-id", Usage: "only process a single channel id"},
			&cli.IntFlag{Name: "channel-batch-size", Value: 128, Usage: "number of channels loaded from postgres per page"},
			&cli.IntFlag{Name: "write-batch-size", Value: 64, Usage: "number of Cassandra updates per unlogged batch"},
			&cli.BoolFlag{Name: "dry-run", Usage: "compute and print what would be changed without writing"},
			&cli.BoolFlag{Name: "rewrite", Usage: "recompute all message positions in a channel instead of only filling missing ones"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
			cfg, err := apiconfig.LoadConfig(logger)
			if err != nil {
				return err
			}

			cql, pg, err := openBackfillDatabases(logger, cfg)
			if err != nil {
				return err
			}
			defer func() { _ = cql.Close() }()
			defer func() { _ = pg.Close() }()

			channelID := cmd.Int64("channel-id")
			channelBatchSize := cmd.Int("channel-batch-size")
			if channelBatchSize <= 0 {
				return fmt.Errorf("channel-batch-size must be positive")
			}
			writeBatchSize := cmd.Int("write-batch-size")
			if writeBatchSize <= 0 {
				return fmt.Errorf("write-batch-size must be positive")
			}

			opts := backfillOptions{
				channelID:        channelID,
				channelBatchSize: channelBatchSize,
				writeBatchSize:   writeBatchSize,
				dryRun:           cmd.Bool("dry-run"),
				rewrite:          cmd.Bool("rewrite"),
			}
			return runMessagePositionBackfill(ctx, logger, cql, pg.Conn(), opts)
		},
	}
}

type backfillOptions struct {
	channelID        int64
	channelBatchSize int
	writeBatchSize   int
	dryRun           bool
	rewrite          bool
}

func openBackfillDatabases(logger *slog.Logger, cfg *apiconfig.Config) (*db.CQLCon, *pgdb.DB, error) {
	if len(cfg.Cluster) == 0 {
		return nil, nil, fmt.Errorf("CLUSTER must be configured")
	}
	if strings.TrimSpace(cfg.PGDSN) == "" {
		return nil, nil, fmt.Errorf("PG_DSN must be configured")
	}

	cql, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, nil, err
	}

	pg := pgdb.NewDB(logger)
	if err := pg.Connect(cfg.PGDSN, cfg.PGRetries); err != nil {
		_ = cql.Close()
		return nil, nil, err
	}

	return cql, pg, nil
}

func runMessagePositionBackfill(ctx context.Context, logger *slog.Logger, cql *db.CQLCon, pg *sqlx.DB, opts backfillOptions) error {
	var (
		afterID         int64
		totalChannels   int
		totalMessages   int
		totalUpdated    int
		totalCursorBump int
	)

	for {
		channels, err := loadChannelsForBackfill(ctx, pg, opts.channelID, afterID, opts.channelBatchSize)
		if err != nil {
			return err
		}
		if len(channels) == 0 {
			break
		}

		for _, channel := range channels {
			afterID = channel.ID
			totalChannels++

			if channel.LastMessage == 0 {
				continue
			}

			messages, err := loadChannelMessagesForBackfill(ctx, cql.Session(), channel.ID, channel.LastMessage)
			if err != nil {
				return fmt.Errorf("channel %d: %w", channel.ID, err)
			}
			totalMessages += len(messages)

			plan := planMessagePositionBackfill(messages, channel.MessagePosition, opts.rewrite)
			if len(plan.Updates) == 0 && plan.Cursor <= channel.MessagePosition {
				continue
			}

			if logger != nil {
				logger.Info("prepared channel backfill",
					slog.Int64("channel_id", channel.ID),
					slog.Int("messages", len(messages)),
					slog.Int("position_updates", len(plan.Updates)),
					slog.Int64("cursor_before", channel.MessagePosition),
					slog.Int64("cursor_after", plan.Cursor),
					slog.Bool("dry_run", opts.dryRun),
					slog.Bool("rewrite", opts.rewrite))
			}

			if opts.dryRun {
				totalUpdated += len(plan.Updates)
				if plan.Cursor > channel.MessagePosition {
					totalCursorBump++
				}
				continue
			}

			if err := applyMessagePositionBackfill(ctx, cql.Session(), pg, channel.ID, plan, opts.writeBatchSize); err != nil {
				return fmt.Errorf("channel %d: %w", channel.ID, err)
			}

			totalUpdated += len(plan.Updates)
			if plan.Cursor > channel.MessagePosition {
				totalCursorBump++
			}
		}

		if opts.channelID != 0 || len(channels) < opts.channelBatchSize {
			break
		}
	}

	if logger != nil {
		logger.Info("message position backfill complete",
			slog.Int("channels_scanned", totalChannels),
			slog.Int("messages_scanned", totalMessages),
			slog.Int("messages_updated", totalUpdated),
			slog.Int("channels_with_cursor_updates", totalCursorBump),
			slog.Bool("dry_run", opts.dryRun))
	}
	return nil
}

func loadChannelsForBackfill(ctx context.Context, pg *sqlx.DB, channelID, afterID int64, limit int) ([]channelPositionRow, error) {
	rows := make([]channelPositionRow, 0, limit)
	if err := pg.SelectContext(ctx, &rows, selectChannelsForPositionBackfill, channelID, afterID, limit); err != nil {
		if strings.Contains(err.Error(), "message_position") {
			return nil, fmt.Errorf("channels.message_position column is missing; apply the postgres migration first: %w", err)
		}
		return nil, fmt.Errorf("unable to load channels for position backfill: %w", err)
	}
	return rows, nil
}

func loadChannelMessagesForBackfill(ctx context.Context, session *gocql.Session, channelID, lastMessageID int64) ([]messagePositionRow, error) {
	startBucket := idgen.GetBucket(channelID)
	endBucket := idgen.GetBucket(lastMessageID)
	if endBucket < startBucket {
		return nil, nil
	}

	result := make([]messagePositionRow, 0)
	for bucket := startBucket; bucket <= endBucket; bucket++ {
		iter := session.Query(selectBucketMessagePositions, channelID, bucket).WithContext(ctx).Iter()
		bucketRows := make([]messagePositionRow, 0)
		var (
			id       int64
			position int64
		)
		for iter.Scan(&id, &position) {
			bucketRows = append(bucketRows, messagePositionRow{
				ID:       id,
				Bucket:   bucket,
				Position: position,
			})
		}
		if err := iter.Close(); err != nil {
			if strings.Contains(err.Error(), "Undefined column name position") || strings.Contains(err.Error(), "Unknown identifier position") {
				return nil, fmt.Errorf("messages.position column is missing; apply the cassandra migration first: %w", err)
			}
			return nil, fmt.Errorf("unable to load channel messages from bucket %d: %w", bucket, err)
		}

		slices.Reverse(bucketRows)
		result = append(result, bucketRows...)
	}

	return result, nil
}

func planMessagePositionBackfill(messages []messagePositionRow, currentCursor int64, rewrite bool) backfillPlan {
	plan := backfillPlan{
		Updates: make([]messagePositionUpdate, 0),
		Cursor:  currentCursor,
	}

	var next int64
	if rewrite {
		next = 0
		for _, message := range messages {
			next++
			if message.Position != next {
				plan.Updates = append(plan.Updates, messagePositionUpdate{
					ID:       message.ID,
					Bucket:   message.Bucket,
					Position: next,
				})
			}
		}
		if next > plan.Cursor {
			plan.Cursor = next
		}
		return plan
	}

	next = currentCursor
	for _, message := range messages {
		if message.Position > next {
			next = message.Position
		}
		if message.Position != 0 {
			continue
		}
		next++
		plan.Updates = append(plan.Updates, messagePositionUpdate{
			ID:       message.ID,
			Bucket:   message.Bucket,
			Position: next,
		})
	}

	if next > plan.Cursor {
		plan.Cursor = next
	}
	return plan
}

func applyMessagePositionBackfill(ctx context.Context, session *gocql.Session, pg *sqlx.DB, channelID int64, plan backfillPlan, writeBatchSize int) error {
	for start := 0; start < len(plan.Updates); start += writeBatchSize {
		end := start + writeBatchSize
		if end > len(plan.Updates) {
			end = len(plan.Updates)
		}
		batch := session.NewBatch(gocql.UnloggedBatch)
		batch = batch.WithContext(ctx)
		for _, update := range plan.Updates[start:end] {
			batch.Query(updateMessagePosition, update.Position, channelID, update.Bucket, update.ID)
		}
		if err := session.ExecuteBatch(batch); err != nil {
			return fmt.Errorf("unable to apply message position batch: %w", err)
		}
	}

	if _, err := pg.ExecContext(ctx, updateChannelMessagePosition, channelID, plan.Cursor); err != nil {
		return fmt.Errorf("unable to update channel message_position: %w", err)
	}
	return nil
}
