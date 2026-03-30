[<- Observability](README.md)

# Local Development

## Supported local stack

The supported local workflow is:

```powershell
docker compose down --remove-orphans
docker compose up -d
go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
go run ./cmd/tools observability smoke --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
go run ./cmd/tools observability cleanup --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
```

This stack includes:

- OpenObserve
- the local OpenTelemetry Collector
- all non-SFU application services

It intentionally does not include the SFU.

## What local Compose validates

- collector health
- OpenObserve authentication
- dashboard bootstrap assets and live dashboard sync
- alert seed parsing and optional alert creation
- one synthetic HTTP request that must appear as a log, trace, and metric in OpenObserve
- recent native PostgreSQL probe metrics from at least one PostgreSQL-backed service

## Understanding large event counts

OpenObserve event totals in local development are usually dominated by metrics, not logs.

- Histogram bucket streams can produce large document counts quickly.
- Old local stacks may still contain stale exporter-era streams from removed Prometheus bridges.
- Use `go run ./cmd/tools observability cleanup ...` in dry-run mode before assuming there is active log spam.
- The shared Go telemetry runtime now defaults metric export to once per `60s`. Set `OTEL_METRIC_EXPORT_INTERVAL` only when you need temporarily denser metric resolution.

## Windows and Docker Desktop notes

- The local stack ships container logs through Docker's `fluentd` logging driver.
- If orphaned legacy observability containers still exist, the smoke check fails until you run `docker compose down --remove-orphans`.
- Processes started directly on the host are not collected by the Compose collector automatically. That includes a standalone SFU binary.

## Local standalone SFU testing

If you want to test the SFU locally, run it outside Compose and configure the local telemetry gateway endpoint:

- `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318`
- `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`
- `OTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer <jwt>`
- `OTEL_METRIC_EXPORT_INTERVAL=60000`
- `WEBHOOK_TOKEN=<same-jwt>`

See [External SFU](ExternalSFU.md) for the full example.
