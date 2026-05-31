package botsearch

type Bot struct {
	BotUserID     int64    `json:"bot_user_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Bio           string   `json:"bio"`
	Tags          []string `json:"tags"`
	Public        bool     `json:"public"`
	Disabled      bool     `json:"disabled"`
	InstallsCount int64    `json:"installs_count"`
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
				BotUserID int64 `json:"bot_user_id"`
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

type osBotsMapping struct {
	Properties osBotsProperties `json:"properties"`
}

type osBotsProperties struct {
	BotUserID     osProperty `json:"bot_user_id"`
	Name          osProperty `json:"name"`
	Description   osProperty `json:"description"`
	Bio           osProperty `json:"bio"`
	Tags          osProperty `json:"tags"`
	Public        osProperty `json:"public"`
	Disabled      osProperty `json:"disabled"`
	InstallsCount osProperty `json:"installs_count"`
}

type osCreateBotsIndexRequest struct {
	Settings osSettings    `json:"settings"`
	Mappings osBotsMapping `json:"mappings"`
}

var defaultBotsIndex = osCreateBotsIndexRequest{
	Settings: osSettings{
		Index: osSettingsIndex{
			NumberOfShards:   3,
			NumberOfReplicas: 1,
		},
	},
	Mappings: osBotsMapping{
		Properties: osBotsProperties{
			BotUserID: osProperty{Type: "long"},
			Name: osProperty{
				Type: "text",
				Fields: map[string]osProperty{
					"keyword": {Type: "keyword"},
				},
			},
			Description:   osProperty{Type: "text"},
			Bio:           osProperty{Type: "text"},
			Tags:          osProperty{Type: "keyword"},
			Public:        osProperty{Type: "boolean"},
			Disabled:      osProperty{Type: "boolean"},
			InstallsCount: osProperty{Type: "long"},
		},
	},
}
