# OpenObserve Assets

This directory is the repo-managed source for the OpenObserve cutover.

Files:

- `dashboards.yaml`: human-editable dashboard intent and panel inventory.
- `alerts.yaml`: human-editable alert inventory and thresholds.
- `bootstrap/dashboards/*.dashboard.json`: repo-managed dashboard bootstrap payloads.
- `bootstrap/alerts/alerts.seed.yaml`: alert bootstrap seed to bind to your org's alert destinations.

## Local bootstrap

1. Start fresh with `docker compose down --remove-orphans`.
2. Bring the stack up with `docker compose up -d`.
3. Bootstrap OpenObserve assets with:
   `go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123`
4. If you already created an OpenObserve alert destination, bind alerts during bootstrap:
   `go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123 --alert-destination-id <destination-name>`
5. Run the smoke check after the stack settles:
   `go run ./cmd/tools observability smoke --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123`
6. Review legacy metric streams before deleting them:
   `go run ./cmd/tools observability cleanup --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123`
7. Delete only the exporter-era metric streams when you are ready:
   `go run ./cmd/tools observability cleanup --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123 --delete-legacy-streams`

Local Compose does not run the SFU service anymore. Voice/SFU dashboards and
alerts remain in OpenObserve for externally deployed SFU nodes. In this stage
the SFU ships traces, metrics, and best-effort logs directly to OpenObserve and
does not require a sidecar, daemon, or collector process on the host.

See `docs/project/observability/ExternalSFU.md` for the standalone SFU
environment contract and `docs/project/observability/README.md` for the full
project observability documentation set.

## Streams

- Logs are written to the `gochat_logs` stream.
- Metrics are queryable as individual metric streams such as `gochat_http_server_requests`.
- PostgreSQL availability is reported through native service-side probe streams such as `gochat_postgres_probe_status`.
- Traces are queryable through the `gochat_traces` stream.

## Volume note

If OpenObserve shows a very large event count in local development, it is usually metric-heavy volume rather than log spam.

- The hottest live streams are typically histogram bucket streams such as `gochat_dependency_duration_bucket` and `gochat_http_server_duration_bucket`.
- Older local stacks may also still contain stale exporter-era metric families such as `pg_*`, `scrape_*`, `promhttp_*`, `postgres_exporter_*`, `citus_*`, `go_*`, `process_*`, `http_client_*`, and `up`.
- Use `go run ./cmd/tools observability cleanup ...` first in dry-run mode to confirm how much of the volume is historical.

## Environment

- OTLP ingress for app services: `http://otel-collector:4318`
- Deployment environment override: `GOCHAT_DEPLOYMENT_ENV`
- OpenObserve org env in compose: `OPENOBSERVE_ORG`
- Collector health endpoint: `http://localhost:13133/`
- Standalone SFU direct logs env: `OPENOBSERVE_LOGS_ENABLED`,
  `OPENOBSERVE_LOGS_ENDPOINT`, `OPENOBSERVE_LOGS_AUTH`,
  `OPENOBSERVE_LOGS_STREAM`
- Standalone SFU direct OTLP env:
  `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`,
  `OTEL_EXPORTER_OTLP_TRACES_HEADERS`,
  `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`,
  `OTEL_EXPORTER_OTLP_METRICS_HEADERS`

## Windows / Docker Desktop note

The local stack now ships container logs through Docker's `fluentd` logging
driver into the collector on `localhost:24224`. That is the supported path for
Docker Desktop because the Docker daemon emits logs itself and does not use the
collector container's internal DNS name. Go processes started directly on the
host still need their own OTLP/log shipping if you want them to appear in
OpenObserve.
