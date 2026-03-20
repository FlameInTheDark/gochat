package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanDashboardUpsertsIsIdempotent(t *testing.T) {
	local := []openObserveDashboard{
		{
			DashboardID: "7416801000000000001",
			Title:       "Edge/API",
			FolderID:    "default",
			Payload: map[string]any{
				"dashboardId": "7416801000000000001",
				"title":       "Edge/API",
			},
		},
		{
			DashboardID: "7416801000000000002",
			Title:       "Realtime/WS",
			FolderID:    "default",
			Payload: map[string]any{
				"dashboardId": "7416801000000000002",
				"title":       "Realtime/WS",
			},
		},
	}
	existing := map[string]openObserveDashboardSummary{
		"id:7416801000000000001": {
			DashboardID: "7416801000000000001",
			Title:       "Edge/API",
			FolderID:    "default",
			Hash:        "hash-1",
		},
		"title:realtime/ws": {
			DashboardID: "9000000000000000002",
			Title:       "Realtime/WS",
			FolderID:    "default",
			Hash:        "hash-2",
		},
	}

	ops := planDashboardUpserts(local, existing)
	if len(ops) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(ops))
	}
	for _, op := range ops {
		if op.Method != "PUT" {
			t.Fatalf("expected update operation, got %#v", op)
		}
	}
	if ops[1].DashboardID != "9000000000000000002" {
		t.Fatalf("expected title match to reuse existing dashboard id, got %#v", ops[1])
	}
	if got := ops[1].Payload["dashboardId"]; got != "9000000000000000002" {
		t.Fatalf("expected payload dashboardId rewrite, got %#v", got)
	}
}

func TestBuildDashboardSyncPlansReportsInSyncUpdateAndCreate(t *testing.T) {
	samePayload := map[string]any{
		"dashboardId": "7416801000000000001",
		"title":       "Edge/API",
		"tabs":        []any{},
	}
	sameHash, err := dashboardPayloadHash(samePayload)
	if err != nil {
		t.Fatalf("dashboardPayloadHash failed: %v", err)
	}

	changedPayload := map[string]any{
		"dashboardId": "7416801000000000002",
		"title":       "Realtime/WS",
		"description": "new repo payload",
		"tabs":        []any{},
	}

	local := []openObserveDashboard{
		{DashboardID: "7416801000000000001", Title: "Edge/API", FolderID: "default", Payload: samePayload},
		{DashboardID: "7416801000000000002", Title: "Realtime/WS", FolderID: "default", Payload: changedPayload},
		{DashboardID: "7416801000000000003", Title: "Async Workers", FolderID: "default", Payload: map[string]any{"dashboardId": "7416801000000000003", "title": "Async Workers", "tabs": []any{}}},
	}
	existing := map[string]openObserveDashboardSummary{
		"id:7416801000000000001": {
			DashboardID: "7416801000000000001",
			Title:       "Edge/API",
			FolderID:    "default",
			Hash:        "hash-1",
			PayloadHash: sameHash,
		},
		"title:realtime/ws": {
			DashboardID: "9000000000000000002",
			Title:       "Realtime/WS",
			FolderID:    "default",
			Hash:        "hash-2",
			PayloadHash: "different",
		},
	}

	plans := buildDashboardSyncPlans(local, existing)
	if len(plans) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(plans))
	}
	if plans[0].Action != dashboardActionInSync {
		t.Fatalf("expected first dashboard to be in-sync, got %#v", plans[0])
	}
	if plans[1].Action != dashboardActionUpdate || plans[1].Operation.DashboardID != "9000000000000000002" {
		t.Fatalf("expected second dashboard update plan, got %#v", plans[1])
	}
	if plans[2].Action != dashboardActionCreate {
		t.Fatalf("expected third dashboard create plan, got %#v", plans[2])
	}
}

func TestLoadAlertSeedsParsesSeedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alerts.seed.yaml")
	content := []byte("alerts:\n  - name: Example Alert\n    signal: metrics\n    metric: gochat.http.server.errors\n    window: 5m\n    threshold: \"> 10\"\n    severity: high\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	seeds, err := loadAlertSeeds(path)
	if err != nil {
		t.Fatalf("loadAlertSeeds returned error: %v", err)
	}
	if len(seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(seeds))
	}
	if seeds[0].Name != "Example Alert" || seeds[0].Signal != "metrics" {
		t.Fatalf("unexpected seed: %#v", seeds[0])
	}
}

