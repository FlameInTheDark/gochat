[<- Documentation](README.md)

# Tools CLI

The `cmd/tools` application provides helper commands for operating the platform.

## Generate Webhook Token

Generate a JWT for services that authenticate to the Webhook.

Flags
- `--type` Service type (e.g., `sfu`, `attachments`, `prom`).
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

# Prometheus HTTP SD token
# Use the output token as a Bearer token when scraping the secured SD endpoint:
#   Authorization: Bearer <token>
tools token webhook generate --type prom --secret supersecret --header --curl
```

Use the output token as `webhook_token` in `sfu_config.yaml` or as the value for `X-Webhook-Token` when calling Webhook endpoints from trusted services.

