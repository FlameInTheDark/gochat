[<- Documentation](README.md)

# Guild Moderation

Guild member moderation adds kick, ban, unban, and ban listing routes at guild scope.

## Permissions

A request is allowed when the acting user is one of:
- the guild owner
- a member with `Administrator`
- a member with `PermMembershipKickMembers` for kick
- a member with `PermMembershipBanMembers` for ban, unban, and ban listing

Additional hierarchy rules:
- the guild owner cannot be kicked or banned
- members with `Administrator` can only be kicked, banned, or unbanned by the guild owner

## Routes

- `POST /guild/{guild_id}/member/{user_id}/kick`
- `POST /guild/{guild_id}/member/{user_id}/ban`
- `DELETE /guild/{guild_id}/member/{user_id}/ban`
- `GET /guild/{guild_id}/bans`

### Ban body

```json
{
  "reason": "Spam links"
}
```

- `reason` is optional
- maximum length is 256 Unicode characters

## Message visibility for banned authors

Banning a user does not modify stored messages in Cassandra. Instead, guild message history responses redact banned authors at API read time:
- `content` is returned as an empty string
- `attachments` are omitted
- `embeds` are omitted
- `flags` includes the banned-author flag so clients can render a placeholder state

This keeps original message data intact so unbanning a user restores normal reads without any data migration.

## WebSocket updates

Guild subscribers receive a dedicated moderation dispatch event (`t=207`) with:
- `guild_id`
- `user_id`
- `actor_id`
- `action` as `kick`, `ban`, or `unban`
- optional `reason` for ban actions

Kick and ban also continue to emit the normal guild member removal event because membership changed.
