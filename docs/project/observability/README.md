[<- Documentation](../README.md)

# Observability

This section documents how telemetry works across the project today.

- [Architecture](Architecture.md) explains the signal flow for local services, the collector, OpenObserve, and the externally deployed SFU.
- [Signals And Conventions](SignalsAndConventions.md) defines streams, fields, metric names, correlation rules, and redaction expectations.
- [Local Development](LocalDevelopment.md) covers the supported Docker Compose workflow, bootstrap commands, and smoke validation.
- [External SFU](ExternalSFU.md) describes how to run the SFU as a standalone binary that self-ships traces, metrics, and logs.
- [Dashboards And Alerts](DashboardsAndAlerts.md) lists the repo-managed OpenObserve assets and shows the supported bootstrap flow.
- [Runbooks](Runbooks.md) gives troubleshooting steps for missing logs, traces, metrics, and SFU-specific incidents.

## Current model

- `api`, `auth`, `attachments`, `ws`, `webhook`, `indexer`, and `embedder` send traces and metrics to the local OpenTelemetry Collector, and their logs reach OpenObserve through the local collector path.
- `api`, `auth`, `attachments`, and `ws` also emit native PostgreSQL coordinator probe metrics from the same DSN they use for application traffic.
- `sfu` is deployed outside local Compose and does not require any sidecar or adjacent collector binary.
- External SFU nodes push traces, metrics, and logs over OTLP HTTP to the public telemetry gateway, which validates SFU JWTs and forwards the requests to the in-cluster collector while the SFU keeps JSON logs on stdout.

## Canonical streams

- Logs: `gochat_logs`
- Traces: `gochat_traces`
- Metrics: one stream per metric, for example `gochat_http_server_requests` or `gochat_sfu_peers_active`
