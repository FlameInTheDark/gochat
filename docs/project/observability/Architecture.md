[<- Observability](README.md)

# Architecture

The project uses two signal paths on purpose.

- Local application services use the collector path because Docker Compose already provides a shared network and a single bootstrap surface.
- External SFU nodes use a self-shipping path because they may run as plain binaries on any host or operating system.

## Signal flow

```mermaid
flowchart LR
    clients["Clients"]
    subgraph local["Local Services"]
        api["API / Auth / WS / Attachments / Webhook / Workers"]
    end
    gateway["Telemetry Gateway"]
    collector["OpenTelemetry Collector"]
    oo["OpenObserve"]
    subgraph sfu["External SFU Node"]
        sfuapp["gochat-sfu binary"]
        stdout["JSON stdout logs"]
    end

    clients --> api
    api --> collector
    collector --> oo

    clients --> sfuapp
    sfuapp --> gateway
    gateway --> collector
    stdout --> sfuapp
```

## Local services

- HTTP services use shared request middleware for `traceparent`, `X-Request-ID`, request spans, HTTP metrics, and structured request logs.
- Async workers propagate trace context and request identity through NATS headers.
- The collector forwards traces, metrics, and container logs into OpenObserve.

## External SFU

- Traces, metrics, and logs all go from the SFU process to the public OTLP HTTP telemetry gateway.
- The telemetry gateway accepts only `/v1/traces`, `/v1/metrics`, and `/v1/logs`, validates the SFU JWT, rate-limits by `service.instance.id`, and forwards valid requests to the private collector.
- The collector stays private and exports the accepted signals to OpenObserve.
- Logs still stay on stdout locally and are also mirrored through the built-in async OTLP log exporter.
- The log exporter is best-effort by design: it uses a bounded queue, batching, retry/backoff, and dropped-log accounting so media flow is never blocked by observability.

## Identity model

These dimensions are the canonical way to identify an SFU node:

- `service.name=gochat-sfu`
- `voice.region`
- `service.instance.id`
- `deployment.environment`

They are attached to SFU spans and metrics as resource attributes and are also emitted in SFU logs.
