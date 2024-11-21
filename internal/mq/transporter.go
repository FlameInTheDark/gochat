package mq

import "github.com/FlameInTheDark/gochat/internal/mq/mqmsg"

type SendTransporter interface {
	SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error
	SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error
	SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error
}
