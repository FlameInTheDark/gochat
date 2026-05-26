package mq

import (
	"context"
	"errors"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type FanoutTransporter struct {
	transports []SendTransporter
}

func NewFanoutTransporter(transports ...SendTransporter) *FanoutTransporter {
	filtered := make([]SendTransporter, 0, len(transports))
	for _, transport := range transports {
		if transport != nil {
			filtered = append(filtered, transport)
		}
	}
	return &FanoutTransporter{transports: filtered}
}

func (t *FanoutTransporter) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return t.SendChannelMessageContext(context.Background(), channelId, message)
}

func (t *FanoutTransporter) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	return t.SendGuildUpdateContext(context.Background(), guildId, message)
}

func (t *FanoutTransporter) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return t.SendUserUpdateContext(context.Background(), userId, message)
}

func (t *FanoutTransporter) SendChannelMessageContext(ctx context.Context, channelId int64, message mqmsg.EventDataMessage) error {
	var errs []error
	for _, transport := range t.transports {
		if err := SendChannelMessage(ctx, transport, channelId, message); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (t *FanoutTransporter) SendGuildUpdateContext(ctx context.Context, guildId int64, message mqmsg.EventDataMessage) error {
	var errs []error
	for _, transport := range t.transports {
		if err := SendGuildUpdate(ctx, transport, guildId, message); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (t *FanoutTransporter) SendUserUpdateContext(ctx context.Context, userId int64, message mqmsg.EventDataMessage) error {
	var errs []error
	for _, transport := range t.transports {
		if err := SendUserUpdate(ctx, transport, userId, message); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
