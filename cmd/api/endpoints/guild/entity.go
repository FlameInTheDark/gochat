package guild

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/indexmq"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/entities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/invite"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

const entityName = "guild"

func (e *entity) Init(router fiber.Router) {
	router.Post("", e.Create)
	router.Get("/:guild_id<int>", e.Get)
	router.Patch("/:guild_id<int>", e.Update)
	router.Delete("/:guild_id<int>", e.Delete)
	router.Post("/:guild_id<int>", e.SetSystemMessagesChannel)

	// Icons
	router.Post("/:guild_id<int>/icon", e.CreateIcon)
	router.Get("/:guild_id<int>/icons", e.ListIcons)
	router.Delete("/:guild_id<int>/icons/:icon_id<int>", e.DeleteIcon)

	// Channels
	router.Get("/:guild_id<int>/channel/:channel_id<int>", e.GetChannel)
	router.Get("/:guild_id<int>/channel", e.GetChannels)
	router.Post("/:guild_id<int>/channel", e.CreateChannel)
	router.Patch("/:guild_id<int>/channel/order", e.PatchChannelOrder)
	router.Patch("/:guild_id<int>/channel/:channel_id<int>", e.PatchChannel)
	router.Post("/:guild_id<int>/category", e.CreateCategory)
	router.Delete("/:guild_id<int>/channel/:channel_id<int>", e.DeleteChannel)
	router.Delete("/:guild_id<int>/category/:category_id<int>", e.DeleteCategory)

	// Voice
	router.Post("/:guild_id<int>/voice/:channel_id<int>/join", e.JoinVoice)
	router.Patch("/:guild_id<int>/voice/:channel_id<int>/region", e.SetVoiceRegion)
	router.Post("/:guild_id<int>/voice/move", e.MoveMember)

	// Members
	router.Get("/:guild_id<int>/members", e.GetMembers)

	// Roles
	router.Get("/:guild_id<int>/roles", e.GetGuildRoles)
	router.Post("/:guild_id<int>/roles", e.CreateGuildRole)
	router.Patch("/:guild_id<int>/roles/:role_id<int>", e.PatchGuildRole)
	router.Delete("/:guild_id<int>/roles/:role_id<int>", e.DeleteGuildRole)
	router.Get("/:guild_id<int>/member/:user_id<int>/roles", e.GetMemberRoles)
	router.Put("/:guild_id<int>/member/:user_id<int>/roles/:role_id<int>", e.AddMemberRole)
	router.Delete("/:guild_id<int>/member/:user_id<int>/roles/:role_id<int>", e.RemoveMemberRole)
	router.Get("/:guild_id<int>/channel/:channel_id<int>/roles", e.GetChannelRolePermissions)
	router.Get("/:guild_id<int>/channel/:channel_id<int>/roles/:role_id<int>", e.GetChannelRolePermission)
	router.Put("/:guild_id<int>/channel/:channel_id<int>/roles/:role_id<int>", e.SetChannelRolePermission)
	router.Patch("/:guild_id<int>/channel/:channel_id<int>/roles/:role_id<int>", e.UpdateChannelRolePermission)
	router.Delete("/:guild_id<int>/channel/:channel_id<int>/roles/:role_id<int>", e.RemoveChannelRolePermission)

	// Invites
	router.Get("/invites/receive/:invite_code", e.ReceiveInvite)
	router.Post("/invites/accept/:invite_code", e.AcceptInvite)
	router.Get("/invites/:guild_id<int>", e.ListInvites)
	router.Delete("/invites/:guild_id<int>/:invite_id<int>", e.DeleteInvite)
	router.Post("/invites/:guild_id<int>", e.CreateInvite)
}

type entity struct {
	name string

	// Services
	log   *slog.Logger
	mqt   mq.SendTransporter
	imq   *indexmq.IndexMQ
	cache cache.Cache

	// DB entities
	user  user.User
	disc  discriminator.Discriminator
	ch    channel.Channel
	g     guild.Guild
	gc    guildchannels.GuildChannels
	msg   message.Message
	at    attachment.Attachment
	perm  rolecheck.RoleCheck
	uperm channeluserperm.ChannelUserPerm
	rperm channelroleperm.ChannelRolePerm
	role  role.Role
	ur    userrole.UserRole
	icon  icon.Icon
	memb  member.Member
	inv   invite.Invite
	av    avatar.Avatar
	// Config
	attachTTL          int64
	authSecret         string
	defaultVoiceRegion string
	disco              discovery.Manager
	allowedRegions     map[string]struct{}
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, pg *pgdb.DB, mqt mq.SendTransporter, imq *indexmq.IndexMQ, cache cache.Cache, attachTTLSeconds int64, authSecret string, defaultVoiceRegion string, disco discovery.Manager, allowedRegions []string, log *slog.Logger) server.Entity {
	ar := make(map[string]struct{}, len(allowedRegions))
	for _, r := range allowedRegions {
		if r == "" {
			continue
		}
		ar[r] = struct{}{}
	}
	return &entity{
		name:               entityName,
		log:                log,
		mqt:                mqt,
		imq:                imq,
		cache:              cache,
		user:               user.New(pg.Conn()),
		disc:               discriminator.New(pg.Conn()),
		ch:                 channel.New(pg.Conn()),
		g:                  guild.New(pg.Conn()),
		gc:                 guildchannels.New(pg.Conn()),
		msg:                message.New(dbcon),
		at:                 attachment.New(dbcon),
		perm:               rolecheck.New(dbcon, pg),
		uperm:              channeluserperm.New(pg.Conn()),
		rperm:              channelroleperm.New(pg.Conn()),
		role:               role.New(pg.Conn()),
		ur:                 userrole.New(pg.Conn()),
		icon:               icon.New(dbcon),
		memb:               member.New(pg.Conn()),
		inv:                invite.New(pg.Conn()),
		av:                 avatar.New(dbcon),
		attachTTL:          attachTTLSeconds,
		authSecret:         authSecret,
		defaultVoiceRegion: defaultVoiceRegion,
		disco:              disco,
		allowedRegions:     ar,
	}
}
