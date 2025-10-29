[<- Documentation](../README.md) - [Voice](README.md)

# SFU Permissions (Bitmasks)

The SFU enforces a subset of platform permissions on media/control actions. Permissions are represented as a bitmask (`int64`).

Values below use the same notation as Permissions.md (`1 << n`).

| Name                    | Value     | Description |
|-------------------------|-----------|-------------|
| PermVoiceConnect        | `1 << 19` | Required to join/connect to a voice channel |
| PermVoiceSpeak          | `1 << 20` | Required to publish audio |
| PermVoiceVideo          | `1 << 21` | Required to publish video |
| PermVoiceMuteMembers    | `1 << 22` | Privileged: mute members for everyone |
| PermVoiceDeafenMembers  | `1 << 23` | Privileged: deafen members (receive no one) |
| PermVoiceMoveMembers    | `1 << 24` | Privileged: kick/move members, block/unblock joins |
| PermAdministrator       | `1 << 25` | Override: treated as allow‑all for checks |

Notes
- The `moved=true` token flag allows bypassing a room‑level block for a forced move and grants audio/video publish permissions for that session.
