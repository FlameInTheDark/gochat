package msgsearch

import (
	"encoding/json"
	"testing"
)

func TestBuildOSQueryOmitsGuildFilterWhenGuildIDIsNil(t *testing.T) {
	query, err := buildOSQuery(SearchRequest{ChannelId: 42, From: 10})
	if err != nil {
		t.Fatalf("buildOSQuery returned error: %v", err)
	}

	var req osSearchRequest
	if err := json.Unmarshal(query, &req); err != nil {
		t.Fatalf("unable to decode query: %v", err)
	}

	if !hasTermFilter(req.Query.Bool.Filter, "channel_id", 42) {
		t.Fatalf("expected channel filter in query: %+v", req.Query.Bool.Filter)
	}
	if hasAnyTermFilter(req.Query.Bool.Filter, "guild_id") {
		t.Fatalf("did not expect guild filter for DM/private search: %+v", req.Query.Bool.Filter)
	}
}

func TestBuildOSQueryIncludesGuildFilterWhenProvided(t *testing.T) {
	guildID := int64(77)
	query, err := buildOSQuery(SearchRequest{GuildId: &guildID, ChannelId: 42})
	if err != nil {
		t.Fatalf("buildOSQuery returned error: %v", err)
	}

	var req osSearchRequest
	if err := json.Unmarshal(query, &req); err != nil {
		t.Fatalf("unable to decode query: %v", err)
	}

	if !hasTermFilter(req.Query.Bool.Filter, "channel_id", 42) {
		t.Fatalf("expected channel filter in query: %+v", req.Query.Bool.Filter)
	}
	if !hasTermFilter(req.Query.Bool.Filter, "guild_id", guildID) {
		t.Fatalf("expected guild filter in query: %+v", req.Query.Bool.Filter)
	}
}

func hasAnyTermFilter(filters []osSearchQuery, field string) bool {
	for _, filter := range filters {
		if _, ok := filter.Term[field]; ok {
			return true
		}
	}
	return false
}

func hasTermFilter(filters []osSearchQuery, field string, want int64) bool {
	for _, filter := range filters {
		value, ok := filter.Term[field]
		if !ok {
			continue
		}
		switch v := value.(type) {
		case float64:
			return int64(v) == want
		case int64:
			return v == want
		case int:
			return int64(v) == want
		}
	}
	return false
}
