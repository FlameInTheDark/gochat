package mq

type MQChannelType int64

const (
	MQMessageChannel MQChannelType = iota
	MQPresenceChannel
	MQGuildChannel
	MQPersonalChannel
)
