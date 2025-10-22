[<- Documentation](../README.md) - [Voice](README.md)

# Regions Configuration

Voice regions are configured statically in the API configuration file (`api_config.yaml`). This avoids a database dependency for region metadata that rarely changes, and lets you define friendly display names.

## Example api_config.yaml

```yaml
voice_region: global           # default region id
voice_regions:
  - id: global
    name: Global
  - id: eu
    name: Europe (Frankfurt)
  - id: us-east
    name: US East (Ashburn)
```

Notes
- `voice_region` sets the default region id used when a channel has no explicit region assigned.
- `voice_regions` is the allowlist of valid regions. IDs must match the region identifiers used by your SFU discovery/registration (e.g., etcd).
- Friendly names (`name`) are for operator/UX use and are not currently returned by the `GET /voice/regions` endpoint.

## API Behavior
- `GET /api/v1/voice/regions` returns the configured regions with id and name:
```json
{
  "regions": [
    { "id": "global", "name": "Global" },
    { "id": "eu",     "name": "Europe (Frankfurt)" },
    { "id": "us-east", "name": "US East (Ashburn)" }
  ]
}
```
- Channel region is stored per‑channel in the database (as the region id). Admins can change it via `PATCH /guild/{guild_id}/voice/{channel_id}/region`.
- When an admin changes a channel’s region:
  1. API discovers and pre‑binds a new SFU for that region under `voice:route:{channel}`.
  2. API broadcasts an RTC rebind message via WS so clients call JoinVoice and connect to the preselected SFU.
