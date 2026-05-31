[<- Bots](README.md)

# Permissions

Bot access is calculated from normal guild membership, roles, and channel overrides, then capped by the guild install grant in `bot_guilds.granted_permissions`.

Effective bot permissions:

```text
effective = normal_member_role_and_channel_permissions & bot_guilds.granted_permissions
```

Rules:

- A bot must be installed in a guild before it can use guild runtime routes.
- A bot must be able to view a channel before it receives channel events or performs channel actions.
- Message history, send, manage messages, and add reactions use the same permission bits as regular users, with the bot grant cap applied.
- Server admins can still manage the bot member's roles.
- Changing the granted permission cap requires removing and reinstalling the bot.

Guild-scoped lifecycle events use a table-driven event-to-permission map. Channel-scoped events require current channel visibility plus any event-specific permission.
