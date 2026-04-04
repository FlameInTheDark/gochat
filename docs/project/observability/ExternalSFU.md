[<- Observability](README.md)

# External SFU

The SFU is no longer part of the local Compose deployment. It is expected to run as a standalone binary on external infrastructure and push telemetry through the public telemetry gateway.

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

- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `OTEL_EXPORTER_OTLP_HEADERS`
- `OTEL_EXPORTER_OTLP_PROTOCOL`
- `OTEL_METRIC_EXPORT_INTERVAL`

Authentication:

- `WEBHOOK_TOKEN`

## YAML config equivalents

If you provision the SFU with `config.yaml`, you can set the same OTLP values there instead of exporting them separately:

- `telemetry_otlp_endpoint`
- `telemetry_otlp_headers`
- `telemetry_otlp_protocol`
- `telemetry_metric_export_interval`

## Endpoint format

- Use the shared OTLP base URL for the telemetry gateway:
  - `OTEL_EXPORTER_OTLP_ENDPOINT=https://telemetry.example.com`
- The Go OTLP HTTP exporters append:
  - traces: `/v1/traces`
  - metrics: `/v1/metrics`
  - logs: `/v1/logs`

## Authentication format

- Provision one unique `(service_id, jwt)` pair per SFU node.
- The JWT must be an HS256 token with `typ=sfu` and `id=<service_id>`.
- Reuse the same JWT for discovery heartbeat and telemetry:
  - `WEBHOOK_TOKEN=<jwt>`
  - `OTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer <jwt>`
- `SFU_SERVICE_ID` must match the JWT `id` claim.

## Example PowerShell environment

```powershell
$env:GOCHAT_DEPLOYMENT_ENV = "staging"
$env:SFU_REGION = "eu-central"
$env:SFU_SERVICE_ID = "sfu-eu-1"
$env:WEBHOOK_TOKEN = "<jwt-with-typ-sfu-id-sfu-eu-1>"

$env:OTEL_EXPORTER_OTLP_ENDPOINT = "https://telemetry.example.com"
$env:OTEL_EXPORTER_OTLP_PROTOCOL = "http/protobuf"
$env:OTEL_EXPORTER_OTLP_HEADERS = "Authorization=Bearer $($env:WEBHOOK_TOKEN)"
$env:OTEL_METRIC_EXPORT_INTERVAL = "60000"
```

Equivalent `config.yaml` snippet:

```yaml
webhook_token: "<same-jwt>"
service_id: "sfu-eu-1"
dtls_certificate_file: "/etc/gochat/sfu.crt"
dtls_private_key_file: "/etc/gochat/sfu.key"
telemetry_otlp_endpoint: "https://telemetry.example.com"
telemetry_otlp_headers: "Authorization=Bearer <same-jwt>"
telemetry_otlp_protocol: "http/protobuf"
telemetry_metric_export_interval: "60000"
```

You can generate the DTLS PEM pair with:

```bash
go run ./cmd/tools certificates dtls generate \
  --cert-out ./certs/sfu.crt \
  --key-out ./certs/sfu.key \
  --common-name sfu-eu-1
```

If `dtls_certificate_file` and `dtls_private_key_file` are left empty, the SFU generates a self-signed DTLS certificate at startup instead.

## Runtime behavior

- SFU logs always continue to stdout as JSON.
- When OTLP log export is configured, the same structured records are mirrored through the telemetry gateway asynchronously.
- The OTLP log exporter is bounded and best-effort:
  - records may be dropped if the in-memory queue is full
  - retries are attempted for transient send failures
  - exporter metrics record enqueue, success, failure, drop, latency, and last-success state

## Network requirements

An external SFU node needs outbound access to:

- the public telemetry gateway OTLP HTTP endpoint
- the internal webhook URL used for discovery heartbeat and join/leave notifications
