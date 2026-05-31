[<- Bots](README.md)

# Management API

Bot owners manage bots through the normal API service. These routes require a regular user access token.

Base path:

```text
/api/v1/developer
```

Routes:

| Method | Route | Purpose |
|--------|-------|---------|
| `POST` | `/bots` | Create a bot user and bot config |
| `GET` | `/bots` | List bots owned by the current user |
| `GET` | `/bots/public` | Search public bots |
| `GET` | `/bots/:bot_id` | Get owned bot details |
| `PATCH` | `/bots/:bot_id` | Update bot profile/configuration |
| `DELETE` | `/bots/:bot_id` | Disable/delete bot configuration |
| `POST` | `/bots/:bot_id/avatar` | Create bot avatar upload metadata |
| `POST` | `/bots/:bot_id/banner` | Create bot banner upload metadata |
| `POST` | `/bots/:bot_id/tokens` | Create a runtime token |
| `GET` | `/bots/:bot_id/tokens` | List token metadata |
| `DELETE` | `/bots/:bot_id/tokens/:token_id` | Revoke a runtime token |
| `POST` | `/bots/:bot_id/grants` | Create an install grant token |
| `GET` | `/bots/:bot_id/grants` | List install grants |
| `DELETE` | `/bots/:bot_id/grants/:grant_id` | Revoke an install grant |
| `GET` | `/bots/authorize/preview` | Preview a public bot or install grant before authorization |

Token creation responses include the plain runtime token once. Clients must store it immediately.

## Guild Installation API

Guild bot installation is handled by server-admin routes in the normal API service.

Base path:

```text
/api/v1/guild/:guild_id/bots
```

Routes:

| Method | Route | Purpose |
|--------|-------|---------|
| `GET` | `` | List installed bots in the guild |
| `POST` | `` | Install a public bot or install by grant token |
| `DELETE` | `/:bot_id` | Remove a bot from the guild |

Related route:

| Method | Route | Purpose |
|--------|-------|---------|
| `GET` | `/api/v1/guild/bots/authorize-guilds` | List guilds where the current user can authorize bots |

Only the guild owner or a member with administrator permission may install or remove bots. Bot owners cannot install their bot into another guild unless they also have the required guild permission there.

Installing a bot creates or reuses the guild member row for the bot user, then writes the `bot_guilds` record. Removing a bot deletes the install record and removes the bot member from the guild.
