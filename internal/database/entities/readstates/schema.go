package readstates

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
)

const (
	getReadStates    = `SELECT channels FROM gochat.read_states WHERE user_id = ?`
	setReadState     = `UPDATE gochat.read_states SET channels[?] = ? WHERE user_id = ?`
	setReadStateMany = `UPDATE gochat.read_states SET channels = channels + ? WHERE user_id = ?`
)

func (e *Entity) GetReadStates(ctx context.Context, userId int64) (map[int64]int64, error) {
	var readStates map[int64]int64
	err := e.c.Session().
		Query(getReadStates).
		WithContext(ctx).
		Bind(userId).
		Scan(&readStates)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return readStates, nil
}

func (e *Entity) GetReadState(ctx context.Context, userId, channelId int64) (int64, error) {
	readStates, err := e.GetReadStates(ctx, userId)
	if err != nil {
		return 0, err
	}
	if len(readStates) == 0 {
		return 0, nil
	}
	return readStates[channelId], nil
}

func (e *Entity) SetReadState(ctx context.Context, userId, channelId, lastMessageId int64) error {
	current, err := e.GetReadState(ctx, userId, channelId)
	if err != nil {
		return err
	}
	if lastMessageId <= current {
		return nil
	}

	err = e.c.Session().
		Query(setReadState).
		WithContext(ctx).
		Bind(channelId, lastMessageId, userId).
		Exec()
	return err
}

func (e *Entity) SetReadStateMany(ctx context.Context, userId, values map[int64]int64) error {
	err := e.c.Session().
		Query(setReadStateMany).
		WithContext(ctx).
		Bind(values, userId).
		Exec()
	return err
}
