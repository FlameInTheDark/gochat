# gochat Code Style Guide

This document is the authoritative style reference for all Go code in this repository. It extends the [Go standard style](https://google.github.io/styleguide/go/) with project-specific rules.

---

## 1. Naming

### 1.1 Initialisms must be all-caps
| Wrong | Correct |
|-------|---------|
| `BaseUrl` | `BaseURL` |
| `NatsConnString` | `NATSConnString` |
| `HttpError` | `HTTPError` |
| `userId` | `userID` |
| `guildId` | `guildID` |

Applies to: struct fields, variables, function parameters, config keys.

### 1.2 Handler structs
Unexported handler struct is named `handler`. Its constructor is always `New(...)`.
```go
// correct
type handler struct { ... }
func New(...) *handler { return &handler{...} }

// wrong — too generic, loses package context in stack traces
type entity struct { ... }
```

### 1.3 DB entity structs
DB package structs are named `Entity` (exported) and always accessed through their interface:
```go
// internal/database/pgentities/guild/entity.go
type Guild interface { ... }
type Entity struct { c *sqlx.DB }
func New(c *sqlx.DB) Guild { return &Entity{c: c} }
```
The concrete type must never escape the package.

### 1.4 Error sentinel variables
All package-level error strings follow the pattern `ErrVerbNoun` and live in `errors.go` or `scheme.go`:
```go
const (
    ErrUnableToGetUser     = "unable to get user"
    ErrPermissionsRequired = "permissions required"
    ErrIncorrectChannelID  = "incorrect channel ID"
)
```
- Use `const` for static strings, `var` only for `errors.New(...)` sentinel errors.
- Every string must be lowercase, no trailing punctuation.
- Never pass `err.Error()` directly as a fiber error message.

### 1.5 Test fake types
Test fake types are named `fake<Interface>` (lowercase, unexported):
```go
type fakeCache struct { testutil.Noop }
type fakeRoleRepo struct { ... }
```

---

## 2. Error Handling

### 2.1 HTTP handler errors
All HTTP errors use `fiber.NewError(status, ErrConstant)`:
```go
// correct
return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
return fiber.NewError(fiber.StatusNotFound, ErrChannelNotFound)

// wrong — leaks internal error details to clients
return fiber.NewError(fiber.StatusBadRequest, err.Error())
return fmt.Errorf("channel not found: %w", err)  // wrong in handler context
```

### 2.2 Internal errors (logged, not returned)
Use `fmt.Errorf("verb noun: %w", err)` to preserve the error chain:
```go
// correct
return fmt.Errorf("get guild channel %d: %w", channelID, err)

// wrong — context is lost
return err
return errors.New("failed")
```

### 2.3 Sentinel error matching
Always use `errors.Is` / `errors.As`, never string comparison:
```go
// correct
if errors.Is(err, sql.ErrNoRows) { ... }
if errors.Is(err, gocql.ErrNotFound) { ... }

// wrong
if err.Error() == "sql: no rows in result set" { ... }
```

### 2.4 Never ignore errors silently
```go
// wrong
_ = someOperation()

// correct — if the error truly cannot be acted on, log it
if err := someOperation(); err != nil {
    log.Error("failed to do X", "error", err)
}
```
Exception: cleanup in `defer` where you have already returned the primary error is acceptable, but log the cleanup error.

---

## 3. HTTP Handlers

### 3.1 Handler decomposition pattern
Every handler method that performs more than one logical step must be decomposed:
```
func (h *handler) DoThing(c *fiber.Ctx) error {
    req, user, id, err := h.parseDoThingRequest(c)   // 1. parse + auth
    if err != nil { return err }

    if err := h.validateDoThingPermissions(c, id, user.ID); err != nil {  // 2. authz
        return err
    }

    result, err := h.executeDoThing(c, req, user, id)  // 3. business logic
    if err != nil { return err }

    return c.JSON(result)
}
```

### 3.2 Context in goroutines
Goroutines spawned inside a Fiber handler must detach the context before capturing it. Fiber recycles `*Ctx` objects after the handler returns.
```go
// correct
bgCtx := observability.BackgroundFromContext(c.UserContext())
go func() { doWork(bgCtx) }()

// wrong — use-after-free
go func() { doWork(c.UserContext()) }()
```

### 3.3 Parameter parsing
All route parameters are parsed via helper functions that return the typed value and a fiber error:
```go
func parseChannelID(c *fiber.Ctx) (int64, error) {
    id, err := strconv.ParseInt(c.Params("channel_id"), 10, 64)
    if err != nil {
        return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
    }
    return id, nil
}
```
Never inline `strconv.ParseInt` in handler body more than once.

### 3.4 Swagger annotations
Every exported handler method has a complete Swagger annotation block immediately above it:
```go
// @Summary  Short action description
// @Tags     category
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    channel_id path int true "Channel ID"
// @Success  200 {object} dto.Message
// @Failure  400 {string} string "Bad request"
// @Failure  403 {string} string "Forbidden"
// @Failure  500 {string} string "Internal server error"
// @Router   /message/channel/{channel_id} [post]
func (h *handler) SendMessage(c *fiber.Ctx) error {
```

---

## 4. Dependency Injection

### 4.1 All dependencies via constructor
Entities never use global variables. Every dependency (DB repo, cache, MQ publisher, logger) is injected through `New(...)`:
```go
type handler struct {
    log  *slog.Logger
    ch   channelrepo.Channel
    cache cache.KV
    mqt  mq.Publisher
}

func New(log *slog.Logger, ch channelrepo.Channel, ...) *handler {
    return &handler{log: log, ch: ch, ...}
}
```

### 4.2 Nil dependencies are a construction-time error
If a dependency is required, validate at construction — never at call time:
```go
// correct
func New(cache cache.KV) (*handler, error) {
    if cache == nil {
        return nil, errors.New("cache is required")
    }
    return &handler{cache: cache}, nil
}

// wrong — runtime nil check leaks into every request
func (h *handler) requireCache() error {
    if h.cache == nil {
        return fiber.NewError(500, "cache not configured")
    }
    return nil
}
```

### 4.3 Interfaces at consumption points, not definition points
Define interfaces where they are used, not where the implementation lives:
```go
// internal/database/pgentities/guild/entity.go — defines its own interface
type Guild interface { GetGuildById(ctx, id) (model.Guild, error); ... }

// cmd/api/endpoints/guild — uses it by importing the interface type
type handler struct {
    guild guildiface.Guild
}
```

---

## 5. Testing

### 5.1 Use `testutil.Noop` for cache fakes
Fake cache implementations embed `testutil.Noop` and override only the methods under test:
```go
// cmd/api/endpoints/guild/roles_test.go
type fakeCache struct {
    testutil.Noop  // satisfies the full cache.Cache interface with zero values
}
// override only what this test needs
func (f *fakeCache) GetJSON(ctx context.Context, key string, v interface{}) error {
    return json.Unmarshal(f.stored[key], v)
}
```

### 5.2 Table-driven tests
All unit tests use table-driven format with named subtests:
```go
func TestSomething(t *testing.T) {
    cases := []struct {
        name    string
        input   SomeInput
        want    SomeOutput
        wantErr bool
    }{
        {name: "valid input", input: ..., want: ...},
        {name: "missing field", input: ..., wantErr: true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := doSomething(tc.input)
            if (err != nil) != tc.wantErr {
                t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
            }
            if !tc.wantErr && got != tc.want {
                t.Errorf("got %v, want %v", got, tc.want)
            }
        })
    }
}
```

### 5.3 Test file placement
- Handler logic: `internal/` package, not `cmd/*/main` package.
- Tests that exercise Fiber routing: use `package <name>_test` (black-box).
- Tests that test unexported helpers: use `package <name>` (white-box).
- Never place business logic tests in `package main`.

### 5.4 No real infrastructure in unit tests
Unit tests must never connect to Postgres, ScyllaDB, Redis, or NATS. Use fake implementations or interfaces. Integration tests that need real infrastructure live in `*_integration_test.go` files and are gated with `//go:build integration`.

---

## 6. Concurrency

### 6.1 Goroutine ownership
Every goroutine spawned must have an identifiable owner responsible for waiting on it. Prefer `sync.WaitGroup` or `errgroup` over fire-and-forget goroutines in request handlers. Fire-and-forget is acceptable for non-critical side effects (e.g., cache eviction) but must use a background context.

### 6.2 Context deadlines
Any outbound call (DB, cache, HTTP, NATS) must use a context derived from the request context:
```go
// correct
result, err := h.ch.GetChannel(c.UserContext(), channelID)

// wrong
result, err := h.ch.GetChannel(context.Background(), channelID)
```
Exception: goroutines performing background work after the request completes must use `observability.BackgroundFromContext(c.UserContext())`.

### 6.3 Mutex naming
Mutexes are named `mu` (field-scoped) or `<thing>Mu` (when multiple mutexes protect different fields):
```go
type handler struct {
    mu      sync.Mutex
    stateMu sync.RWMutex
}
```

---

## 7. Logging

### 7.1 Use structured `slog`
All logging uses `log/slog` with key-value pairs:
```go
// correct
log.Error("unable to send message", "channel_id", channelID, "error", err)

// wrong
log.Printf("unable to send message to channel %d: %v", channelID, err)
```

### 7.2 Log at the right level
| Level | When |
|-------|------|
| `Error` | Unrecoverable within the request; returned as 5xx |
| `Warn` | Degraded path taken; service continues |
| `Info` | Significant lifecycle events (startup, shutdown, migrations) |
| `Debug` | Per-request detail, off in production |

### 7.3 Logger from context
Inside handlers, always obtain a logger enriched with trace/span IDs from the request context:
```go
log := observability.LoggerFromFiber(c, h.log)
```

### 7.4 No log-and-return
Never log an error and also return it — the caller will log it too:
```go
// wrong — double-logged
log.Error("get channel failed", "error", err)
return fiber.NewError(500, ErrUnableToGetChannel)

// correct — log OR return, not both (at the same level)
return fiber.NewError(500, ErrUnableToGetChannel)
// the error middleware logs the 5xx
```
Exception: fire-and-forget goroutines where the error cannot propagate must log.

---

## 8. Configuration

### 8.1 Config struct conventions
- YAML and env tags must be present on every field.
- Sensitive fields must not silently fall back to empty values — use `env-required:"true"` or fail at startup validation.
- All config fields use `snake_case` in yaml/env tags.

### 8.2 Fail-fast on bad config
`LoadConfig` must validate that required secrets are non-empty before returning. Weak or local-placeholder secrets should emit warnings unless a service explicitly requires stricter validation.

---

## 9. Database

### 9.1 Query building
Use `Masterminds/squirrel` for dynamic queries; raw SQL strings only for static, single-use queries.

### 9.2 `sql.ErrNoRows` handling
When "not found" is a valid application state, translate `sql.ErrNoRows` at the repository layer, not in handlers:
```go
// internal/database/pgentities/guild/entity.go
func (e *Entity) GetGuildByID(ctx context.Context, id int64) (model.Guild, error) {
    var g model.Guild
    err := e.c.GetContext(ctx, &g, query, id)
    if errors.Is(err, sql.ErrNoRows) {
        return model.Guild{}, ErrGuildNotFound  // domain error
    }
    return g, err
}
```
Handlers then use `errors.Is(err, guildiface.ErrGuildNotFound)` to produce 404 responses.

### 9.3 Migrations
- One logical change per migration file.
- Migrations are always reversible (include a `down` migration).
- Never modify an existing migration after it has been merged to `main`.

---

## 10. Security

### 10.1 No raw secrets in source
Secrets never appear in source code, config files committed to the repo, or log output. Config files contain only references (env var names).

### 10.2 Input size limits
Every endpoint that accepts a request body enforces a maximum size. Fiber's `BodyLimit` is set at the server level; individual handlers add domain-specific limits (e.g., message content length).

### 10.3 Rate limiting
All public endpoints are covered by the global rate limiter. Endpoints that perform expensive operations (search, embed generation) have additional per-endpoint rate limiting.

### 10.4 Permission checks before data access
Permissions are always checked before the first DB read that produces data the caller might not be authorized to see:
```go
// correct order
if !canRead { return fiber.NewError(403, ErrPermissionsRequired) }
data, err := h.repo.GetSensitiveData(ctx, id)
```

---

## Quick Reference

```
File layout per endpoint package:
  handlers.go   — exported handler methods + Swagger annotations
  scheme.go     — request/response structs, validation, Err* constants
  errors.go     — (optional) Err* constants if scheme.go grows too large
  entity.go     — handler struct + New() constructor
  *_test.go     — tests

File layout per DB entity package:
  entity.go     — interface + Entity struct + New()
  queries.go    — SQL query constants (optional)
  errors.go     — domain-level ErrXxx sentinel errors
```
