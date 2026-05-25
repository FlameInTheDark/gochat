[<- Documentation](README.md)

# Tools CLI

The `cmd/tools` application provides helper commands for operating the platform.

## Generate DTLS Certificate

Generate a DTLS certificate/key pair for SFU WebRTC transport and write PEM files you can reference from `sfu_config.yaml`.

Flags
- `--cert-out` Output path for the certificate PEM file.
- `--key-out` Output path for the private key PEM file.
- `--common-name` Optional certificate common name. Defaults to `gochat-sfu`.
- `--valid-for` Optional certificate validity duration. Defaults to `8760h` (one year).
- `--overwrite` Replace existing output files.
- `--format` Output format: `text` (default) or `json`.

Examples
```
go run ./cmd/tools certificates dtls generate \
  --cert-out ./certs/sfu.crt \
  --key-out ./certs/sfu.key

go run ./cmd/tools certificates dtls generate \
  --cert-out ./certs/sfu.crt \
  --key-out ./certs/sfu.key \
  --common-name sfu-eu-1 \
  --valid-for 2160h \
  --format json
```

## Generate Webhook Token

Generate a JWT for services that authenticate to the Webhook.

Flags
- `--type` Service type (e.g., `sfu`, `attachments`).
- `--id` Optional service id (UUIDv4). When omitted, a random UUID is generated. For SFU, this should match `service_id` in `sfu_config.yaml`.
- `--secret` HS256 secret used by the Webhook service (`jwt_secret`).
- `--format` Output format: `text` (default) or `json`.
- `--header` Print the `X-Webhook-Token` header line.
- `--curl` Print a ready-to-run cURL example for the selected type.

Examples
```
# Generate SFU token with a fixed id
tools token webhook generate --type sfu --id 26a58109-fbc4-4205-ad3e-8bef10e9d8d5 --secret supersecret

# Print as header and curl example
tools token webhook generate --type sfu --secret supersecret --header --curl

# JSON output (contains id and token fields)
tools token webhook generate --type attachments --secret supersecret --format json
```

Use the output token as `webhook_token` in `sfu_config.yaml` or as the value for `X-Webhook-Token` when calling Webhook endpoints from trusted services.

## Observability Commands

Use the observability subcommands in `cmd/tools` as the supported operator entrypoints for the local stack.

Bootstrap OpenObserve dashboards and alerts:

```
go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
```

Run the local smoke check:

```
go run ./cmd/tools observability smoke --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123
```

The local Postgres exporter path is still a temporary internal bridge behind the collector. It is not a user-facing monitoring workflow.

## YugabyteDB Verification

Compare two YugabyteDB YSQL databases by table row counts and checksums:

```
go run ./cmd/tools yugabyte verify \
  --source-dsn "postgres://yugabyte:yugabyte@127.0.0.1:5433/gochat?sslmode=disable" \
  --target-dsn "postgres://yugabyte:yugabyte@127.0.0.1:5433/gochat_copy?sslmode=disable"
```

For a direct run outside Docker, provide DSNs that are reachable from the current shell:

```
go run ./cmd/tools yugabyte verify \
  --source-dsn "postgres://yugabyte:yugabyte@127.0.0.1:5433/gochat?sslmode=disable" \
  --target-dsn "postgres://yugabyte:yugabyte@127.0.0.1:5433/gochat_copy?sslmode=disable"
```

The command returns a non-zero exit code if a table is missing, a row count differs, or a checksum differs. Use `--skip-checksum` only when the checksum query is too expensive and a separate validation method is recorded.

