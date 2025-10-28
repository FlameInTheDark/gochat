package sfu

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// promTargetGroup matches Prometheus HTTP SD target group JSON structure.
type promTargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// PromSD serves a Prometheus HTTP Service Discovery document for SFU instances.
// It discovers instances via etcd manager and returns one target group per instance
// with labels: region and sfu_id. Targets are host:port values for /metrics path.
func (e *entity) PromSD(c *fiber.Ctx) error {
	// Require auth: accept either X-Webhook-Token (custom) or Authorization: Bearer <token>
	token := c.Get(hdrToken)
	if token == "" {
		authz := c.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			token = strings.TrimSpace(authz[7:])
		}
	}
	if !e.tokens.Validate("prom", "", token) {
		return fiber.ErrUnauthorized
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
	defer cancel()

	regions, err := e.disco.Regions(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "discovery error")
	}

	var groups []promTargetGroup
	for _, r := range regions {
		insts, err := e.disco.List(ctx, r)
		if err != nil {
			continue
		}
		for _, inst := range insts {
			// Derive host:port for metrics from public URL (ws[s]://host[:port]/signal)
			hostport := deriveMetricsTarget(inst.URL)
			if hostport == "" {
				continue
			}
			groups = append(groups, promTargetGroup{
				Targets: []string{hostport},
				Labels: map[string]string{
					"region": r,
					"sfu_id": inst.ID,
				},
			})
		}
	}

	return c.JSON(groups)
}

// deriveMetricsTarget converts a ws(s) public URL into a Prometheus target "host:port".
// If no port is present, defaults to 80 for ws and 443 for wss.
func deriveMetricsTarget(publicURL string) string {
	u, err := url.Parse(publicURL)
	if err != nil {
		return ""
	}
	host := u.Host
	if host == "" {
		return ""
	}
	// If host already includes :port, keep it. Otherwise infer from scheme.
	if strings.Contains(host, ":") {
		return host
	}
	switch strings.ToLower(u.Scheme) {
	case "wss", "https":
		return host + ":443"
	default: // ws/http
		return host + ":80"
	}
}
