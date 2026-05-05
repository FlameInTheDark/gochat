package guildsearch

import (
	"encoding/json"
	"testing"
)

func TestBuildOSQueryAlwaysFiltersPublicAndCapsLimit(t *testing.T) {
	data, err := buildOSQuery(SearchRequest{
		Query: "chat",
		Tags:  []string{"go", "voice"},
		Sort:  SortPopularity,
		From:  32,
		Size:  50,
	})
	if err != nil {
		t.Fatalf("buildOSQuery returned error: %v", err)
	}

	var req map[string]any
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("decode query: %v", err)
	}
	if got := int(req["size"].(float64)); got != 16 {
		t.Fatalf("size = %d, want capped 16", got)
	}
	assertBoolFilter(t, req, "public", true)
	assertBoolFilter(t, req, "tags", "go")
	assertBoolFilter(t, req, "tags", "voice")
}

func TestBuildOSTagsQueryAlwaysFiltersPublic(t *testing.T) {
	data, err := buildOSTagsQuery("go", 99)
	if err != nil {
		t.Fatalf("buildOSTagsQuery returned error: %v", err)
	}

	var req map[string]any
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("decode query: %v", err)
	}
	assertBoolFilter(t, req, "public", true)
	assertPrefixFilter(t, req, "tags", "go")

	aggs := req["aggs"].(map[string]any)
	tags := aggs["tags"].(map[string]any)
	terms := tags["terms"].(map[string]any)
	if got := int(terms["size"].(float64)); got != 16 {
		t.Fatalf("tag agg size = %d, want capped 16", got)
	}
}

func assertBoolFilter(t *testing.T, req map[string]any, field string, want any) {
	t.Helper()
	for _, filter := range filters(t, req) {
		term, ok := filter["term"].(map[string]any)
		if !ok {
			continue
		}
		if got, ok := term[field]; ok && got == want {
			return
		}
	}
	t.Fatalf("missing term filter %s=%v in %#v", field, want, req)
}

func assertPrefixFilter(t *testing.T, req map[string]any, field, want string) {
	t.Helper()
	for _, filter := range filters(t, req) {
		prefix, ok := filter["prefix"].(map[string]any)
		if !ok {
			continue
		}
		if got, ok := prefix[field]; ok && got == want {
			return
		}
	}
	t.Fatalf("missing prefix filter %s=%s in %#v", field, want, req)
}

func filters(t *testing.T, req map[string]any) []map[string]any {
	t.Helper()
	query := req["query"].(map[string]any)
	boolQuery := query["bool"].(map[string]any)
	rawFilters := boolQuery["filter"].([]any)
	out := make([]map[string]any, 0, len(rawFilters))
	for _, raw := range rawFilters {
		out = append(out, raw.(map[string]any))
	}
	return out
}
