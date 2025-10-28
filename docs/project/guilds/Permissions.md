[<- Documentation](../README.md) - [Guilds](README.md)

# Permissions

Permissions are represented as a single `int64` bitmask using bit shifting.

```go
type RolePermission int64
const PermServerViewChannels RolePermission = 1 << iota
```

## Permission Table

| Name                         | Value     | Description |
|------------------------------|-----------|-------------|
| View Channels                | `1 << 0`  | View channels in a guild |
| Manage Channels              | `1 << 1`  | Create, edit, delete channels |
| Manage Roles                 | `1 << 2`  | Create, edit, delete roles |
| View Audit Log               | `1 << 3`  | Read guild audit logs |
| Manage Server                | `1 << 4`  | Change guild settings |
| Create Invite                | `1 << 5`  | Create invite links |
| Change Nickname              | `1 << 6`  | Change own nickname |
| Manage Nicknames             | `1 << 7`  | Change other members' nicknames |
| Kick Members                 | `1 << 8`  | Remove members from guild |
| Ban Members                  | `1 << 9`  | Ban members from guild |
| Timeout Members              | `1 << 10` | Apply communication timeouts |
| Send Message                 | `1 << 11` | Send messages in text channels |
| Send Message in Threads      | `1 << 12` | Send messages in threads |
| Attach Files                 | `1 << 13` | Upload files/images |
| Add Reactions                | `1 << 14` | Add message reactions |
| Mention Roles                | `1 << 15` | Mention roles in messages |
| Manage Messages              | `1 << 16` | Delete others' messages, pin |
| Manage Threads               | `1 << 17` | Create/close/modify threads |
| Read Message History         | `1 << 18` | Read channel history |
| Voice: Connect               | `1 << 19` | Connect to voice channels |
| Voice: Speak                 | `1 << 20` | Publish audio in voice |
| Voice: Video                 | `1 << 21` | Publish video in voice |
| Voice: Mute Members          | `1 << 22` | Mute members (server-wide in voice) |
| Voice: Deafen Members        | `1 << 23` | Deafen members (server-wide in voice) |
| Voice: Move Members          | `1 << 24` | Move/kick members between voice channels |
| Administrator                | `1 << 25` | All permissions; overrides checks |

Default guild permissions are `7927905` but can be changed in the guild settings.

```go
var DefaultPermissions = CreatePermissions(
    PermServerViewChannels,
    PermMembershipCreateInvite,
    PermMembershipChangeNickname,
    PermTextSendMessage,
    PermTextSendMessageInThreads,
    PermTextCreateThreads,
    PermTextAddReactions,
    PermTextAttachFiles,
    PermTextReadMessageHistory,
    PermVoiceConnect,
    PermVoiceSpeak,
    PermVoiceVideo,
)
```
