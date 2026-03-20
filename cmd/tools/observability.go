package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	cli "github.com/urfave/cli/v3"
)

const (
	defaultOpenObserveURL   = "http://localhost:5080"
	defaultOpenObserveOrg   = "default"
	defaultOpenObserveUser  = "root@example.com"
	defaultOpenObservePass  = "Complexpass#123"
	defaultDashboardDir     = "monitoring/openobserve/bootstrap/dashboards"
	defaultAlertsFile       = "monitoring/openobserve/bootstrap/alerts/alerts.seed.yaml"
	defaultObservabilityURL = "http://localhost/api/v1/auth"
	defaultLogsStream       = "gochat_logs"
	defaultPostgresStream   = "gochat_postgres_probe_status"
	defaultTracesStream     = "gochat_traces"
	collectorHealthURL      = "http://localhost:13133/"
)

var (
	promQLMetricPattern   = regexp.MustCompile(`\bgochat_[a-zA-Z0-9_]+\b`)
	postgresProbeServices = []string{"gochat-api", "gochat-auth", "gochat-attachments", "gochat-ws"}
)

type openObserveClient struct {
	baseURL    string
	org        string
	user       string
	password   string
	httpClient *http.Client
}

type openObserveDashboard struct {
	Path        string
	DashboardID string
	Title       string
	FolderID    string
	Payload     map[string]any
}

type openObserveDashboardSummary struct {
	DashboardID string
	Title       string
	FolderID    string
	Hash        string
	PayloadHash string
}

type dashboardListResponse struct {
	Dashboards []map[string]any `json:"dashboards"`
}

type dashboardOperation struct {
	Method      string
	DashboardID string
	Title       string
	FolderID    string
	Hash        string
	Payload     map[string]any
}

type openObserveStream struct {
	Name       string                 `json:"name"`
	StreamType string                 `json:"stream_type"`
	Stats      openObserveStreamStats `json:"stats"`
}

type openObserveStreamStats struct {
	DocNum     int64 `json:"doc_num"`
	DocTimeMax int64 `json:"doc_time_max"`
}

type streamsResponse struct {
	List []openObserveStream `json:"list"`
}

type openObserveAlertSummary struct {
	AlertID  string
	Name     string
	FolderID string
}

type alertsListResponse struct {
	List []map[string]any `json:"list"`
}

type alertSeedFile struct {
	Alerts []alertSeed `yaml:"alerts"`
}

type alertSeed struct {
	Name        string   `yaml:"name"`
	Signal      string   `yaml:"signal"`
	Metric      string   `yaml:"metric"`
	Filter      string   `yaml:"filter"`
	Window      string   `yaml:"window"`
	Threshold   string   `yaml:"threshold"`
	Aggregation string   `yaml:"aggregation"`
	GroupBy     []string `yaml:"group_by"`
	Severity    string   `yaml:"severity"`
}

type alertRequest struct {
	FolderID          string            `json:"folder_id,omitempty"`
	Name              string            `json:"name"`
	StreamType        string            `json:"stream_type"`
	StreamName        string            `json:"stream_name"`
	IsRealTime        bool              `json:"is_real_time"`
	QueryCondition    alertQuery        `json:"query_condition"`
	TriggerCondition  alertTrigger      `json:"trigger_condition"`
	Destinations      []string          `json:"destinations"`
	Description       string            `json:"description"`
	Enabled           bool              `json:"enabled"`
	ContextAttributes map[string]string `json:"context_attributes,omitempty"`
}

type alertQuery struct {
	Type            string          `json:"type"`
	SQL             string          `json:"sql,omitempty"`
	PromQL          string          `json:"promql,omitempty"`
	PromQLCondition *alertCondition `json:"promql_condition,omitempty"`
	SearchEventType string          `json:"search_event_type,omitempty"`
}

type alertTrigger struct {
	Period        int64  `json:"period"`
	Operator      string `json:"operator"`
	Threshold     any    `json:"threshold"`
	Frequency     int64  `json:"frequency"`
	FrequencyType string `json:"frequency_type"`
	Silence       int64  `json:"silence"`
	AlignTime     bool   `json:"align_time"`
}

type alertCondition struct {
	Column     string `json:"column"`
	Operator   string `json:"operator"`
	Value      any    `json:"value"`
	IgnoreCase bool   `json:"ignore_case,omitempty"`
}

type searchRequest struct {
	Query searchQuery `json:"query"`
}

