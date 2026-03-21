package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	cli "github.com/urfave/cli/v3"
)

const (
	dashboardActionCreate = "created"
	dashboardActionUpdate = "updated"
	dashboardActionInSync = "in-sync"
)

var legacyMetricStreamPrefixes = []string{
	"pg_",
	"scrape_",
	"promhttp_",
	"postgres_exporter_",
	"citus_",
	"go_",
	"process_",
	"http_client_",
}

type dashboardSyncPlan struct {
	Title     string
	Action    string
	Operation dashboardOperation
}

func buildDashboardSyncPlans(local []openObserveDashboard, existing map[string]openObserveDashboardSummary) []dashboardSyncPlan {
	plans := make([]dashboardSyncPlan, 0, len(local))
	for _, dashboard := range local {
		localHash, err := dashboardPayloadHash(dashboard.Payload)
		if err != nil {
			plans = append(plans, dashboardSyncPlan{
				Title:  dashboard.Title,
				Action: dashboardActionUpdate,
				Operation: dashboardOperation{
					Method:      http.MethodPut,
					DashboardID: dashboard.DashboardID,
					Title:       dashboard.Title,
					FolderID:    dashboard.FolderID,
					Payload:     cloneMap(dashboard.Payload),
				},
			})
			continue
		}

		payload := cloneMap(dashboard.Payload)
		if match, ok := existing["id:"+dashboard.DashboardID]; ok {
			if match.PayloadHash == localHash {
				plans = append(plans, dashboardSyncPlan{Title: dashboard.Title, Action: dashboardActionInSync})
				continue
			}
			payload["dashboardId"] = match.DashboardID
			plans = append(plans, dashboardSyncPlan{
				Title:  dashboard.Title,
				Action: dashboardActionUpdate,
				Operation: dashboardOperation{
					Method:      http.MethodPut,
					DashboardID: match.DashboardID,
					Title:       dashboard.Title,
					FolderID:    defaultIfEmpty(match.FolderID, dashboard.FolderID),
					Hash:        match.Hash,
					Payload:     payload,
				},
			})
			continue
		}
		if match, ok := existing["title:"+strings.ToLower(dashboard.Title)]; ok {
			if match.PayloadHash == localHash {
				plans = append(plans, dashboardSyncPlan{Title: dashboard.Title, Action: dashboardActionInSync})
				continue
			}
			payload["dashboardId"] = match.DashboardID
			plans = append(plans, dashboardSyncPlan{
				Title:  dashboard.Title,
				Action: dashboardActionUpdate,
				Operation: dashboardOperation{
					Method:      http.MethodPut,
					DashboardID: match.DashboardID,
					Title:       dashboard.Title,
					FolderID:    defaultIfEmpty(match.FolderID, dashboard.FolderID),
					Hash:        match.Hash,
					Payload:     payload,
				},
			})
			continue
		}
		plans = append(plans, dashboardSyncPlan{
			Title:  dashboard.Title,
			Action: dashboardActionCreate,
			Operation: dashboardOperation{
				Method:      http.MethodPost,
				DashboardID: dashboard.DashboardID,
				Title:       dashboard.Title,
				FolderID:    dashboard.FolderID,
				Payload:     payload,
			},
		})
	}
	return plans
}

func verifyDashboardSync(ctx context.Context, client *openObserveClient, dashboardDir string) ([]dashboardSyncPlan, error) {
	dashboards, err := loadDashboardDefinitions(dashboardDir)
	if err != nil {
		return nil, err
	}
	existingDashboards, err := client.ListDashboards(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dashboards: %w", err)
	}
	plans := buildDashboardSyncPlans(dashboards, existingDashboards)
	drifted := make([]string, 0)
	for _, plan := range plans {
		if plan.Action != dashboardActionInSync {
			drifted = append(drifted, fmt.Sprintf("%s (%s)", plan.Title, plan.Action))
		}
	}
	if len(drifted) > 0 {
		return plans, fmt.Errorf("OpenObserve dashboards are missing or drifted: %s. Run `go run ./cmd/tools observability bootstrap --url %s --org %s --user %s --password <password>`", strings.Join(drifted, ", "), client.baseURL, client.org, client.user)
	}
	return plans, nil
}

func observabilityCleanup() *cli.Command {
	return &cli.Command{
		Name:  "cleanup",
		Usage: "Show or delete legacy OpenObserve metric streams left from earlier migrations",
		Flags: append(observabilityCommonFlags(false),
			&cli.BoolFlag{Name: "delete-legacy-streams", Usage: "delete matching legacy metric streams instead of printing a dry-run report"},
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, _, err := newOpenObserveClientFromCommand(cmd)
			if err != nil {
				return err
			}

			streams, err := client.ListStreams(ctx, "metrics")
			if err != nil {
				return fmt.Errorf("list metrics streams: %w", err)
			}
			legacy := filterLegacyMetricStreams(streams)
			if len(legacy) == 0 {
				fmt.Println("no legacy metric streams matched cleanup rules")
				return nil
			}

			deleteStreams := cmd.Bool("delete-legacy-streams")
			action := "dry-run"
			if deleteStreams {
				action = "delete"
			}
			fmt.Printf("%s legacy metric streams (%d matches)\n", action, len(legacy))
			for _, stream := range legacy {
				fmt.Printf("%s docs=%d latest=%s\n", stream.Name, stream.Stats.DocNum, formatUnixMicros(stream.Stats.DocTimeMax))
				if !deleteStreams {
					continue
				}
				if err := client.DeleteStream(ctx, stream.Name, "metrics", true); err != nil {
					return fmt.Errorf("delete metric stream %q: %w", stream.Name, err)
				}
				fmt.Printf("scheduled delete for %s\n", stream.Name)
			}

			if deleteStreams {
				fmt.Println("stream deletion is asynchronous in OpenObserve and may take several minutes to complete")
			}
			return nil
		},
	}
}

func filterLegacyMetricStreams(streams []openObserveStream) []openObserveStream {
	filtered := make([]openObserveStream, 0)
	for _, stream := range streams {
		if stream.StreamType != "" && stream.StreamType != "metrics" {
			continue
		}
		if !legacyMetricStreamMatches(stream.Name) {
			continue
		}
		filtered = append(filtered, stream)
	}
	slices.SortFunc(filtered, func(a, b openObserveStream) int {
		switch {
		case a.Stats.DocNum > b.Stats.DocNum:
			return -1
		case a.Stats.DocNum < b.Stats.DocNum:
			return 1
		default:
			return strings.Compare(a.Name, b.Name)
		}
	})
	return filtered
}

func legacyMetricStreamMatches(name string) bool {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return false
	}
	if name == "up" {
		return true
	}
	for _, prefix := range legacyMetricStreamPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func formatUnixMicros(value int64) string {
	if value <= 0 {
		return "unknown"
	}
	return time.UnixMicro(value).UTC().Format(time.RFC3339)
}

func (c *openObserveClient) DeleteStream(ctx context.Context, streamName, streamType string, deleteAll bool) error {
	if c == nil {
		return errors.New("OpenObserve client is nil")
	}
	query := url.Values{}
	query.Set("type", streamType)
	query.Set("delete_all", fmt.Sprintf("%t", deleteAll))
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/%s/streams/%s", url.PathEscape(c.org), url.PathEscape(streamName)), query, nil, nil)
}
