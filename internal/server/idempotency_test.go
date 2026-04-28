package server

import (
	"fmt"
	"io"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	fiberidempotency "github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/redis/go-redis/v9"
)

func TestRedisIdempotencyGetMissReturnsNil(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis run failed: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer client.Close()

	store := NewRedisIdempotency(client)
	val, err := store.Get("missing")
	if err != nil {
		t.Fatalf("Get returned error for cache miss: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil value for cache miss, got %q", string(val))
	}
}

func TestRedisLockerSerializesConcurrentIdempotentRequestsAcrossInstances(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis run failed: %v", err)
	}
	defer mini.Close()

	var executions atomic.Int32
	newApp := func() (*fiber.App, *redis.Client) {
		client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
		app := fiber.New()
		app.Use(fiberidempotency.New(fiberidempotency.Config{
			Lifetime:          time.Minute,
			KeyHeader:         "X-Idempotency-Key",
			KeyHeaderValidate: func(string) error { return nil },
			Storage:           NewRedisIdempotency(client),
			Lock:              NewRedisLocker(client),
		}))
		app.Post("/", func(c *fiber.Ctx) error {
			execution := executions.Add(1)
			time.Sleep(100 * time.Millisecond)
			return c.SendString(fmt.Sprintf("execution-%d", execution))
		})
		return app, client
	}

	app1, client1 := newApp()
	defer client1.Close()
	app2, client2 := newApp()
	defer client2.Close()

	type result struct {
		body string
		err  error
	}
	run := func(app *fiber.App, out chan<- result) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("X-Idempotency-Key", "shared-key")
		resp, err := app.Test(req, -1)
		if err != nil {
			out <- result{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		out <- result{body: string(body), err: err}
	}

	results := make(chan result, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		run(app1, results)
	}()
	go func() {
		defer wg.Done()
		run(app2, results)
	}()
	wg.Wait()
	close(results)

	var bodies []string
	for res := range results {
		if res.err != nil {
			t.Fatalf("request failed: %v", res.err)
		}
		bodies = append(bodies, res.body)
	}

	if got := executions.Load(); got != 1 {
		t.Fatalf("expected exactly one handler execution, got %d", got)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected two responses, got %d", len(bodies))
	}
	if bodies[0] != "execution-1" || bodies[1] != "execution-1" {
		t.Fatalf("expected both responses to reuse the first execution, got %#v", bodies)
	}
}
