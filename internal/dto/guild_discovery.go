package dto

type GuildDiscovery struct {
	Id           int64    `json:"id" example:"2230469276416868352"`
	Name         string   `json:"name" example:"My Guild"`
	Description  string   `json:"description" example:"Community for Go developers"`
	MembersCount int64    `json:"members_count" example:"42"`
	Icon         *Icon    `json:"icon,omitempty"`
	Tags         []string `json:"tags"`
}

type GuildDiscoverySearchResponse struct {
	Guilds []GuildDiscovery `json:"guilds"`
	Pages  int              `json:"pages"`
}

type GuildDiscoveryUpdateResponse struct {
	Guild           GuildDiscovery `json:"guild"`
	IndexingPending bool           `json:"indexing_pending"`
}
