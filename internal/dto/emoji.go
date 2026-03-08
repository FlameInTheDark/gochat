package dto

type EmojiRef struct {
	Name string `json:"name"`
	Id   int64  `json:"id,string"`
}

type GuildEmoji struct {
	Id       int64  `json:"id,string"`
	GuildId  int64  `json:"guild_id,string"`
	Name     string `json:"name"`
	Animated bool   `json:"animated"`
}

type EmojiUpload struct {
	Id      int64  `json:"id,string"`
	GuildId int64  `json:"guild_id,string"`
	Name    string `json:"name"`
}
