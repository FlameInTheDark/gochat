package message

import "github.com/FlameInTheDark/gochat/internal/embed"

type SendRequest struct {
	Content     string        `json:"content"`
	Attachments []int64       `json:"attachments"`
	Embeds      []embed.Embed `json:"embeds"`
	Reference   *int64        `json:"reference"`
}

type UpdateRequest struct {
	Content *string        `json:"content"`
	Embeds  *[]embed.Embed `json:"embeds"`
	Flags   *int           `json:"flags"`
}

type GetReactionUsersRequest struct {
	After *int64 `query:"after" json:"after" example:"2230469276416868352"`
	Limit *int   `query:"limit" json:"limit" example:"50"`
}
