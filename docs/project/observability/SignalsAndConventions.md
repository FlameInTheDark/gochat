[<- Observability](README.md)

# Signals And Conventions

## Streams

- Logs are stored in `gochat_logs`.
- Traces are stored in `gochat_traces`.
- Metrics are stored as individual metric streams. Examples:
  - `gochat_http_server_requests`
  - `gochat_http_server_duration`
  - `gochat_postgres_probe_status`
  - `gochat_sfu_peers_active`
  - `gochat_logs_exporter_send_failure`

## Query field conventions

OpenObserve normalizes dotted JSON keys into query-friendly columns. Use these field names in SQL queries and filters:

- `service_name`
- `deployment_environment`
- `request_id`
- `trace_id`
- `span_id`
- `user_id`
- `voice_region`
- `service_instance_id`
- `host_name`
- `dependency_system`
- `dependency_operation`
- `dependency_result`
- `http_route`

The raw JSON log payload still contains dotted keys such as `service.name` and `voice.region`, but queries should prefer the normalized forms above.

## Correlation

- HTTP:
  - inbound `traceparent` is honored
  - inbound `X-Request-ID` is honored
  - every HTTP response returns `X-Request-ID`
- NATS:
  - `traceparent`
  - `tracestate`
  - `baggage`
  - `X-Request-ID`
- Logs:
  - request-scoped logs should include `request_id`
  - active spans should add `trace_id` and `span_id`
  - authenticated request paths may include `user_id`

## Metric naming

- Metric names use dot notation in code, for example `gochat.sfu.heartbeats`.
- OpenObserve metric streams normalize those names to underscores, for example `gochat_sfu_heartbeats`.
- PromQL and SQL examples in repo-managed assets should prefer normalized label and field names such as `service_name`, `http_route`, `voice_region`, `service_instance_id`, `dependency_system`, `dependency_operation`, and `dependency_result`.
- Histogram helper streams also follow OpenTelemetry conventions:
  - `_bucket`
  - `_sum`
  - `_count`

## Redaction rules

These values must never be emitted to stdout logs or shipped directly to OpenObserve:

- access tokens, refresh tokens, webhook tokens, and cookies
- password reset email bodies
- raw request or response bodies
- message content
- attachment secrets or pre-signed storage credentials

Allowed alternatives:

- hash user email addresses when needed
- emit `[REDACTED]` for secrets
- log request metadata, result status, and stable identifiers instead of raw payloads
