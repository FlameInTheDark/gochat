package dto

type Message struct {
	Id        int64  `json:"id"`
	ChannelId int64  `json:"channel_id"`
	Author    User   `json:"author_id"`
	Content   string `json:"content"`
}