type searchQuery struct {
	SQL       string `json:"sql"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	From      int    `json:"from"`
	Size      int    `json:"size"`
}

type smokeConfig struct {
	client   *openObserveClient
	probeURL string
	timeout  time.Duration
	project  string
}

type observabilitySettings struct {
	url              string
	org              string
	user             string
	password         string
	dashboardDir     string
	alertsFile       string
	alertDestination string
}

func observability() *cli.Command {
	return &cli.Command{
		Name:  "observability",
		Usage: "Bootstrap and validate the local OpenObserve observability stack",
		Commands: []*cli.Command{
			observabilityBootstrap(),
			observabilitySmoke(),
			observabilityCleanup(),
		},
	}
}

func observabilityBootstrap() *cli.Command {
	return &cli.Command{
		Name:  "bootstrap",
		Usage: "Upsert OpenObserve dashboards and alerts from repo-managed assets",
		Flags: observabilityCommonFlags(true),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, settings, err := newOpenObserveClientFromCommand(cmd)
			if err != nil {
				return err
			}

			dashboardDir, err := absolutePath(settings.dashboardDir)
			if err != nil {
				return err
			}
			alertsFile, err := absolutePath(settings.alertsFile)
			if err != nil {
				return err
			}

			dashboards, err := loadDashboardDefinitions(dashboardDir)
			if err != nil {
				return err
			}
			existingDashboards, err := client.ListDashboards(ctx)
			if err != nil {
				return fmt.Errorf("list dashboards: %w", err)
			}
			for _, plan := range buildDashboardSyncPlans(dashboards, existingDashboards) {
				switch plan.Action {
				case dashboardActionCreate:
					if err := client.CreateDashboard(ctx, plan.Operation); err != nil {
						return fmt.Errorf("create dashboard %q: %w", plan.Title, err)
					}
				case dashboardActionUpdate:
					if err := client.UpdateDashboard(ctx, plan.Operation); err != nil {
						return fmt.Errorf("update dashboard %q: %w", plan.Title, err)
					}
				case dashboardActionInSync:
				default:
					return fmt.Errorf("unsupported dashboard action %q for %q", plan.Action, plan.Title)
				}
				fmt.Printf("%s dashboard %s\n", plan.Action, plan.Title)
			}

			dashboardStatuses, err := verifyDashboardSync(ctx, client, dashboardDir)
			if err != nil {
				return err
			}
			for _, status := range dashboardStatuses {
				fmt.Printf("verified dashboard %s (%s)\n", status.Title, status.Action)
			}

			if settings.alertDestination == "" {
				fmt.Println("alerts skipped: --alert-destination-id not provided")
				return nil
			}

			destinationNames, err := client.ListDestinationNames(ctx)
			if err != nil {
				return fmt.Errorf("list alert destinations: %w", err)
			}
			if len(destinationNames) == 0 {
				return errors.New("no OpenObserve alert destinations are configured; create one before bootstrapping alerts")
			}
			if !slices.Contains(destinationNames, settings.alertDestination) {
				return fmt.Errorf("alert destination %q was not found; available destinations: %s", settings.alertDestination, strings.Join(destinationNames, ", "))
			}

			seeds, err := loadAlertSeeds(alertsFile)
			if err != nil {
				return err
			}
			existingAlerts, err := client.ListAlerts(ctx)
			if err != nil {
				return fmt.Errorf("list alerts: %w", err)
			}
			for _, seed := range seeds {
				req, err := buildAlertRequest(seed, settings.alertDestination)
				if err != nil {
					return fmt.Errorf("build alert %q: %w", seed.Name, err)
				}

				existing, ok := existingAlerts[strings.ToLower(seed.Name)]
				if ok {
					if err := client.UpdateAlert(ctx, existing.AlertID, req); err != nil {
						return fmt.Errorf("update alert %q: %w", seed.Name, err)
					}
					fmt.Printf("updated alert %s (%s)\n", seed.Name, existing.AlertID)
					continue
				}
				if err := client.CreateAlert(ctx, req); err != nil {
					return fmt.Errorf("create alert %q: %w", seed.Name, err)
				}
				fmt.Printf("created alert %s\n", seed.Name)
			}

			return nil
		},
	}
}

func observabilitySmoke() *cli.Command {
	return &cli.Command{
		Name:  "smoke",
		Usage: "Verify local OpenObserve, collector, and signal ingestion end-to-end",
		Flags: append(observabilityCommonFlags(false),
			&cli.StringFlag{Name: "probe-url", Value: defaultObservabilityURL, Usage: "safe local HTTP URL used to generate a request log, metric, and trace"},
			&cli.DurationFlag{Name: "timeout", Value: 90 * time.Second, Usage: "maximum time to wait for all three signals to appear in OpenObserve"},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, _, err := newOpenObserveClientFromCommand(cmd)
			if err != nil {
				return err
			}
			cfg := smokeConfig{
				client:   client,
				probeURL: strings.TrimSpace(cmd.String("probe-url")),
				timeout:  cmd.Duration("timeout"),
				project:  composeProjectName(),
			}
			return runObservabilitySmoke(ctx, cfg)
		},
	}
}

func observabilityCommonFlags(includeBootstrapPaths bool) []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{Name: "url", Value: envOrDefault("OPENOBSERVE_URL", defaultOpenObserveURL), Usage: "OpenObserve base URL"},
		&cli.StringFlag{Name: "org", Value: envOrDefault("OPENOBSERVE_ORG", defaultOpenObserveOrg), Usage: "OpenObserve organization"},
		&cli.StringFlag{Name: "user", Value: envOrDefault("OPENOBSERVE_USER", defaultOpenObserveUser), Usage: "OpenObserve user email"},
		&cli.StringFlag{Name: "password", Value: envOrDefault("OPENOBSERVE_PASSWORD", defaultOpenObservePass), Usage: "OpenObserve password"},
	}
	if includeBootstrapPaths {
		flags = append(flags,
			&cli.StringFlag{Name: "dashboard-dir", Value: defaultDashboardDir, Usage: "directory containing dashboard bootstrap JSON files"},
			&cli.StringFlag{Name: "alerts-file", Value: defaultAlertsFile, Usage: "alert seed YAML file"},
			&cli.StringFlag{Name: "alert-destination-id", Usage: "OpenObserve alert destination name to bind when creating alerts"},
		)
	}
	return flags
}

func newOpenObserveClientFromCommand(cmd *cli.Command) (*openObserveClient, observabilitySettings, error) {
	settings := observabilitySettings{
		url:              strings.TrimSpace(cmd.String("url")),
		org:              strings.TrimSpace(cmd.String("org")),
		user:             strings.TrimSpace(cmd.String("user")),
		password:         cmd.String("password"),
		dashboardDir:     strings.TrimSpace(cmd.String("dashboard-dir")),
		alertsFile:       strings.TrimSpace(cmd.String("alerts-file")),
		alertDestination: strings.TrimSpace(cmd.String("alert-destination-id")),
	}
	if settings.url == "" {
		return nil, settings, errors.New("--url is required")
	}
	if settings.org == "" {
		return nil, settings, errors.New("--org is required")
	}
	if settings.user == "" {
		return nil, settings, errors.New("--user is required")
	}
	if settings.password == "" {
		return nil, settings, errors.New("--password is required")
	}
	return &openObserveClient{
		baseURL:  strings.TrimRight(settings.url, "/"),
		org:      settings.org,
		user:     settings.user,
		password: settings.password,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, settings, nil
}

func (c *openObserveClient) ListDashboards(ctx context.Context) (map[string]openObserveDashboardSummary, error) {
	var resp dashboardListResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/%s/dashboards", url.PathEscape(c.org)), nil, nil, &resp); err != nil {
		return nil, err
	}

	out := make(map[string]openObserveDashboardSummary, len(resp.Dashboards)*2)
	for _, row := range resp.Dashboards {
		payload := extractDashboardPayload(row)
		payloadHash, err := dashboardPayloadHash(payload)
		if err != nil {
			return nil, fmt.Errorf("hash dashboard payload: %w", err)
		}
		summary := openObserveDashboardSummary{
			DashboardID: firstString(row, "dashboardId", "dashboard_id"),
			Title:       defaultIfEmpty(firstString(row, "title"), firstString(payload, "title")),
			FolderID:    firstString(row, "folderId", "folder_id"),
			Hash:        firstString(row, "hash"),
			PayloadHash: payloadHash,
		}
		if summary.FolderID == "" {
			summary.FolderID = "default"
		}
		if summary.DashboardID != "" {
			out["id:"+summary.DashboardID] = summary
		}
		if summary.Title != "" {
			out["title:"+strings.ToLower(summary.Title)] = summary
		}
	}
	return out, nil
}

func (c *openObserveClient) CreateDashboard(ctx context.Context, op dashboardOperation) error {
	query := url.Values{}
	query.Set("folder", defaultIfEmpty(op.FolderID, "default"))
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/api/%s/dashboards", url.PathEscape(c.org)), query, op.Payload, nil)
}

func (c *openObserveClient) UpdateDashboard(ctx context.Context, op dashboardOperation) error {
	query := url.Values{}
	query.Set("folder", defaultIfEmpty(op.FolderID, "default"))
	if op.Hash != "" {
		query.Set("hash", op.Hash)
	}
	return c.do(ctx, http.MethodPut, fmt.Sprintf("/api/%s/dashboards/%s", url.PathEscape(c.org), url.PathEscape(op.DashboardID)), query, op.Payload, nil)
}

func (c *openObserveClient) ListAlerts(ctx context.Context) (map[string]openObserveAlertSummary, error) {
	var resp alertsListResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v2/%s/alerts", url.PathEscape(c.org)), nil, nil, &resp); err != nil {
		return nil, err
	}

	out := make(map[string]openObserveAlertSummary, len(resp.List))
	for _, row := range resp.List {
		summary := openObserveAlertSummary{
			AlertID:  firstString(row, "alert_id", "alertId", "id"),
			Name:     firstString(row, "name"),
			FolderID: firstString(row, "folder_id", "folderId"),
		}
		if summary.Name != "" {
			out[strings.ToLower(summary.Name)] = summary
		}
	}
	return out, nil
}

func (c *openObserveClient) ListDestinationNames(ctx context.Context) ([]string, error) {
	var raw []map[string]any
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/%s/alerts/destinations", url.PathEscape(c.org)), nil, nil, &raw); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(raw))
	for _, row := range raw {
		if name := firstString(row, "name", "destination_name"); name != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names, nil
}

func (c *openObserveClient) CreateAlert(ctx context.Context, req alertRequest) error {
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v2/%s/alerts", url.PathEscape(c.org)), nil, req, nil)
}

func (c *openObserveClient) UpdateAlert(ctx context.Context, alertID string, req alertRequest) error {
	req.FolderID = ""
	return c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v2/%s/alerts/%s", url.PathEscape(c.org), url.PathEscape(alertID)), nil, req, nil)
}

func (c *openObserveClient) ListStreams(ctx context.Context, streamType string) ([]openObserveStream, error) {
	query := url.Values{}
	query.Set("type", streamType)
	var resp streamsResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/%s/streams", url.PathEscape(c.org)), query, nil, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

func (c *openObserveClient) SearchRows(ctx context.Context, streamType, sql string, start, end int64) ([]map[string]any, error) {
	req := searchRequest{
		Query: searchQuery{
			SQL:       sql,
			StartTime: start,
			EndTime:   end,
			From:      0,
			Size:      20,
		},
	}
	path := fmt.Sprintf("/api/%s/_search", url.PathEscape(c.org))
	query := url.Values{}
	query.Set("type", streamType)

	var raw map[string]any
	if err := c.do(ctx, http.MethodPost, path, query, req, &raw); err != nil {
		return nil, err
	}
	return extractSearchRows(raw), nil
}

func (c *openObserveClient) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var payload io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, payload)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, out)
}

func loadDashboardDefinitions(dir string) ([]openObserveDashboard, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dashboard dir %q: %w", dir, err)
	}

	dashboards := make([]openObserveDashboard, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		payloadBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read dashboard file %q: %w", path, err)
		}
		var payload map[string]any
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return nil, fmt.Errorf("parse dashboard file %q: %w", path, err)
		}
		normalizeDashboardPayload(payload)

		dashboard := openObserveDashboard{
			Path:        path,
			DashboardID: firstString(payload, "dashboardId", "dashboard_id"),
			Title:       firstString(payload, "title"),
			FolderID:    defaultIfEmpty(firstString(payload, "folderId", "folder_id"), "default"),
			Payload:     payload,
		}
		if dashboard.DashboardID == "" {
			return nil, fmt.Errorf("dashboard file %q is missing dashboardId", path)
		}
		if dashboard.Title == "" {
			return nil, fmt.Errorf("dashboard file %q is missing title", path)
		}
		dashboards = append(dashboards, dashboard)
	}

	slices.SortFunc(dashboards, func(a, b openObserveDashboard) int {
		return strings.Compare(a.Path, b.Path)
	})
	return dashboards, nil
}

func normalizeDashboardPayload(payload map[string]any) {
	tabs, ok := payload["tabs"].([]any)
	if !ok {
		return
	}

	for _, tabValue := range tabs {
		tab, ok := tabValue.(map[string]any)
		if !ok {
			continue
		}
		panels, ok := tab["panels"].([]any)
		if !ok {
			continue
		}
		for _, panelValue := range panels {
			panel, ok := panelValue.(map[string]any)
			if !ok {
				continue
			}
			queries, ok := panel["queries"].([]any)
			if !ok {
				continue
			}
			for _, queryValue := range queries {
				query, ok := queryValue.(map[string]any)
				if !ok {
					continue
				}
				fields, ok := query["fields"].(map[string]any)
				if !ok {
					continue
				}
				streamType := strings.TrimSpace(firstString(fields, "stream_type"))
				if streamType != "metrics" {
					continue
				}
				currentStream := strings.TrimSpace(firstString(fields, "stream"))
				if currentStream != "" && currentStream != "gochat-metrics" {
					continue
				}
				streamName := deriveMetricsStreamName(firstString(query, "query"))
				if streamName == "" {
					continue
				}
				fields["stream"] = streamName
			}
		}
	}
}

func deriveMetricsStreamName(query string) string {
	matches := promQLMetricPattern.FindAllString(query, -1)
	for _, match := range matches {
		name := strings.TrimSpace(match)
		switch {
		case strings.HasSuffix(name, "_bucket"):
			return strings.TrimSuffix(name, "_bucket")
		case strings.HasSuffix(name, "_total"):
			return strings.TrimSuffix(name, "_total")
		case strings.HasSuffix(name, "_sum"):
			return strings.TrimSuffix(name, "_sum")
		case strings.HasSuffix(name, "_count"):
			return strings.TrimSuffix(name, "_count")
		default:
			return name
		}
	}
	return ""
}

func planDashboardUpserts(local []openObserveDashboard, existing map[string]openObserveDashboardSummary) []dashboardOperation {
	ops := make([]dashboardOperation, 0, len(local))
	for _, dashboard := range local {
		payload := cloneMap(dashboard.Payload)
		if match, ok := existing["id:"+dashboard.DashboardID]; ok {
			payload["dashboardId"] = match.DashboardID
			ops = append(ops, dashboardOperation{
				Method:      http.MethodPut,
				DashboardID: match.DashboardID,
				Title:       dashboard.Title,
				FolderID:    defaultIfEmpty(match.FolderID, dashboard.FolderID),
				Hash:        match.Hash,
				Payload:     payload,
			})
			continue
		}
		if match, ok := existing["title:"+strings.ToLower(dashboard.Title)]; ok {
			payload["dashboardId"] = match.DashboardID
			ops = append(ops, dashboardOperation{
				Method:      http.MethodPut,
				DashboardID: match.DashboardID,
				Title:       dashboard.Title,
				FolderID:    defaultIfEmpty(match.FolderID, dashboard.FolderID),
				Hash:        match.Hash,
				Payload:     payload,
			})
			continue
		}
		ops = append(ops, dashboardOperation{
			Method:      http.MethodPost,
			DashboardID: dashboard.DashboardID,
			Title:       dashboard.Title,
			FolderID:    dashboard.FolderID,
			Payload:     payload,
		})
	}
	return ops
}

func loadAlertSeeds(path string) ([]alertSeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read alert seed %q: %w", path, err)
	}
	var file alertSeedFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse alert seed %q: %w", path, err)
	}
	return file.Alerts, nil
}

func buildAlertRequest(seed alertSeed, destination string) (alertRequest, error) {
	windowMinutes, err := parseWindowMinutes(seed.Window)
	if err != nil {
		return alertRequest{}, fmt.Errorf("parse window %q: %w", seed.Window, err)
	}
	operator, value, err := parseThreshold(seed.Threshold)
	if err != nil {
		return alertRequest{}, fmt.Errorf("parse threshold %q: %w", seed.Threshold, err)
	}

	req := alertRequest{
		FolderID:     "default",
		Name:         seed.Name,
		StreamType:   strings.TrimSpace(seed.Signal),
		IsRealTime:   false,
		Destinations: []string{destination},
		Description:  fmt.Sprintf("%s (%s window, severity %s)", seed.Name, seed.Window, defaultIfEmpty(seed.Severity, "medium")),
		Enabled:      true,
		ContextAttributes: map[string]string{
			"severity": defaultIfEmpty(seed.Severity, "medium"),
			"source":   "gochat-bootstrap",
		},
		TriggerCondition: alertTrigger{
			Period:        windowMinutes,
			Operator:      operator,
			Threshold:     value,
			Frequency:     maxInt64(1, minInt64(windowMinutes, 5)),
			FrequencyType: "minutes",
			Silence:       maxInt64(windowMinutes, 5),
			AlignTime:     true,
		},
	}

	switch seed.Name {
	case "HTTP 5xx spike":
		req.StreamType = "metrics"
		req.StreamName = "gochat_http_server_errors"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (service_name, http_route) (increase(gochat_http_server_errors_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "HTTP latency p95 regression":
		req.StreamType = "metrics"
		req.StreamName = "gochat_http_server_duration"
		req.TriggerCondition.Operator = ">"
		req.TriggerCondition.Threshold = int64(1500)
		req.QueryCondition = promQLAlertQuery(
			"histogram_quantile(0.95, sum by (le, service_name, http_route) (rate(gochat_http_server_duration_bucket[10m]))) * 1000",
			">",
			int64(1500),
		)
	case "Auth failure burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_http_server_auth_failures"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (service_name, http_route) (increase(gochat_http_server_auth_failures_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "HTTP rate-limit burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_http_server_rate_limit_hits"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (service_name, http_route) (increase(gochat_http_server_rate_limit_hits_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "HTTP idempotency anomaly":
		req.StreamType = "metrics"
		req.StreamName = "gochat_http_server_idempotency_hits"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (service_name, http_route) (increase(gochat_http_server_idempotency_hits_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "Indexer stalled":
		req.StreamType = "metrics"
		req.StreamName = "gochat_indexer_consume_success"
		req.TriggerCondition.Operator = "<="
		req.TriggerCondition.Threshold = int64(0)
		req.QueryCondition = promQLAlertQuery("sum(increase(gochat_indexer_consume_success_total[15m]))", "<=", int64(0))
	case "Embedder stalled":
		req.StreamType = "metrics"
		req.StreamName = "gochat_embedder_consume_success"
		req.TriggerCondition.Operator = "<="
		req.TriggerCondition.Threshold = int64(0)
		req.QueryCondition = promQLAlertQuery("sum(increase(gochat_embedder_consume_success_total[15m]))", "<=", int64(0))
	case "Worker decode failures":
		req.StreamType = "metrics"
		req.StreamName = "gochat_indexer_decode_failure"
		req.QueryCondition = promQLAlertQuery(
			"sum(increase(gochat_indexer_decode_failure_total[10m])) + sum(increase(gochat_embedder_decode_failure_total[10m]))",
			">",
			int64(0),
		)
	case "WS auth failure burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_ws_auth_failures"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (reason) (increase(gochat_ws_auth_failures_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "WS heartbeat timeout burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_ws_heartbeat_timeouts"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum(increase(gochat_ws_heartbeat_timeouts_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "WS dropped delivery burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_ws_messages_dropped"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (topic) (increase(gochat_ws_messages_dropped_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "Redis/cache dependency error burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_dependency_errors"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (dependency_system, dependency_operation) (increase(gochat_dependency_errors_total{dependency_system=\"redis\",dependency_result=\"error\"}[%s]))", seed.Window),
			operator,
			value,
		)
	case "OpenSearch latency regression":
		req.StreamType = "metrics"
		req.StreamName = "gochat_dependency_duration"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("histogram_quantile(0.95, sum by (le, dependency_system, dependency_operation) (rate(gochat_dependency_duration_bucket{dependency_system=\"opensearch\"}[%s])))", seed.Window),
			operator,
			value,
		)
	case "S3 dependency failure burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_dependency_errors"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (dependency_system, dependency_operation) (increase(gochat_dependency_errors_total{dependency_system=\"s3\",dependency_result=\"error\"}[%s]))", seed.Window),
			operator,
			value,
		)
	case "SFU heartbeat failure logs":
		req.StreamType = "logs"
		req.StreamName = defaultLogsStream
		req.QueryCondition = sqlAlertQuery(`SELECT coalesce(voice_region, 'unknown') AS voice_region, coalesce(service_instance_id, 'unknown') AS service_instance_id, count(*) AS value FROM "gochat_logs" WHERE service_name = 'gochat-sfu' AND body LIKE '%heartbeat request failed%' GROUP BY voice_region, service_instance_id`)
	case "Postgres probe failing":
		req.StreamType = "metrics"
		req.StreamName = defaultPostgresStream
		req.QueryCondition = promQLAlertQuery("min by (service_name) (last_over_time(gochat_postgres_probe_status[5m]))", "<", int64(1))
	case "Postgres probe latency regression":
		req.StreamType = "metrics"
		req.StreamName = "gochat_postgres_probe_duration"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("histogram_quantile(0.95, sum by (le, service_name) (rate(gochat_postgres_probe_duration_bucket[%s])))", seed.Window),
			operator,
			value,
		)
	case "SFU bitrate disconnect burst":
		req.StreamType = "metrics"
		req.StreamName = "gochat_sfu_bitrate_disconnects"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("sum by (voice_region, service_instance_id) (increase(gochat_sfu_bitrate_disconnects_total[%s]))", seed.Window),
			operator,
			value,
		)
	case "SFU admin-close latency regression":
		req.StreamType = "metrics"
		req.StreamName = "gochat_sfu_admin_close_duration"
		req.QueryCondition = promQLAlertQuery(
			fmt.Sprintf("histogram_quantile(0.95, sum by (le, status, voice_region, service_instance_id) (rate(gochat_sfu_admin_close_duration_bucket[%s])))", seed.Window),
			operator,
			value,
		)
	case "Error log burst":
		req.StreamType = "logs"
		req.StreamName = defaultLogsStream
		req.QueryCondition = sqlAlertQuery(`SELECT service_name, count(*) AS value FROM "gochat_logs" WHERE level = 'ERROR' GROUP BY service_name`)
	default:
		return alertRequest{}, fmt.Errorf("unsupported alert seed %q", seed.Name)
	}

	return req, nil
}

func promQLAlertQuery(expr, operator string, value any) alertQuery {
	return alertQuery{
		Type:            "promql",
		PromQL:          expr,
		SearchEventType: "alerts",
		PromQLCondition: &alertCondition{
			Column:   "value",
			Operator: operator,
			Value:    value,
		},
	}
}

func sqlAlertQuery(statement string) alertQuery {
	return alertQuery{
		Type:            "sql",
		SQL:             statement,
		SearchEventType: "alerts",
	}
}

func runObservabilitySmoke(ctx context.Context, cfg smokeConfig) error {
	if cfg.timeout <= 0 {
		cfg.timeout = 90 * time.Second
	}
	if strings.TrimSpace(cfg.probeURL) == "" {
		cfg.probeURL = defaultObservabilityURL
	}

	legacyContainers, err := runningLegacyContainers(ctx, cfg.project)
	if err != nil {
		return err
	}
	if len(legacyContainers) > 0 {
		return fmt.Errorf("legacy observability containers are still running for project %q: %s. Run `docker compose down --remove-orphans` first", cfg.project, strings.Join(legacyContainers, ", "))
	}

	if err := checkCollectorHealth(ctx); err != nil {
		return err
	}
	if _, err := cfg.client.ListDashboards(ctx); err != nil {
		return fmt.Errorf("OpenObserve auth check failed: %w", err)
	}
	dashboardDir, err := absolutePath(defaultDashboardDir)
	if err != nil {
		return err
	}
	if _, err := verifyDashboardSync(ctx, cfg.client, dashboardDir); err != nil {
		return err
	}

	metricsStreams, err := cfg.client.ListStreams(ctx, "metrics")
	if err != nil {
		return fmt.Errorf("list metrics streams: %w", err)
	}
	if !hasStream(metricsStreams, "gochat_http_server_requests") {
		return errors.New("gochat_http_server_requests metric stream is missing from OpenObserve")
	}
	if !hasStream(metricsStreams, defaultPostgresStream) {
		return fmt.Errorf("%s metric stream is missing from OpenObserve", defaultPostgresStream)
	}
	traceStreams, err := cfg.client.ListStreams(ctx, "traces")
	if err != nil {
		return fmt.Errorf("list trace streams: %w", err)
	}
	if !hasStream(traceStreams, defaultTracesStream) {
		return fmt.Errorf("%s trace stream is missing from OpenObserve", defaultTracesStream)
	}

	requestID := fmt.Sprintf("gctools-smoke-%d", time.Now().UnixNano())
	probeStart := time.Now().UTC().Add(-5 * time.Second)
	statusCode, err := issueProbeRequest(ctx, cfg.probeURL, requestID)
	if err != nil {
		return err
	}
	fmt.Printf("probe request completed with status %d and request id %s\n", statusCode, requestID)

	deadline := time.Now().Add(cfg.timeout)
	var logsStream string
	var logRow map[string]any
	var traceRows []map[string]any
	var metricRows []map[string]any
	var postgresProbeRows []map[string]any
	for time.Now().Before(deadline) {
		logStreams, err := cfg.client.ListStreams(ctx, "logs")
		if err == nil {
			logsStream = pickLogsStream(logStreams)
		}
		if logsStream != "" && logRow == nil {
			logRow, _ = findLogRow(ctx, cfg.client, logsStream, requestID, probeStart)
		}
		if logRow != nil && len(traceRows) == 0 {
			traceRows, _ = findTraceRows(ctx, cfg.client, logRow, probeStart)
		}
		if logRow != nil && len(metricRows) == 0 {
			metricRows, _ = findMetricRows(ctx, cfg.client, logRow, probeStart)
		}
		if len(postgresProbeRows) == 0 {
			postgresProbeRows, _ = findPostgresProbeRows(ctx, cfg.client, time.Now().UTC().Add(-10*time.Minute))
		}
		if logRow != nil && len(traceRows) > 0 && len(metricRows) > 0 && len(postgresProbeRows) > 0 {
			fmt.Printf("observability smoke succeeded: logs stream=%s traces=%d metrics=%d postgres_probes=%d\n", logsStream, len(traceRows), len(metricRows), len(postgresProbeRows))
			return nil
		}
		time.Sleep(3 * time.Second)
	}

	if logsStream == "" {
		return errors.New("OpenObserve still has no log streams; confirm the collector fluentforward receiver is healthy and restart the stack with `docker compose down --remove-orphans && docker compose up -d`")
	}
	if logRow == nil {
		return fmt.Errorf("did not find request_id %q in log stream %q", requestID, logsStream)
	}
	if len(traceRows) == 0 {
		return fmt.Errorf("did not find trace rows for request_id %q", requestID)
	}
	if len(postgresProbeRows) == 0 {
		return fmt.Errorf("did not find recent %s rows for the Postgres-backed services", defaultPostgresStream)
	}
	return fmt.Errorf("did not find recent gochat_http_server_requests rows matching request_id %q route/service", requestID)
}

func checkCollectorHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, collectorHealthURL, nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("collector health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("collector health check returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func issueProbeRequest(ctx context.Context, probeURL, requestID string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return 0, fmt.Errorf("probe request failed: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func runningLegacyContainers(ctx context.Context, project string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("docker ps failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("run docker ps: %w", err)
	}

	prefix := strings.ToLower(strings.TrimSpace(project)) + "-"
	legacy := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		lower := strings.ToLower(name)
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		if strings.Contains(lower, "prometheus") || strings.Contains(lower, "grafana") || strings.Contains(lower, "promtail") || strings.Contains(lower, "loki") || strings.Contains(lower, "postgres-exporter") {
			legacy = append(legacy, name)
		}
	}
	slices.Sort(legacy)
	return legacy, nil
}

func findLogRow(ctx context.Context, client *openObserveClient, logsStream, requestID string, start time.Time) (map[string]any, error) {
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE request_id = '%s' ORDER BY _timestamp DESC LIMIT 5`, logsStream, escapeSQLString(requestID))
	rows, err := client.SearchRows(ctx, "logs", sql, start.UnixMicro(), time.Now().UTC().UnixMicro())
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func findTraceRows(ctx context.Context, client *openObserveClient, logRow map[string]any, start time.Time) ([]map[string]any, error) {
	traceID := firstString(logRow, "trace_id")
	if traceID == "" {
		return nil, nil
	}
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE trace_id = '%s' ORDER BY _timestamp DESC LIMIT 10`, defaultTracesStream, escapeSQLString(traceID))
	return client.SearchRows(ctx, "traces", sql, start.UnixMicro(), time.Now().UTC().UnixMicro())
}

func findMetricRows(ctx context.Context, client *openObserveClient, logRow map[string]any, start time.Time) ([]map[string]any, error) {
	rows, err := client.SearchRows(ctx, "metrics", `SELECT * FROM "gochat_http_server_requests" ORDER BY _timestamp DESC LIMIT 20`, start.UnixMicro(), time.Now().UTC().UnixMicro())
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	serviceName := normalizeMetricValue(firstString(logRow, "service.name", "service_name"))
	route := firstString(logRow, "route", "http_route")
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if serviceName != "" {
			currentService := normalizeMetricValue(firstString(row, "service_name", "service", "service.name"))
			if currentService != "" && currentService != serviceName {
				continue
			}
		}
		if route != "" {
			currentRoute := firstString(row, "http_route", "route")
			if currentRoute != "" && currentRoute != route {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	if len(filtered) == 0 {
		return nil, nil
	}
	return filtered, nil
}

func findPostgresProbeRows(ctx context.Context, client *openObserveClient, start time.Time) ([]map[string]any, error) {
	rows, err := client.SearchRows(ctx, "metrics", fmt.Sprintf(`SELECT * FROM "%s" ORDER BY _timestamp DESC LIMIT 50`, defaultPostgresStream), start.UnixMicro(), time.Now().UTC().UnixMicro())
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		serviceName := normalizeMetricValue(firstString(row, "service_name", "service", "service.name"))
		for _, allowed := range postgresProbeServices {
			if serviceName == normalizeMetricValue(allowed) {
				filtered = append(filtered, row)
				break
			}
		}
	}
	if len(filtered) == 0 {
		return nil, nil
	}
	return filtered, nil
}

func pickLogsStream(streams []openObserveStream) string {
	if hasStream(streams, defaultLogsStream) {
		return defaultLogsStream
	}
	for _, stream := range streams {
		if strings.Contains(strings.ToLower(stream.Name), "gochat") {
			return stream.Name
		}
	}
	return ""
}

func hasStream(streams []openObserveStream, name string) bool {
	for _, stream := range streams {
		if stream.Name == name {
			return true
		}
	}
	return false
}

func parseWindowMinutes(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasSuffix(raw, "m") {
		return 0, fmt.Errorf("unsupported window %q", raw)
	}
	value, err := strconv.ParseInt(strings.TrimSuffix(raw, "m"), 10, 64)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("window must be positive")
	}
	return value, nil
}

func parseThreshold(raw string) (string, any, error) {
	raw = strings.TrimSpace(raw)
	for _, operator := range []string{">=", "<=", "!=", ">", "<", "="} {
		if strings.HasPrefix(raw, operator) {
			value := strings.TrimSpace(strings.TrimPrefix(raw, operator))
			if value == "" {
				return "", nil, fmt.Errorf("missing threshold value")
			}
			if strings.Contains(value, ".") {
				parsed, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return "", nil, err
				}
				return operator, parsed, nil
			}
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return "", nil, err
			}
			return operator, parsed, nil
		}
	}
	return "", nil, fmt.Errorf("unsupported threshold %q", raw)
}

func extractSearchRows(raw map[string]any) []map[string]any {
	for _, key := range []string{"hits", "data"} {
		value, ok := raw[key]
		if !ok {
			continue
		}
		items, ok := value.([]any)
		if !ok {
			continue
		}
		rows := make([]map[string]any, 0, len(items))
		for _, item := range items {
			row, ok := item.(map[string]any)
			if ok {
				rows = append(rows, row)
			}
		}
		if len(rows) > 0 {
			return rows
		}
	}
	return nil
}

func composeProjectName() string {
	if value := strings.TrimSpace(os.Getenv("COMPOSE_PROJECT_NAME")); value != "" {
		return value
	}
	wd, err := os.Getwd()
	if err != nil {
		return "gochat"
	}
	return filepath.Base(wd)
}

func absolutePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(path) {
		return path, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return typed
			}
		case fmt.Stringer:
			text := strings.TrimSpace(typed.String())
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func deepCloneMap(in map[string]any) (map[string]any, error) {
	if in == nil {
		return map[string]any{}, nil
	}
	buf, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(buf, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func extractDashboardPayload(row map[string]any) map[string]any {
	for _, key := range []string{"v8", "v7", "v6", "v5", "v4", "v3", "v2", "v1"} {
		if value, ok := row[key].(map[string]any); ok {
			payload, err := deepCloneMap(value)
			if err == nil {
				normalizeDashboardPayload(payload)
				return payload
			}
		}
	}
	payload, err := deepCloneMap(row)
	if err != nil {
		return map[string]any{}
	}
	normalizeDashboardPayload(payload)
	return payload
}

func dashboardPayloadHash(payload map[string]any) (string, error) {
	canonical, err := canonicalDashboardPayload(payload)
	if err != nil {
		return "", err
	}
	buf, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(buf)
	return fmt.Sprintf("%x", sum[:]), nil
}

func canonicalDashboardPayload(payload map[string]any) (map[string]any, error) {
	cloned, err := deepCloneMap(payload)
	if err != nil {
		return nil, err
	}
	delete(cloned, "dashboardId")
	delete(cloned, "dashboard_id")
	delete(cloned, "owner")
	delete(cloned, "role")
	delete(cloned, "created")
	delete(cloned, "updatedAt")
	delete(cloned, "hash")
	delete(cloned, "folderId")
	delete(cloned, "folder_id")
	delete(cloned, "folder_name")
	delete(cloned, "version")
	normalizeDashboardPayload(cloned)
	return cloned, nil
}

func normalizeMetricValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ".", "_")
	return strings.ReplaceAll(value, "-", "_")
}

func escapeSQLString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
