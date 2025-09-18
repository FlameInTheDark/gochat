package model

type DMChannel struct {
	UserId        int64 `db:"user_id"`
	ParticipantId int64 `db:"participant_id"`
	ChannelId     int64 `db:"channel_id"`
}
