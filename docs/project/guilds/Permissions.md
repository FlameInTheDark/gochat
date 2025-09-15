[<- Documentation](../README.md) - [Guilds](README.md)

# Permissions

Permissions in the application are simply one `int64` value with bit shifting.

```go
type RolePermission int64
const PermServerViewChannels RolePermission = 1 << iota
```

### Server permissions
| Permission             | Value    |
|------------------------|----------|
| View Channels          | `1 << 0` |
| Manage Channels        | `1 << 1` |
| Manage Roles           | `1 << 2` |
| View Audit Log         | `1 << 3` |
| Manage Server          | `1 << 4` |

### Membership permissions
| Permission             | Value     |
|------------------------|-----------|
| Create Invite          | `1 << 5`  |
| Change Nickname        | `1 << 6`  |
| Manage Nicknames       | `1 << 7`  |
| Kick Members           | `1 << 8`  |
| Ban Members            | `1 << 9`  |
| Timeout Members        | `1 << 10` |

### Text permissions
| Permission             | Value     |
|------------------------|-----------|
| Send Message           | `1 << 11` |
| Send Message in Thread | `1 << 12` |
| Attach Files           | `1 << 13` |
| Add Reactions          | `1 << 14` |
| Mention Roles          | `1 << 15` |
| Manage Messages        | `1 << 16` |
| Manage Threads         | `1 << 17` |
| Read Message History   | `1 << 18` |

### Voice permissions
| Permission     | Value     |
|----------------|-----------|
| Connect        | `1 << 19` |
| Speak          | `1 << 20` |
| Video          | `1 << 21` |
| Mute Members   | `1 << 22` |
| Deafen Members | `1 << 23` |
| Move Members   | `1 << 24` |

### Administrative permissions
| Permission    | Value     |
|---------------|-----------|
| Administrator | `1 << 25` |

This role combines in a single `int64` value.

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
	PermVoiceVideo)
```