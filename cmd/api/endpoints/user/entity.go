package user

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/avatar"
	"github.com/FlameInTheDark/gochat/internal/database/entities/dmchannelmessages"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannelmessages"
	"github.com/FlameInTheDark/gochat/internal/database/entities/icon"
	"github.com/FlameInTheDark/gochat/internal/database/entities/mention"
	"github.com/FlameInTheDark/gochat/internal/database/entities/readstates"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/friend"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/usersettings"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "user"

func (e *entity) Init(router fiber.Router) {
	router.Get("/:user_id", e.GetUser)
	router.Patch("/me", e.ModifyUser)
	router.Get("/me/guilds", e.GetUserGuilds)
	router.Get("/me/guilds/:guild_id<int>/member", e.GetMyGuildMember)
	router.Delete("/me/guilds/:guild_id<int>", e.LeaveGuild)
	router.Get("/me/channels", e.GetMyDMChannels)
	router.Post("/me/channels", e.CreateDM)

	// Avatar
	router.Post("/me/avatar", e.CreateAvatar)
	router.Get("/me/avatars", e.ListAvatars)
	router.Delete("/me/avatars/:avatar_id<int>", e.DeleteAvatar)

	// Friends
	router.Get("/me/friends", e.GetFriends)                        // Get a friends list
	router.Get("/me/friends/:user_id<int>", e.GetOrCreateFriendDM) // Get DM channel or create it if not exist
	router.Post("/me/friends", e.CreateFriendRequest)              // Send friend request (search by discriminator string and send request if user did not block us)
	router.Delete("/me/friends", e.Unfriend)                       // Unfriend users
	router.Get("/me/friends/requests", e.GetFriendRequests)        // Get a list of friend requests
	router.Post("/me/friends/requests", e.AcceptFriendRequest)     // Accept a friend request
	router.Delete("/me/friends/requests", e.DeclineFriendRequest)  // Decline a friend request by deleting it

	// User settings
	router.Get("/me/settings", e.GetUserSettings)
	router.Post("/me/settings", e.SetUserSettings)
}

type entity struct {
	name string

	// Services
	log   *slog.Logger
	mqt   mq.SendTransporter
	cache cache.Cache

	// DB entities
	user    user.User
	member  member.Member
	guild   guild.Guild
	urole   userrole.UserRole
	ch      channel.Channel
	dm      dmchannel.DmChannel
	gdm     groupdmchannel.GroupDMChannel
	disc    discriminator.Discriminator
	fr      friend.Friend
	uset    usersettings.UserSettings
	rs      readstates.ReadStates
	gclm    guildchannelmessages.GuildChannelMessages
	dmlm    *dmchannelmessages.Entity
	av      avatar.Avatar
	icon    icon.Icon
	mention mention.Mention
	gc      guildchannels.GuildChannels

	// Config
	s3Base    string
	attachTTL int64
}

func (e *entity) Name() string {
	return e.name
}

func New(cql *db.CQLCon, pg *pgdb.DB, mqt mq.SendTransporter, cache cache.Cache, attachTTLSeconds int64, log *slog.Logger) server.Entity {
	return &entity{
		name:      entityName,
		log:       log,
		mqt:       mqt,
		cache:     cache,
		attachTTL: attachTTLSeconds,
		user:      user.New(pg.Conn()),
		member:    member.New(pg.Conn()),
		guild:     guild.New(pg.Conn()),
		urole:     userrole.New(pg.Conn()),
		ch:        channel.New(pg.Conn()),
		dm:        dmchannel.New(pg.Conn()),
		gdm:       groupdmchannel.New(pg.Conn()),
		disc:      discriminator.New(pg.Conn()),
		fr:        friend.New(pg.Conn()),
		uset:      usersettings.New(pg.Conn()),
		rs:        readstates.New(cql),
		gclm:      guildchannelmessages.New(cql),
		dmlm:      dmchannelmessages.New(cql),
		av:        avatar.New(cql),
		icon:      icon.New(cql),
		mention:   mention.New(cql),
		gc:        guildchannels.New(pg.Conn()),
	}
}
