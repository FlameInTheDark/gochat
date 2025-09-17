package dto

// IndexMessage is a structure with message data that the search engine will index and search by
type IndexMessage struct {
	MessageId int64  `json:"message_id"`
	UserId    int64  `json:"user_id"`
	ChannelId int64  `json:"channel_id"`
	GuildId   *int64 `json:"guild_id"`
	// Mentions contains users IDs
	Mentions []int64 `json:"mentions"`
	// Has is a list of features that the message contains (url, image, video, file)
	Has     []string `json:"has"`
	Content string   `json:"content"`
}

type IndexDeleteMessage struct {
	ChannelId int64 `json:"channel_id"`
	MessageId int64 `json:"message_id"`
}
