package dto

// IconUpload describes newly created guild icon placeholder
type IconUpload struct {
	Id      int64 `json:"id" example:"2230469276416868352"`
	GuildId int64 `json:"guild_id" example:"2230469276416868352"`
}

// Icon is a full guild icon description returned in guild payloads
type Icon struct {
	Id       int64  `json:"id" example:"2230469276416868352"`
	URL      string `json:"url" example:"https://cdn.example.com/icons/2230/2231.webp"`
	Filesize int64  `json:"filesize" example:"12345"`
	Width    int64  `json:"width" example:"128"`
	Height   int64  `json:"height" example:"128"`
}
