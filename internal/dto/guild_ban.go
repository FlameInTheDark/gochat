package dto

type GuildBan struct {
	User   User    `json:"user"`
	Reason *string `json:"reason,omitempty"`
}
