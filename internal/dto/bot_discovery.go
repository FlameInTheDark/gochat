package dto

type BotDiscovery struct {
	BotUserId          int64    `json:"bot_user_id" example:"2230469276416868352"`
	User               User     `json:"user"`
	Description        string   `json:"description" example:"Moderation and utility bot"`
	Tags               []string `json:"tags"`
	DefaultPermissions int64    `json:"default_permissions" example:"7927905"`
	InstallsCount      int64    `json:"installs_count" example:"42"`
	CreatedAt          int64    `json:"created_at" example:"1714500000000"`
	UpdatedAt          int64    `json:"updated_at" example:"1714500000000"`
}

type BotDiscoverySearchResponse struct {
	Bots  []BotDiscovery `json:"bots"`
	Pages int            `json:"pages"`
}
