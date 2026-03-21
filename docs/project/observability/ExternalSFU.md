[<- Observability](README.md)

# External SFU

The SFU is no longer part of the local Compose deployment. It is expected to run as a standalone binary on external infrastructure and self-ship telemetry.

## Design goals

- no sidecar
- no host agent
- no Docker or Kubernetes requirement
- no observability dependency that can block signaling or media

## Required environment

Identity:

- `SFU_REGION`
- `SFU_SERVICE_ID`
- `GOCHAT_DEPLOYMENT_ENV`

Tracing:

- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`
- `OTEL_EXPORTER_OTLP_TRACES_HEADERS`

Metrics:

- `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`
- `OTEL_EXPORTER_OTLP_METRICS_HEADERS`

Direct logs:

- `OPENOBSERVE_LOGS_ENABLED`
- `OPENOBSERVE_LOGS_ENDPOINT`
- `OPENOBSERVE_LOGS_AUTH`
- `OPENOBSERVE_LOGS_STREAM`

## Endpoint format

Use signal-specific OTLP HTTP endpoints for the SFU:

- traces: `https://observe.example.com/api/default/v1/traces`
- metrics: `https://observe.example.com/api/default/v1/metrics`

For direct logs, use the OpenObserve org base URL and let the SFU append the stream path:

- `OPENOBSERVE_LOGS_ENDPOINT=https://observe.example.com/api/default`
- `OPENOBSERVE_LOGS_STREAM=gochat_logs`

The log exporter sends to:

- `https://observe.example.com/api/default/gochat_logs/_json`

## Authentication format

- `OTEL_EXPORTER_OTLP_*_HEADERS` should include the raw OTLP HTTP headers, for example `Authorization=Basic <base64-user-pass>`
- `OPENOBSERVE_LOGS_AUTH` should be the full HTTP `Authorization` header value, for example `Basic <base64-user-pass>`

## Example PowerShell environment

```powershell
$env:GOCHAT_DEPLOYMENT_ENV = "staging"
$env:SFU_REGION = "eu-central"
$env:SFU_SERVICE_ID = "sfu-eu-1"

$env:OTEL_EXPORTER_OTLP_TRACES_ENDPOINT = "https://observe.example.com/api/default/v1/traces"
$env:OTEL_EXPORTER_OTLP_TRACES_HEADERS = "Authorization=Basic <base64-user-pass>"
$env:OTEL_EXPORTER_OTLP_METRICS_ENDPOINT = "https://observe.example.com/api/default/v1/metrics"
$env:OTEL_EXPORTER_OTLP_METRICS_HEADERS = "Authorization=Basic <base64-user-pass>"

$env:OPENOBSERVE_LOGS_ENABLED = "true"
$env:OPENOBSERVE_LOGS_ENDPOINT = "https://observe.example.com/api/default"
$env:OPENOBSERVE_LOGS_AUTH = "Basic <base64-user-pass>"
$env:OPENOBSERVE_LOGS_STREAM = "gochat_logs"
```

## Runtime behavior

- SFU logs always continue to stdout as JSON.
- When direct log shipping is enabled, the same structured records are mirrored to OpenObserve asynchronously.
- The direct log exporter is bounded and best-effort:
  - records may be dropped if the in-memory queue is full
  - retries are attempted for transient send failures
  - exporter metrics record enqueue, success, failure, drop, latency, and last-success state

## Network requirements

An external SFU node needs outbound access to:

- the OpenObserve OTLP HTTP traces endpoint
- the OpenObserve OTLP HTTP metrics endpoint
- the OpenObserve JSON log ingestion endpoint
- the internal webhook URL used for discovery heartbeat and join/leave notifications
