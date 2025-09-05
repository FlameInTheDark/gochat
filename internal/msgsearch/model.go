package msgsearch

type AddMessage struct {
	GuildId   int64    `json:"guild_id"`
	ChannelId int64    `json:"channel_id"`
	AuthorId  int64    `json:"author_id"`
	MessageId int64    `json:"message_id"`
	Has       []string `json:"has"`
	Mentions  []int64  `json:"mentions"`
	Content   string   `json:"content"`
}

type SearchMessageResponse struct {
	GuildId   []int64  `json:"guild_id"`
	ChannelId []int64  `json:"channel_id"`
	AuthorId  []int64  `json:"author_id"`
	MessageId []int64  `json:"message_id"`
	Has       []string `json:"has"`
	Mentions  []int64  `json:"mentions"`
	Content   []string `json:"content"`
}

type SearchRequest struct {
	GuildId   int64    `json:"guild_id"`
	ChannelId *int64   `json:"channel_id"`
	AuthorId  *int64   `json:"author_id"`
	Content   *string  `json:"content"`
	Mentions  []int64  `json:"mentions"`
	Has       []string `json:"has"`
	From      int      `json:"from"`
}
