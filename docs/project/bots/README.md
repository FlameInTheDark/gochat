[<- Documentation](../README.md)

# Bots

Bots are first-class flagged user accounts with owner-managed configuration, token-based runtime authentication, guild installation, public discovery, a dedicated runtime REST API, and a dedicated WebSocket gateway.

## Pages

- [Overview](Overview.md)
- [Management API](ManagementAPI.md)
- [Runtime REST API](RuntimeAPI.md)
- [Gateway](Gateway.md)
- [Event Routing](EventRouting.md)
- [Permissions](Permissions.md)
- [Presence](Presence.md)
- [Operations](Operations.md)

## Surfaces

| Surface | Service | Base path | Auth | Purpose |
|---------|---------|-----------|------|---------|
| Bot management | `cmd/api` | `/api/v1/developer/bots` | User access token | Create bots, edit profile/configuration, issue runtime tokens, create install grants |
| Guild installation | `cmd/api` | `/api/v1/guild/:guild_id/bots` | User access token | Server admins list, install, and remove bots in a guild |
| Bot runtime REST | `cmd/botapi` | `/bot/api/v1` | `Authorization: Bot <token>` | Bot account, guild, channel, message, typing, reaction, and read-state actions |
| Bot gateway | `cmd/botws` | `/bot/ws` | `Authorization: Bot <token>` | Dedicated bot WebSocket event stream and bot presence updates |
| Bot event router | `cmd/botrouter` | NATS worker | Internal | Partitioned fan-out from bot events to active bot gateway sessions |

Bot runtime traffic is intentionally separated from normal user API and user WebSocket workload. Runtime HTTP handlers live under `cmd/botapi/endpoints`, gateway logic lives under `cmd/botws`, and bot event fan-out lives under `cmd/botrouter`.
