# Gateway Events

Gateway events are sent as opcode `0` dispatch envelopes:

```json
{
  "op": 0,
  "t": 100,
  "d": {}
}
```

`t` is an integer event type. `d` is the event-specific payload.

The current bot router delivers:

- Guild-scoped events for guilds where the bot is installed and permitted.
- Channel-scoped events for channels where the bot has visibility and any event-specific permission.
- Direct-message events only for `USER_DM_MESSAGE`.

Other shared protocol event types exist in the codebase and are documented here for completeness, but they are not all routed to bot gateway sessions today.

## Event Type Summary

### Gateway

| Type | Name | Delivered to bots | Payload |
|------|------|-------------------|---------|
| `1` | `GATEWAY_READY` | Yes, after identify | [Ready](payloads.md#ready) |

### Message And Channel Events

| Type | Name | Delivered to bots | Payload |
|------|------|-------------------|---------|
| `100` | `MESSAGE_CREATE` | Yes | [MessageCreate](payloads.md#messagecreate) |
| `101` | `MESSAGE_UPDATE` | Yes | [MessageUpdate](payloads.md#messageupdate) |
| `102` | `MESSAGE_DELETE` | Yes | [MessageDelete](payloads.md#messagedelete) |
| `103` | `MESSAGE_REACTION_ADD` | Yes | [MessageReactionAdd](payloads.md#messagereactionadd) |
| `104` | `MESSAGE_REACTION_REMOVE` | Yes | [MessageReactionRemove](payloads.md#messagereactionremove) |
| `105` | `GUILD_CREATE` | Protocol-defined | [GuildCreate](payloads.md#guildcreate) |
| `106` | `GUILD_UPDATE` | Yes, when published | [GuildUpdate](payloads.md#guildupdate) |
| `107` | `GUILD_DELETE` | Protocol-defined | [GuildDelete](payloads.md#guilddelete) |
| `108` | `CHANNEL_CREATE` | Yes | [ChannelCreate](payloads.md#channelcreate) |
| `109` | `CHANNEL_UPDATE` | Yes | [ChannelUpdate](payloads.md#channelupdate) |
| `110` | `CHANNEL_ORDER_UPDATE` | Yes | [ChannelOrderUpdate](payloads.md#channelorderupdate) |
| `111` | `CHANNEL_DELETE` | Yes | [ChannelDelete](payloads.md#channeldelete) |
| `112` | `GUILD_ROLE_CREATE` | Yes | [GuildRoleCreate](payloads.md#guildrolecreate) |
| `113` | `GUILD_ROLE_UPDATE` | Yes | [GuildRoleUpdate](payloads.md#guildroleupdate) |
| `114` | `GUILD_ROLE_DELETE` | Yes | [GuildRoleDelete](payloads.md#guildroledelete) |
| `115` | `THREAD_CREATE` | Yes | [ThreadCreate](payloads.md#threadcreate) |
| `116` | `THREAD_UPDATE` | Yes | [ThreadUpdate](payloads.md#threadupdate) |
| `117` | `THREAD_DELETE` | Yes | [ThreadDelete](payloads.md#threaddelete) |
| `118` | `GUILD_EMOJI_CREATE` | Yes | [GuildEmojiCreate](payloads.md#guildemojicreate) |
| `119` | `GUILD_EMOJI_UPDATE` | Yes | [GuildEmojiUpdate](payloads.md#guildemojiupdate) |
| `120` | `GUILD_EMOJI_DELETE` | Yes | [GuildEmojiDelete](payloads.md#guildemojidelete) |

### Guild Member And Voice Events

| Type | Name | Delivered to bots | Payload |
|------|------|-------------------|---------|
| `200` | `GUILD_MEMBER_ADD` | Yes | [GuildMemberAdd](payloads.md#guildmemberadd) |
| `201` | `GUILD_MEMBER_UPDATE` | Yes | [GuildMemberUpdate](payloads.md#guildmemberupdate) |
| `202` | `GUILD_MEMBER_REMOVE` | Yes | [GuildMemberRemove](payloads.md#guildmemberremove) |
| `203` | `GUILD_MEMBER_ADD_ROLE` | Yes | [GuildMemberAddRole](payloads.md#guildmemberaddrole) |
| `204` | `GUILD_MEMBER_REMOVE_ROLE` | Yes | [GuildMemberRemoveRole](payloads.md#guildmemberremoverole) |
| `205` | `GUILD_MEMBER_JOIN_VOICE` | Yes | [GuildMemberJoinVoice](payloads.md#guildmemberjoinvoice) |
| `206` | `GUILD_MEMBER_LEAVE_VOICE` | Yes | [GuildMemberLeaveVoice](payloads.md#guildmemberleavevoice) |
| `207` | `GUILD_MEMBER_MODERATION` | Yes | [GuildMemberModeration](payloads.md#guildmembermoderation) |
| `208` | `GUILD_VOICE_REGION_CHANGING` | Yes | [VoiceRegionChanging](payloads.md#voiceregionchanging) |
| `209` | `VOICE_STATE_UPDATE` | Yes | [VoiceStateUpdate](payloads.md#voicestateupdate) |
| `210` | `GUILD_MEMBER_START_STREAM` | Yes | [GuildMemberStartStream](payloads.md#guildmemberstartstream) |
| `211` | `GUILD_MEMBER_STOP_STREAM` | Yes | [GuildMemberStopStream](payloads.md#guildmemberstopstream) |
| `212` | `GUILD_STREAMS_REBIND` | Yes | [GuildStreamsRebind](payloads.md#guildstreamsrebind) |

### Compact Notification Events

| Type | Name | Delivered to bots | Payload |
|------|------|-------------------|---------|
| `300` | `GUILD_CHANNEL_MESSAGE` | Yes, when published | [GuildChannelMessage](payloads.md#guildchannelmessage) |
| `301` | `CHANNEL_USER_TYPING` | Yes | [ChannelUserTyping](payloads.md#channelusertyping) |
| `302` | `MENTION` | Yes, when published | [Mention](payloads.md#mention) |

### User And DM Events

| Type | Name | Delivered to bots | Payload |
|------|------|-------------------|---------|
| `400` | `USER_UPDATE_READ_STATE` | Shared protocol only for bot gateway today | [UpdateReadState](payloads.md#updatereadstate) |
| `401` | `USER_UPDATE_SETTINGS` | Shared protocol only for bot gateway today | [UpdateUserSettings](payloads.md#updateusersettings) |
| `402` | `USER_FRIEND_REQUEST` | Shared protocol only for bot gateway today | [IncomingFriendRequest](payloads.md#incomingfriendrequest) |
| `403` | `USER_FRIEND_ADDED` | Shared protocol only for bot gateway today | [FriendAdded](payloads.md#friendadded) |
| `404` | `USER_FRIEND_REMOVED` | Shared protocol only for bot gateway today | [FriendRemoved](payloads.md#friendremoved) |
| `405` | `USER_DM_MESSAGE` | Yes, for bot DMs | [DMMessage](payloads.md#dmmessage) |
| `406` | `USER_UPDATE` | Shared protocol only for bot gateway today | [UpdateUser](payloads.md#updateuser) |
| `407` | `USER_AUTH_REVOKED` | Shared protocol only for bot gateway today | [UserAuthRevoked](payloads.md#userauthrevoked) |
| `408` | `USER_DM_CALL_STARTED` | Shared protocol only for bot gateway today | [DMCallStarted](payloads.md#dmcallstarted) |
| `409` | `USER_DM_CALL_JOINED` | Shared protocol only for bot gateway today | [DMCallJoined](payloads.md#dmcalljoined) |
| `410` | `USER_DM_CALL_DECLINED` | Shared protocol only for bot gateway today | [DMCallDeclined](payloads.md#dmcalldeclined) |
| `411` | `USER_DM_CALL_LEFT` | Shared protocol only for bot gateway today | [DMCallLeft](payloads.md#dmcallleft) |
| `412` | `USER_DM_CALL_ENDED` | Shared protocol only for bot gateway today | [DMCallEnded](payloads.md#dmcallended) |
| `413` | `USER_DM_CALL_STREAM_STARTED` | Shared protocol only for bot gateway today | [DMCallStreamStarted](payloads.md#dmcallstreamstarted) |
| `414` | `USER_DM_CALL_STREAM_STOPPED` | Shared protocol only for bot gateway today | [DMCallStreamStopped](payloads.md#dmcallstreamstopped) |

### RTC Shared Event Types

The following event types are defined by the shared realtime protocol. The bot gateway server does not currently process client RTC opcodes or route these events as bot events.

| Type | Name | Payload |
|------|------|---------|
| `500` | `RTC_JOIN` | Shared RTC payload |
| `501` | `RTC_OFFER` | Shared RTC payload |
| `502` | `RTC_ANSWER` | Shared RTC payload |
| `503` | `RTC_CANDIDATE` | Shared RTC payload |
| `504` | `RTC_LEAVE` | Shared RTC payload |
| `505` | `RTC_MUTE_SELF` | Shared RTC payload |
| `506` | `RTC_MUTE_USER` | Shared RTC payload |
| `507` | `RTC_SERVER_MUTE_USER` | Shared RTC payload |
| `508` | `RTC_SERVER_DEAFEN_USER` | Shared RTC payload |
| `509` | `RTC_BINDING_ALIVE` | Shared RTC payload |
| `510` | `RTC_SERVER_KICK_USER` | Shared RTC payload |
| `511` | `RTC_SERVER_BLOCK_USER` | Shared RTC payload |
| `512` | `RTC_MOVED` | [VoiceMove](payloads.md#voicemove) |
| `513` | `RTC_SERVER_REBIND` | [VoiceRebind](payloads.md#voicerebind) |
| `514` | `RTC_SPEAKING` | Shared RTC payload |
| `530` | `RTC_IDENTIFY` | RTC v2 payload |
| `531` | `RTC_READY` | RTC v2 payload |
| `532` | `RTC_SELECT_PROTOCOL` | RTC v2 payload |
| `533` | `RTC_SESSION_DESCRIPTION` | RTC v2 payload |
| `539` | `RTC_ERROR` | RTC error payload |

## Examples

### Message Create

```json
{
  "op": 0,
  "t": 100,
  "d": {
    "guild_id": 2230469276416868352,
    "message": {
      "id": 2230469276416868354,
      "channel_id": 2230469276416868353,
      "author": {
        "id": 42,
        "name": "alice",
        "discriminator": "0001",
        "is_bot": false
      },
      "content": "ping",
      "position": 10,
      "type": 0
    }
  }
}
```

### Typing

```json
{
  "op": 0,
  "t": 301,
  "d": {
    "guild_id": 2230469276416868352,
    "channel_id": 2230469276416868353,
    "user_id": 42
  }
}
```

### Direct Message

```json
{
  "op": 0,
  "t": 405,
  "d": {
    "channel_id": 2230469276416868353,
    "message_id": 2230469276416868354,
    "from": {
      "id": 42,
      "name": "alice",
      "discriminator": "0001"
    },
    "message": {
      "id": 2230469276416868354,
      "channel_id": 2230469276416868353,
      "author": {
        "id": 42,
        "name": "alice",
        "discriminator": "0001",
        "is_bot": false
      },
      "content": "hello bot",
      "type": 0
    }
  }
}
```

## Event Delivery Permissions

Guild-scoped events require the bot to be installed in the guild and to pass the guild permission check.

Channel-scoped message and reaction events require:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

Other channel-scoped events require at least channel visibility:

```text
PermServerViewChannels
```

DM events are delivered only to sessions for the authenticated bot user. In sharded bots, only shard `0` receives DM events.

## Unknown Events

Clients should preserve unknown dispatches. New event types may be added without changing the gateway envelope. The Go client exposes raw `Event` handlers for this reason.
