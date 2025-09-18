package model

type GroupDMChannel struct {
	ChannelId int64 `db:"channel_id"`
	UserId    int64 `db:"user_id"`
}
