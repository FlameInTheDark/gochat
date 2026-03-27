package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cfgpkg "github.com/FlameInTheDark/gochat/cmd/telemetrygateway/config"
	"github.com/FlameInTheDark/gochat/internal/serviceauth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TestTelemetryProxyForwardsAuthorizedRequest(t *testing.T) {
	reqBodyCh := make(chan string, 1)
	authHeaderCh := make(chan string, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		reqBodyCh <- string(body)
		authHeaderCh <- r.Header.Get("Authorization")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	}))
	defer upstream.Close()

	proxy := newTestProxy(t, upstream.URL, 100, 100)
	app := fiber.New()
	app.Post("/v1/logs", proxy.handle("logs"))

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", strings.NewReader("payload"))
	req.Header.Set("Authorization", "Bearer "+signedToken(t, "supersecret", "sfu", "node-1"))
	req.Header.Set("Content-Type", "application/x-protobuf")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusAccepted)
	}
	if body := <-reqBodyCh; body != "payload" {
		t.Fatalf("forwarded body = %q, want %q", body, "payload")
	}
	if auth := <-authHeaderCh; auth != "" {
		t.Fatalf("upstream Authorization should be stripped, got %q", auth)
	}
}

func TestTelemetryProxyRejectsInvalidToken(t *testing.T) {
	proxy := newTestProxy(t, "http://example.com", 100, 100)
	app := fiber.New()
	app.Post("/v1/traces", proxy.handle("traces"))

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", strings.NewReader("payload"))
	req.Header.Set("Authorization", "Bearer bad-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestTelemetryProxyRejectsRateLimitedToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()

	proxy := newTestProxy(t, upstream.URL, 1, 1)
	app := fiber.New()
	app.Post("/v1/metrics", proxy.handle("metrics"))

	token := signedToken(t, "supersecret", "sfu", "node-1")
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/metrics", strings.NewReader("payload"))
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()

		if i == 0 && resp.StatusCode != http.StatusAccepted {
			t.Fatalf("first status = %d, want %d", resp.StatusCode, http.StatusAccepted)
		}
		if i == 1 && resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("second status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
		}
	}
}

func newTestProxy(t *testing.T, upstreamURL string, rateLimit float64, burst int) *telemetryProxy {
	t.Helper()
	proxy, err := newTelemetryProxy(&cfgpkg.Config{
		UpstreamEndpoint:   upstreamURL,
		RequestTimeoutMS:   1000,
		RateLimitPerSecond: rateLimit,
		RateLimitBurst:     burst,
	}, nil, serviceauth.NewTokenManager("supersecret"))
	if err != nil {
		t.Fatalf("newTelemetryProxy: %v", err)
	}
	return proxy
}

func signedToken(t *testing.T, secret, typ, id string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"typ": typ,
		"id":  id,
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	return signed
}
