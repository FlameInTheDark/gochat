package msgsearch

type Message struct {
	GuildId   *int64   `json:"guild_id"`
	ChannelId int64    `json:"channel_id"`
	UserId    int64    `json:"user_id"`
	MessageId int64    `json:"message_id"`
	Has       []string `json:"has"`
	Mentions  []int64  `json:"mentions"`
	Content   string   `json:"content"`
}

type DeleteMessage struct {
	ChannelId int64 `json:"channel_id"`
	MessageId int64 `json:"message_id"`
}

type UpdateMessage struct {
	ChannelId int64  `json:"channel_id"`
	MessageId int64  `json:"message_id"`
	Content   string `json:"content"`
}

type SearchMessageResponse struct {
	GuildId   []int64  `json:"guild_id"`
	ChannelId []int64  `json:"channel_id"`
	AuthorId  []int64  `json:"author_id"`
	MessageId []int64  `json:"message_id"`
	Has       []string `json:"has"`
	Mentions  []int64  `json:"mentions"`
	Content   []string `json:"content"`
}

type SearchRequest struct {
	GuildId   int64    `json:"guild_id"`
	ChannelId int64    `json:"channel_id"`
	UserId    *int64   `json:"user_id"`
	Content   *string  `json:"content"`
	Mentions  []int64  `json:"mentions"`
	Has       []string `json:"has"`
	From      int      `json:"from"`
}

type osSearchResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			Source struct {
				MessageId int64 `json:"message_id"`
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

type SortOrder struct {
	Order   string `json:"order,omitempty"`
	Missing string `json:"missing,omitempty"`
}

type osSearchQuery struct {
	Match             map[string]any `json:"match,omitempty"`
	Term              map[string]any `json:"term,omitempty"`
	MatchPhrasePrefix map[string]any `json:"match_phrase_prefix,omitempty"`
}

type osSettingsIndex struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

type osSettings struct {
	Index osSettingsIndex `json:"index"`
}

type osRouting struct {
	Required bool `json:"required"`
}

type osMessagesMapping struct {
	Routing    osRouting            `json:"_routing"`
	Properties osMessagesProperties `json:"properties"`
}

type osProperty struct {
	Type string `json:"type"`
}
type osMessagesProperties struct {
	MessageId osProperty `json:"message_id"`
	UserId    osProperty `json:"user_id"`
	ChannelId osProperty `json:"channel_id"`
	GuildId   osProperty `json:"guild_id"`
	Mentions  osProperty `json:"mentions"`
	Has       osProperty `json:"has"`
	Content   osProperty `json:"content"`
}

type osCreateMessagesIndexRequest struct {
	Settings osSettings        `json:"settings"`
	Mappings osMessagesMapping `json:"mappings"`
}

var defaultMessagesIndex = osCreateMessagesIndexRequest{
	Settings: osSettings{
		Index: osSettingsIndex{
			NumberOfShards:   5,
			NumberOfReplicas: 1,
		},
	},
	Mappings: osMessagesMapping{
		Routing: osRouting{
			Required: true,
		},
		Properties: osMessagesProperties{
			MessageId: osProperty{
				Type: "long",
			},
			UserId: osProperty{
				Type: "long",
			},
			ChannelId: osProperty{
				Type: "long",
			},
			GuildId: osProperty{
				Type: "long",
			},
			Mentions: osProperty{
				Type: "long",
			},
			Has: osProperty{
				Type: "keyword",
			},
			Content: osProperty{
				Type: "text",
			},
		},
	},
}
