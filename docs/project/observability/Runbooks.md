[<- Observability](README.md)

# Runbooks

## SFU logs missing from OpenObserve

1. Confirm stdout still shows structured JSON logs from `gochat-sfu`.
2. Confirm `OTEL_EXPORTER_OTLP_ENDPOINT` points to the telemetry gateway base URL.
3. Confirm `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`.
4. Confirm `OTEL_EXPORTER_OTLP_HEADERS` includes `Authorization=Bearer <jwt>`.
5. Query exporter health metrics:
   - `gochat_logs_exporter_send_failure`
   - `gochat_logs_exporter_dropped`
   - `gochat_logs_exporter_last_success_unix`
6. If failures rise and `last_success_unix` is stale, verify outbound connectivity to the telemetry gateway and confirm the JWT still validates.

## SFU traces missing

1. Confirm `OTEL_EXPORTER_OTLP_ENDPOINT` points to the telemetry gateway base URL.
2. Confirm `OTEL_EXPORTER_OTLP_HEADERS` includes `Authorization=Bearer <jwt>`.
3. Check whether logs still contain `trace_id`; if they do, the app is creating spans and the issue is export or auth.

## SFU metrics missing

1. Confirm `OTEL_EXPORTER_OTLP_ENDPOINT` points to the telemetry gateway base URL.
2. Confirm `OTEL_EXPORTER_OTLP_HEADERS` includes `Authorization=Bearer <jwt>`.
3. Query one SFU metric stream directly, for example `gochat_sfu_peers_active`.
4. If traces work but metrics do not, treat it as a metrics endpoint or auth problem first.

## Local stack has logs and traces but smoke still fails

1. Run `docker compose down --remove-orphans`.
2. Start again with `docker compose up -d`.
3. Re-run:
   - `go run ./cmd/tools observability bootstrap ...`
   - `go run ./cmd/tools observability smoke ...`
4. If the smoke check reports legacy observability containers, remove them before investigating anything else.
5. If the smoke check reports dashboard drift, re-run:
   - `go run ./cmd/tools observability bootstrap ...`
   - `go run ./cmd/tools observability smoke ...`

## OpenObserve shows suspiciously high event volume

1. Check whether the growth is in logs, traces, or metrics first.
2. Query the top metric streams by document count and latest timestamp.
3. Run:
   - `go run ./cmd/tools observability cleanup ...`
4. If the dry-run output is dominated by `pg_*`, `scrape_*`, `promhttp_*`, `postgres_exporter_*`, `citus_*`, `go_*`, `process_*`, `http_client_*`, or `up`, that is stale exporter-era history rather than new application spam.
5. If live growth is mostly `gochat_dependency_duration_bucket` or `gochat_http_server_duration_bucket`, treat it as histogram bucket volume and verify that the services were rebuilt after the metric-view rollout.
6. Only run:
   - `go run ./cmd/tools observability cleanup --delete-legacy-streams ...`
   once you confirm the matched streams are truly legacy data.

## Postgres probe failing

1. Query `gochat_postgres_probe_status` and group by `service_name`.
2. Check `gochat_postgres_probe_failure` for recent increments.
3. Check `gochat_postgres_probe_duration` for latency growth before the failures started.
4. Use `gochat_postgres_probe_last_success_unix` to see which services are stale versus fully down.
5. If only one service is failing, treat it as a service-local connectivity or DSN problem before assuming coordinator-wide downtime.

## Heartbeat alert is firing

1. Query `gochat_logs` for `service_name='gochat-sfu'`.
2. Group by `voice_region` and `service_instance_id`.
3. Check whether failures are auth-related, DNS-related, or upstream webhook failures.
4. Check `gochat_sfu_heartbeat_duration` for latency regressions before the failures started.
5. If one node is isolated, drain or restart only that node rather than assuming a fleet-wide issue.
