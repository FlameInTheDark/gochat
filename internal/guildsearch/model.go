package guildsearch

type Guild struct {
	GuildID      int64    `json:"guild_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Public       bool     `json:"public"`
	MembersCount int64    `json:"members_count"`
}

type SearchRequest struct {
	Query string
	Tags  []string
	Sort  SortMode
	From  int
	Size  int
}

type SortMode string

const (
	SortBestMatch    SortMode = "best_match"
	SortPopularity   SortMode = "popularity"
	SortAlphabetical SortMode = "alphabetical"
)

type Results struct {
	Ids   []int64
	Total int
}

type osSearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source struct {
				GuildID int64 `json:"guild_id"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type osSearchRequest struct {
	From   int      `json:"from,omitempty"`
	Size   int      `json:"size,omitempty"`
	Source []string `json:"_source,omitempty"`
	Query  struct {
		Bool struct {
			Filter []osSearchQuery `json:"filter,omitempty"`
			Must   []osSearchQuery `json:"must,omitempty"`
		} `json:"bool,omitempty"`
	} `json:"query,omitempty"`
	Sort []map[string]any `json:"sort,omitempty"`
}

type osSearchQuery struct {
	Term       map[string]any `json:"term,omitempty"`
	MultiMatch map[string]any `json:"multi_match,omitempty"`
	Prefix     map[string]any `json:"prefix,omitempty"`
}

type SortOrder struct {
	Order   string `json:"order,omitempty"`
	Missing string `json:"missing,omitempty"`
}

type osSettingsIndex struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

type osSettings struct {
	Index osSettingsIndex `json:"index"`
}

type osProperty struct {
	Type   string                `json:"type"`
	Fields map[string]osProperty `json:"fields,omitempty"`
}

type osGuildsMapping struct {
	Properties osGuildsProperties `json:"properties"`
}

type osGuildsProperties struct {
	GuildID      osProperty `json:"guild_id"`
	Name         osProperty `json:"name"`
	Description  osProperty `json:"description"`
	Tags         osProperty `json:"tags"`
	Public       osProperty `json:"public"`
	MembersCount osProperty `json:"members_count"`
}

type osCreateGuildsIndexRequest struct {
	Settings osSettings      `json:"settings"`
	Mappings osGuildsMapping `json:"mappings"`
}

var defaultGuildsIndex = osCreateGuildsIndexRequest{
	Settings: osSettings{
		Index: osSettingsIndex{
			NumberOfShards:   3,
			NumberOfReplicas: 1,
		},
	},
	Mappings: osGuildsMapping{
		Properties: osGuildsProperties{
			GuildID: osProperty{Type: "long"},
			Name: osProperty{
				Type: "text",
				Fields: map[string]osProperty{
					"keyword": {Type: "keyword"},
				},
			},
			Description:  osProperty{Type: "text"},
			Tags:         osProperty{Type: "keyword"},
			Public:       osProperty{Type: "boolean"},
			MembersCount: osProperty{Type: "long"},
		},
	},
}
