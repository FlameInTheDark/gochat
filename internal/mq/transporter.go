package mq

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type SendTransporter interface {
	SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error
	SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error
	SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error
}

type ContextSendTransporter interface {
	SendChannelMessageContext(ctx context.Context, channelId int64, message mqmsg.EventDataMessage) error
	SendGuildUpdateContext(ctx context.Context, guildId int64, message mqmsg.EventDataMessage) error
	SendUserUpdateContext(ctx context.Context, userId int64, message mqmsg.EventDataMessage) error
}

func SendChannelMessage(ctx context.Context, transporter SendTransporter, channelId int64, message mqmsg.EventDataMessage) error {
	if transporter == nil {
		return nil
	}
	if contextual, ok := transporter.(ContextSendTransporter); ok {
		return contextual.SendChannelMessageContext(ctx, channelId, message)
	}
	return transporter.SendChannelMessage(channelId, message)
}

func SendGuildUpdate(ctx context.Context, transporter SendTransporter, guildId int64, message mqmsg.EventDataMessage) error {
	if transporter == nil {
		return nil
	}
	if contextual, ok := transporter.(ContextSendTransporter); ok {
		return contextual.SendGuildUpdateContext(ctx, guildId, message)
	}
	return transporter.SendGuildUpdate(guildId, message)
}

func SendUserUpdate(ctx context.Context, transporter SendTransporter, userId int64, message mqmsg.EventDataMessage) error {
	if transporter == nil {
		return nil
	}
	if contextual, ok := transporter.(ContextSendTransporter); ok {
		return contextual.SendUserUpdateContext(ctx, userId, message)
	}
	return transporter.SendUserUpdate(userId, message)
}