func TestLoadDashboardDefinitionsNormalizesMetricStreams(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "edge.dashboard.json")
	payload := map[string]any{
		"dashboardId": "7416801000000000999",
		"title":       "Edge/API",
		"tabs": []any{
			map[string]any{
				"tabId": "overview",
				"name":  "Overview",
				"panels": []any{
					map[string]any{
						"id":        "Panel_API_001",
						"queryType": "promql",
						"queries": []any{
							map[string]any{
								"query": "sum by (service) (rate(gochat_http_server_requests_total[5m]))",
								"fields": map[string]any{
									"stream":      "gochat-metrics",
									"stream_type": "metrics",
								},
							},
							map[string]any{
								"query": "max(gochat_postgres_probe_status)",
								"fields": map[string]any{
									"stream":      "gochat-metrics",
									"stream_type": "metrics",
								},
							},
						},
					},
				},
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	dashboards, err := loadDashboardDefinitions(dir)
	if err != nil {
		t.Fatalf("loadDashboardDefinitions returned error: %v", err)
	}
	if len(dashboards) != 1 {
		t.Fatalf("expected 1 dashboard, got %d", len(dashboards))
	}

	tabs, ok := dashboards[0].Payload["tabs"].([]any)
	if !ok || len(tabs) != 1 {
		t.Fatalf("expected one tab in normalized payload, got %#v", dashboards[0].Payload["tabs"])
	}
	tab, ok := tabs[0].(map[string]any)
	if !ok {
		t.Fatalf("expected tab map, got %#v", tabs[0])
	}
	panels, ok := tab["panels"].([]any)
	if !ok || len(panels) != 1 {
		t.Fatalf("expected one panel, got %#v", tab["panels"])
	}
	panel, ok := panels[0].(map[string]any)
	if !ok {
		t.Fatalf("expected panel map, got %#v", panels[0])
	}
	queries, ok := panel["queries"].([]any)
	if !ok || len(queries) != 2 {
		t.Fatalf("expected two queries, got %#v", panel["queries"])
	}

	firstQuery, ok := queries[0].(map[string]any)
	if !ok {
		t.Fatalf("expected query map, got %#v", queries[0])
	}
	firstFields, ok := firstQuery["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields map, got %#v", firstQuery["fields"])
	}
	if got := firstFields["stream"]; got != "gochat_http_server_requests" {
		t.Fatalf("expected counter metric stream normalization, got %#v", got)
	}

	secondQuery, ok := queries[1].(map[string]any)
	if !ok {
		t.Fatalf("expected query map, got %#v", queries[1])
	}
	secondFields, ok := secondQuery["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields map, got %#v", secondQuery["fields"])
	}
	if got := secondFields["stream"]; got != "gochat_postgres_probe_status" {
		t.Fatalf("expected gauge metric stream normalization, got %#v", got)
	}
}

func TestBuildAlertRequestForPostgresProbeFailing(t *testing.T) {
	req, err := buildAlertRequest(alertSeed{
		Name:      "Postgres probe failing",
		Signal:    "metrics",
		Metric:    "gochat.postgres.probe.status",
		Window:    "5m",
		Threshold: "< 1",
		Severity:  "high",
	}, "pagerduty-primary")
	if err != nil {
		t.Fatalf("buildAlertRequest returned error: %v", err)
	}

	if req.StreamName != "gochat_postgres_probe_status" {
		t.Fatalf("unexpected stream name: %q", req.StreamName)
	}
	if req.QueryCondition.PromQL != "min by (service_name) (last_over_time(gochat_postgres_probe_status[5m]))" {
		t.Fatalf("unexpected promql query: %q", req.QueryCondition.PromQL)
	}
	if req.QueryCondition.PromQLCondition == nil || req.QueryCondition.PromQLCondition.Operator != "<" {
		t.Fatalf("unexpected promql condition: %#v", req.QueryCondition.PromQLCondition)
	}
}

func TestRepoObservabilityAssetsLoadWithoutLegacyPanels(t *testing.T) {
	dashboards, err := loadDashboardDefinitions(filepath.Join("..", "..", "monitoring", "openobserve", "bootstrap", "dashboards"))
	if err != nil {
		t.Fatalf("loadDashboardDefinitions returned error: %v", err)
	}
	if len(dashboards) == 0 {
		t.Fatal("expected repo dashboards to load")
	}

	alerts, err := loadAlertSeeds(filepath.Join("..", "..", "monitoring", "openobserve", "bootstrap", "alerts", "alerts.seed.yaml"))
	if err != nil {
		t.Fatalf("loadAlertSeeds returned error: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected repo alerts to load")
	}

	for _, dashboard := range dashboards {
		raw, err := json.Marshal(dashboard.Payload)
		if err != nil {
			t.Fatalf("Marshal dashboard %q failed: %v", dashboard.Title, err)
		}
		text := string(raw)
		if containsAny(text, "pg_up", "postgres_exporter", "promhttp_", "scrape_", "citus_", "http_client_") {
			t.Fatalf("dashboard %q still references legacy exporter-era telemetry", dashboard.Title)
		}
	}

	for _, alert := range alerts {
		raw, err := json.Marshal(alert)
		if err != nil {
			t.Fatalf("Marshal alert %q failed: %v", alert.Name, err)
		}
		text := string(raw)
		if containsAny(text, "pg_up", "postgres_exporter", "promhttp_", "scrape_", "citus_", "http_client_") {
			t.Fatalf("alert %q still references legacy exporter-era telemetry", alert.Name)
		}
	}
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func TestFilterLegacyMetricStreamsMatchesOnlyLegacyPrefixes(t *testing.T) {
	streams := []openObserveStream{
		{Name: "pg_up", StreamType: "metrics"},
		{Name: "process_cpu_seconds_total", StreamType: "metrics"},
		{Name: "gochat_http_server_requests", StreamType: "metrics"},
		{Name: "http_client_duration_bucket", StreamType: "metrics"},
		{Name: "up", StreamType: "metrics"},
	}

	filtered := filterLegacyMetricStreams(streams)
	if len(filtered) != 4 {
		t.Fatalf("expected 4 legacy streams, got %d: %#v", len(filtered), filtered)
	}
	names := make([]string, 0, len(filtered))
	for _, stream := range filtered {
		names = append(names, stream.Name)
	}
	for _, allowed := range []string{"pg_up", "process_cpu_seconds_total", "http_client_duration_bucket", "up"} {
		found := false
		for _, name := range names {
			if name == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected %q in filtered streams, got %#v", allowed, names)
		}
	}
}
