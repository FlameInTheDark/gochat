[<- Documentation](../README.md) - [Guilds](README.md)

# Roles and Permissions System

GoChat uses a robust Role-Based Access Control (RBAC) system combined with bitmask permissions to determine what actions a user can perform within a Guild (server) and its Channels.

This document details how Roles, Permissions, and Channel Overrides interact.

---

## 1. Permissions Architecture

Permissions in GoChat are represented as a single `int64` bitmask, where each bit corresponds to a specific capability. This allows for fast privilege checking using bitwise operations (`&`, `|`, `^`).

### Permission Table

The following permissions are defined in `internal/permissions/permissions.go`:

| Name                           | Value     | Description |
|--------------------------------|-----------|-------------|
| **View Channels**              | `1 << 0`  | View channels in a guild |
| **Manage Channels**            | `1 << 1`  | Create, edit, delete channels |
| **Manage Roles**               | `1 << 2`  | Create, edit, delete roles |
| **View Audit Log**             | `1 << 3`  | Read guild audit logs |
| **Manage Server**              | `1 << 4`  | Change guild settings |
| **Create Invite**              | `1 << 5`  | Create invite links |
| **Change Nickname**            | `1 << 6`  | Change own nickname |
| **Manage Nicknames**           | `1 << 7`  | Change other members' nicknames |
| **Kick Members**               | `1 << 8`  | Remove members from guild |
| **Ban Members**                | `1 << 9`  | Ban members from guild |
| **Timeout Members**            | `1 << 10` | Apply communication timeouts |
| **Send Message**               | `1 << 11` | Send messages in text channels |
| **Send Message in Threads**    | `1 << 12` | Send messages in threads |
| **Create Threads**             | `1 << 13` | Create new threads |
| **Attach Files**               | `1 << 14` | Upload files/images |
| **Add Reactions**              | `1 << 15` | Add message reactions |
| **Mention Roles**              | `1 << 16` | Mention roles in messages |
| **Manage Messages**            | `1 << 17` | Delete others' messages, pin messages |
| **Manage Threads**             | `1 << 18` | Modify or close threads |
| **Read Message History**       | `1 << 19` | Read channel history |
| **Voice: Connect**             | `1 << 20` | Connect to voice channels |
| **Voice: Speak**               | `1 << 21` | Publish audio in voice |
| **Voice: Video**               | `1 << 22` | Publish video in voice |
| **Voice: Mute Members**        | `1 << 23` | Privileged: Mute members (server-wide) |
| **Voice: Deafen Members**      | `1 << 24` | Privileged: Deafen members (server-wide) |
| **Voice: Move Members**        | `1 << 25` | Privileged: Move/kick members |
| **Administrator**              | `1 << 26` | Has all permissions and bypasses overrides |

> **Note:** The Administrator permission (`1 << 26` or whatever the highest bit is set to) acts as a catch-all override. Any user with a role possessing this permission will automatically pass any permission check.

### Default Permissions
When a guild is created or a default role (like `@everyone`) is initialized, it receives a standard set of permissions. The default bitmask is typically `7927905`, comprising:
View Channels, Create Invite, Change Nickname, Send Message, Send Message in Threads, Create Threads, Add Reactions, Attach Files, Read Message History, Voice Connect, Speak, and Video.

---

## 2. Roles System

A **Role** is a cohesive set of permissions assigned to members of a guild.

### Role Entity (`internal/database/model/role.go`)
- `id`: Unique identifier for the role.
- `guild_id`: The guild this role belongs to.
- `name`: Display name of the role (e.g., "Moderator").
- `color`: Integer representation of the role's display color.
- `permissions`: The `int64` bitmask of server-wide permissions granted by this role.

### User Role Assignment (`internal/database/model/user_role.go`)
Users can have multiple roles in a guild. A user's base server permissions are calculated by performing a **bitwise OR** (`|`) on the permissions of all their assigned roles.
`BasePermissions = RoleA.Permissions | RoleB.Permissions | ...`

---

## 3. Channel Overrides

For finer control, permissions can be overridden on a per-channel basis.

### Override Entity (`internal/database/model/channel_roles_perm.go`)
- `channel_id`: The channel the override applies to.
- `role_id`: The role the override targets.
- `accept`: Bitmask of permissions explicitly *granted* in this channel.
- `deny`: Bitmask of permissions explicitly *revoked* in this channel.

### Permission Resolution Logic
To calculate a user's effective permissions for a specific channel, the system follows this hierarchy:
1. **Base Permissions:** Accumulate permissions from all of the user's roles.
2. **Admin Check:** If the user has `PermAdministrator`, they instantly receive all permissions (resolution stops).
3. **Deny Overrides:** Subtract the permissions defined in the `deny` bitmask of the channel overrides for the user's roles. `(Permissions &^ DenyMask)`
4. **Accept Overrides:** Add the permissions defined in the `accept` bitmask of the channel overrides for the user's roles. `(Permissions | AcceptMask)`

*Note: Channel overrides tied to specific user IDs (rather than roles) might also exist or be planned, following the same Deny -> Accept flow.*

---

## 4. Voice Server (SFU) Permissions

The underlying Selective Forwarding Unit (SFU) handling Voice/Video connections strictly enforces voice-related subset mappings:
- `PermVoiceConnect`, `PermVoiceSpeak`, `PermVoiceVideo` for media publishing/subscribing.
- Privileged controls: `PermVoiceMuteMembers`, `PermVoiceDeafenMembers`, `PermVoiceMoveMembers`.

**Special Case:** When a user is force-moved by a moderator (e.g., dragged to an empty room), the control server tokens the payload with a `moved=true` flag. This flag instructs the SFU to temporarily bypass room-level connection blocks, granting the moved user basic audio/video rights for that session. See [SFUPermissions.md](../voice/SFUPermissions.md) for deeper voice interactions.

---

## 5. Tooling

GoChat provides a quick CLI utility to decode permission bitmasks, which is highly useful when debugging database records or API payloads.

**Usage:**
```bash
go run cmd/tools/permissions.go decode -v 7927905
```

This will print out the human-readable names of every permission embedded in that integer state, or generate JSON if passed the `-f json` flag.
