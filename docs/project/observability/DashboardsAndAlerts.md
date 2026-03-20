[<- Observability](README.md)

# Dashboards And Alerts

## Bootstrap flow

Dashboards and alert seeds live under `monitoring/openobserve/`.

Bootstrap dashboards:

```powershell
go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
```

The bootstrap command now prints a per-dashboard `created`, `updated`, or `in-sync` result and verifies that the live payload matches the repo-managed JSON after upsert.

Bootstrap alerts when a destination already exists:

```powershell
go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123 --alert-destination-id <destination-name>
```

Review or delete stale exporter-era metric streams:

```powershell
go run ./cmd/tools observability cleanup --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
go run ./cmd/tools observability cleanup --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123 --delete-legacy-streams
```

## Dashboard inventory

- `Edge/API`
- `Realtime/WS`
- `Voice/SFU`
- `Async Workers`
- `Data Stores`

The `Voice/SFU` dashboard is intended for externally deployed nodes and groups metrics by `voice_region` and `service_instance_id`.

Each dashboard is intentionally multi-tab now:

- `Edge/API`: `Overview`, `Latency By Route`, `Errors And Auth`, `Rate Limit And Idempotency`
- `Realtime/WS`: `Connections`, `Auth And Heartbeats`, `Delivery Reliability`, `Message Mix`
- `Async Workers`: `Throughput`, `Latency`, `Failures And Decodes`, `Recent Success`
- `Data Stores`: `Overview`, `Redis And Cache`, `Postgres Probes And Pools`, `Search S3 Etcd`
- `Voice/SFU`: `Topology`, `Signaling Traffic`, `Heartbeats And Admin Close`, `Peer Lifecycle`

## Alert inventory

- `HTTP 5xx spike`
- `HTTP latency p95 regression`
- `Auth failure burst`
- `HTTP rate-limit burst`
- `HTTP idempotency anomaly`
- `Indexer stalled`
- `Embedder stalled`
- `Worker decode failures`
- `WS auth failure burst`
- `WS heartbeat timeout burst`
- `WS dropped delivery burst`
- `Redis/cache dependency error burst`
- `OpenSearch latency regression`
- `S3 dependency failure burst`
- `SFU heartbeat failure logs`
- `Postgres probe failing`
- `Postgres probe latency regression`
- `SFU bitrate disconnect burst`
- `SFU admin-close latency regression`
- `Error log burst`

## Useful queries

SFU heartbeat failures by node:

```sql
SELECT
  coalesce(voice_region, 'unknown') AS voice_region,
  coalesce(service_instance_id, 'unknown') AS service_instance_id,
  count(*) AS failures
FROM "gochat_logs"
WHERE service_name = 'gochat-sfu'
  AND body LIKE '%heartbeat request failed%'
GROUP BY voice_region, service_instance_id
ORDER BY failures DESC
```

Recent traces for one SFU node:

```sql
SELECT *
FROM "gochat_traces"
WHERE service_name = 'gochat-sfu'
  AND service_instance_id = 'sfu-eu-1'
ORDER BY _timestamp DESC
LIMIT 50
```

Active peers by region and instance:

```promql
sum by (voice_region, service_instance_id) (gochat_sfu_peers_active)
```

Postgres probe status by service:

```promql
min by (service_name) (last_over_time(gochat_postgres_probe_status[5m]))
```
