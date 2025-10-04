package channel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	var ch model.Channel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channels").
		Where(squirrel.Eq{"id": id}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return ch, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &ch, raw, args...)
	if err != nil {
		return model.Channel{}, fmt.Errorf("unable to get channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) GetChannelsBulk(ctx context.Context, ids []int64) ([]model.Channel, error) {
	var chs []model.Channel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channels").
		Where(squirrel.Eq{"id": ids}).
		OrderBy("id asc")
	raw, args, err := q.ToSql()
	if err != nil {
		return chs, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &chs, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return chs, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get channels: %w", err)
	}
	return chs, nil
}

func (e *Entity) GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error) {
	var channels []model.Channel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channels").
		Where(
			squirrel.And{
				squirrel.Eq{"parent_id": channelId},
				squirrel.Eq{"type": model.ChannelTypeThread},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return channels, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &channels, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return channels, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get channel threads: %w", err)
	}
	return channels, nil
}

func (e *Entity) CreateChannel(ctx context.Context, id int64, name string, channelType model.ChannelType, parent *int64, permissions *int64, private bool) error {
	q := squirrel.Insert("channels").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "name", "type", "parent_id", "permissions", "private", "last_message").
		Values(id, name, channelType, parent, permissions, private, 0)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to create channel: %w", err)
	}
	return nil
}

func (e *Entity) DeleteChannel(ctx context.Context, id int64) error {
	q := squirrel.Delete("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to delete channel: %w", err)
	}
	return nil
}

func (e *Entity) RenameChannel(ctx context.Context, id int64, newName string) error {
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("name", newName)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to rename channel: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelPermissions(ctx context.Context, id int64, permissions int) error {
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("permissions", permissions)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set channel permissions: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelPrivate(ctx context.Context, id int64, private bool) error {
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("private", private)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set channel private: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelTopic(ctx context.Context, id int64, topic *string) error {
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("topic", topic)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set channel topic: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelParent(ctx context.Context, id int64, parent *int64) error {
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("parent_id", parent)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set channel parent: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelParentBulk(ctx context.Context, id []int64, parent *int64) error {
	tx, err := e.c.Beginx()
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}

	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("parent_id", parent)
	raw, args, err := q.ToSql()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, raw, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to set channel parent bulk: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to commit transaction: %w", err)
	}
	return nil
}

func (e *Entity) SetLastMessage(ctx context.Context, id, lastMessage int64) error {
	tx, err := e.c.Beginx()
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("last_message", lastMessage)
	raw, args, err := q.ToSql()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, raw, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to set channel last message: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to commit transaction: %w", err)
	}
	return nil
}

func (e *Entity) UpdateChannel(ctx context.Context, id int64, parent *int64, private *bool, name, topic *string) (model.Channel, error) {
	q := squirrel.Update("channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING *")
	if parent != nil {
		if *parent == 0 {
			q = q.Set("parent_id", nil)
		} else {
			q = q.Set("parent_id", *parent)
		}
	}
	if private != nil {
		q = q.Set("private", *private)
	}
	if name != nil {
		q = q.Set("name", *name)
	}
	if topic != nil {
		q = q.Set("topic", *topic)
	}
	raw, args, err := q.ToSql()
	if err != nil {
		return model.Channel{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	var ch model.Channel
	err = e.c.GetContext(ctx, &ch, raw, args...)
	if err != nil {
		return model.Channel{}, fmt.Errorf("unable to update channel: %w", err)
	}
	return ch, nil
}
