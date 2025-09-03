package search

const (
	ErrPermissionsRequired = "permissions required"
)

type SearchRequest struct {
	GuildId   int64    `json:"guild_id"`
	ChannelId *int64   `json:"channel_id"`
	Mentions  []int64  `json:"mentions"`
	AuthorId  *int64   `json:"author_id"`
	Content   *string  `json:"content"`
	Has       []string `json:"has"`
}
