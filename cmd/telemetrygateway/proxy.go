package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	cfgpkg "github.com/FlameInTheDark/gochat/cmd/telemetrygateway/config"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/serviceauth"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

type telemetryProxy struct {
	client          *http.Client
	log             *slog.Logger
	tokens          *serviceauth.TokenManager
	upstreamBaseURL *url.URL
	limiters        *tokenRateLimiters
}

type tokenRateLimiters struct {
	mu          sync.Mutex
	entries     map[string]*tokenRateLimiter
	rate        rate.Limit
	burst       int
	idleTTL     time.Duration
	lastCleanup time.Time
}

type tokenRateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newTelemetryProxy(cfg *cfgpkg.Config, logger *slog.Logger, tokens *serviceauth.TokenManager) (*telemetryProxy, error) {
	if cfg == nil {
		return nil, fmt.Errorf("telemetry gateway config is required")
	}
	if logger == nil {
		logger = observability.Logger()
	}
	baseURL, err := url.Parse(strings.TrimSpace(cfg.UpstreamEndpoint))
	if err != nil {
		return nil, fmt.Errorf("parse upstream endpoint: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("upstream endpoint must include scheme and host")
	}
	if tokens == nil {
		return nil, fmt.Errorf("token manager is required")
	}

	timeout := time.Duration(cfg.RequestTimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &telemetryProxy{
		client: &http.Client{
			Timeout: timeout,
		},
		log:             logger,
		tokens:          tokens,
		upstreamBaseURL: baseURL,
		limiters:        newTokenRateLimiters(cfg.RateLimitPerSecond, cfg.RateLimitBurst),
	}, nil
}

func (p *telemetryProxy) handle(signal string) fiber.Handler {
	upstreamURL := p.upstreamURL(signal)
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		token := serviceauth.ExtractBearerToken(c.Get(fiber.HeaderAuthorization))
		claims, err := p.tokens.Parse(token)
		if err != nil || claims.ServiceType != "sfu" || strings.TrimSpace(claims.ServiceID) == "" {
			p.log.WarnContext(ctx, "telemetry request rejected",
				slog.String("reason", "unauthorized"),
				slog.String("signal", signal),
				slog.String("ip", c.IP()),
			)
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		if !p.limiters.Allow(claims.ServiceID) {
			p.log.WarnContext(ctx, "telemetry request rejected",
				slog.String("reason", "rate_limited"),
				slog.String("signal", signal),
				slog.String("service.instance.id", claims.ServiceID),
			)
			return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(c.Body()))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to build upstream request")
		}
		copyHeaderIfPresent(req.Header, "Content-Type", c.Get(fiber.HeaderContentType))
		copyHeaderIfPresent(req.Header, "Content-Encoding", c.Get(fiber.HeaderContentEncoding))
		copyHeaderIfPresent(req.Header, "Accept", c.Get(fiber.HeaderAccept))
		copyHeaderIfPresent(req.Header, "User-Agent", c.Get(fiber.HeaderUserAgent))

		resp, err := p.client.Do(req)
		if err != nil {
			p.log.ErrorContext(ctx, "telemetry upstream request failed",
				slog.String("signal", signal),
				slog.String("service.instance.id", claims.ServiceID),
				slog.String("error", err.Error()),
			)
			return fiber.NewError(fiber.StatusBadGateway, "upstream request failed")
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			p.log.ErrorContext(ctx, "telemetry upstream response read failed",
				slog.String("signal", signal),
				slog.String("service.instance.id", claims.ServiceID),
				slog.String("error", readErr.Error()),
			)
			return fiber.NewError(fiber.StatusBadGateway, "unable to read upstream response")
		}

		if contentType := strings.TrimSpace(resp.Header.Get("Content-Type")); contentType != "" {
			c.Set(fiber.HeaderContentType, contentType)
		}
		return c.Status(resp.StatusCode).Send(body)
	}
}

func (p *telemetryProxy) upstreamURL(signal string) string {
	u := *p.upstreamBaseURL
	u.Path = path.Join("/", strings.TrimSuffix(p.upstreamBaseURL.Path, "/"), "v1", signal)
	return u.String()
}

func copyHeaderIfPresent(headers interface {
	Set(string, string)
}, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	headers.Set(key, value)
}

func newTokenRateLimiters(ratePerSecond float64, burst int) *tokenRateLimiters {
	if ratePerSecond <= 0 {
		ratePerSecond = 20
	}
	if burst <= 0 {
		burst = 40
	}
	return &tokenRateLimiters{
		entries: make(map[string]*tokenRateLimiter),
		rate:    rate.Limit(ratePerSecond),
		burst:   burst,
		idleTTL: 15 * time.Minute,
	}
}

func (r *tokenRateLimiters) Allow(id string) bool {
	if r == nil || strings.TrimSpace(id) == "" {
		return false
	}

	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lastCleanup.IsZero() || now.Sub(r.lastCleanup) >= r.idleTTL {
		for key, entry := range r.entries {
			if now.Sub(entry.lastSeen) >= r.idleTTL {
				delete(r.entries, key)
			}
		}
		r.lastCleanup = now
	}

	entry, ok := r.entries[id]
	if !ok {
		entry = &tokenRateLimiter{
			limiter: rate.NewLimiter(r.rate, r.burst),
		}
		r.entries[id] = entry
	}
	entry.lastSeen = now
	return entry.limiter.Allow()
}
