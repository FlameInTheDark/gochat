package dto

type Role struct {
	Id          int64  `json:"id" example:"2230469276416868352"`       // Role ID
	GuildId     int64  `json:"guild_id" example:"2230469276416868352"` // Guild ID
	Name        string `json:"name" example:"role-name"`               // Role name
	Color       int    `json:"color"`                                  // Role color. Will change username color. Represent RGB color in one Integer value.
	Permissions int64  `json:"permissions"`                            // Role permissions. Check the permissions documentation for more info.
}
