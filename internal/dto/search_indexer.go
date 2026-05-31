package dto

type GuildIndexMessage struct {
	GuildId int64 `json:"guild_id"`
}

type GuildIndexDeleteMessage struct {
	GuildId int64 `json:"guild_id"`
}

type BotIndexMessage struct {
	BotUserId int64 `json:"bot_user_id"`
}

type BotIndexDeleteMessage struct {
	BotUserId int64 `json:"bot_user_id"`
}
