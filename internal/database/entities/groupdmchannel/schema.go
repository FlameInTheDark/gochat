package groupdmchannel

import (
	"context"
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gocql/gocql"
)

const (
	joinGroupDmChannel     = `INSERT INTO gochat.group_dm_channels (channel_id, user_id) VALUES (?, ?)`
	getGroupDmChannel      = `SELECT channel_id, user_id FROM gochat.group_dm_channels WHERE user_id = ?`
	leaveGroupDmChannel    = `DELETE FROM gochat.group_dm_channels WHERE channel_id = ? AND user_id = ?`
	getGroupDmParticipants = `SELECT user_id FROM gochat.group_dm_participants WHERE channel_id = ?`
	isGroupDmParticipant   = `SELECT count(user_id) FROM gochat.group_dm_participants WHERE channel_id = ?`
)

func (e *Entity) JoinGroupDmChannelMany(ctx context.Context, channelId int64, users []int64) error {
	if len(users) == 0 {
		return nil
	}
	batch := e.c.Session().
		NewBatch(gocql.UnloggedBatch).
		WithContext(ctx)
	for _, userId := range users {
		batch.Query(joinGroupDmChannel, channelId, userId)
	}
	err := e.c.Session().
		ExecuteBatch(batch)
	if err != nil {
		return fmt.Errorf("join group dm channel many: %w", err)
	}
	return nil
}

func (e *Entity) JoinGroupDmChannel(ctx context.Context, channelId, userId int64) error {
	err := e.c.Session().
		Query(joinGroupDmChannel).
		WithContext(ctx).
		Bind(channelId, userId).
		Exec()
	if err != nil {
		return fmt.Errorf("join group dm channel error: %w", err)
	}
	return nil
}

func (e *Entity) GetGroupDmChannel(ctx context.Context, userId int64) (model.GroupDMChannel, error) {
	var ch model.GroupDMChannel
	err := e.c.Session().
		Query(getGroupDmChannel).
		WithContext(ctx).
		Bind(userId).
		Scan(&ch.ChannelId, &ch.UserId)
	if err != nil {
		return model.GroupDMChannel{}, fmt.Errorf("get group dm channel error: %w", err)
	}
	return ch, nil
}

func (e *Entity) LeaveGroupDmChannel(ctx context.Context, channelId, userId int64) error {
	err := e.c.Session().
		Query(leaveGroupDmChannel).
		WithContext(ctx).
		Bind(channelId, userId).
		Exec()
	if err != nil {
		return fmt.Errorf("leave group dm channel error: %w", err)
	}
	return nil
}

func (e *Entity) GetGroupDmParticipants(ctx context.Context, channelId int64) ([]model.GroupDMChannel, error) {
	var channels []model.GroupDMChannel
	iter := e.c.Session().
		Query(getGroupDmParticipants).
		WithContext(ctx).
		Bind(channelId).
		Iter()
	var ch model.GroupDMChannel
	for iter.Scan(&ch.ChannelId, &ch.UserId) {
		channels = append(channels, ch)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("get group dm participants error: %w", err)
	}
	return channels, nil
}

func (e *Entity) IsGroupDmParticipant(ctx context.Context, channelId int64, userId int64) (bool, error) {
	var count int64
	err := e.c.Session().
		Query(isGroupDmParticipant).
		WithContext(ctx).
		Bind(channelId, userId).
		Scan(&count)
	if err == gocql.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("check group dm participant error: %w", err)
	}
	return count > 0, nil
}
