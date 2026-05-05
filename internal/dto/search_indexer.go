package dto

type GuildIndexMessage struct {
	GuildId int64 `json:"guild_id"`
}

type GuildIndexDeleteMessage struct {
	GuildId int64 `json:"guild_id"`
}
