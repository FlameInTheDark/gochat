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
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/friend"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guilddiscovery"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/threadmember"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/usersettings"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/searchmq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
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

	router.Post("/me/avatar", e.CreateAvatar)
	router.Get("/me/avatars", e.ListAvatars)
	router.Delete("/me/avatars/:avatar_id<int>", e.DeleteAvatar)

	router.Get("/me/friends", e.GetFriends)
	router.Get("/me/friends/:user_id<int>", e.GetOrCreateFriendDM)
	router.Post("/me/friends", e.CreateFriendRequest)
	router.Delete("/me/friends", e.Unfriend)
	router.Get("/me/friends/requests", e.GetFriendRequests)
	router.Post("/me/friends/requests", e.AcceptFriendRequest)
	router.Delete("/me/friends/requests", e.DeclineFriendRequest)

	router.Get("/me/settings", e.GetUserSettings)
	router.Post("/me/settings", e.SetUserSettings)

	router.Post("/me/channels/:channel_id<int>/call", e.StartDMCall)
	router.Post("/me/channels/:channel_id<int>/call/join", e.JoinDMCall)
	router.Post("/me/channels/:channel_id<int>/call/decline", e.DeclineDMCall)
	router.Delete("/me/channels/:channel_id<int>/call", e.LeaveDMCall)
	router.Get("/me/channels/:channel_id<int>/call/streams", e.ListDMCallStreams)
	router.Post("/me/channels/:channel_id<int>/call/streams", e.StartDMCallStream)
	router.Post("/me/channels/:channel_id<int>/call/streams/:stream_id<int>/join", e.JoinDMCallStream)
	router.Delete("/me/channels/:channel_id<int>/call/streams/:stream_id<int>", e.StopDMCallStream)
}

type entity struct {
	name string

	log   *slog.Logger
	mqt   mq.SendTransporter
	smq   *searchmq.Queue
	cache cache.Cache

	user    user.User
	member  member.Member
	guild   guild.Guild
	gd      guilddiscovery.GuildDiscovery
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
	emoji   emojirepo.Emoji
	tm      threadmember.ThreadMember

	attachTTL    int64
	contentHosts []string

	authSecret         string
	defaultVoiceRegion string
	disco              discovery.Manager
	streamDisco        discovery.Manager
	allowedRegions     map[string]struct{}
	allowedRegionIDs   []string
	voiceSelector      *dmVoiceSelector
	streamSelector     *dmVoiceSelector
}

func (e *entity) Name() string {
	return e.name
}

func New(cql *db.CQLCon, pg *pgdb.DB, mqt mq.SendTransporter, smq *searchmq.Queue, cache cache.Cache, attachTTLSeconds int64, contentHosts []string, authSecret, defaultVoiceRegion string, disco, streamDisco discovery.Manager, allowedRegions []string, log *slog.Logger) server.Entity {
	ar := make(map[string]struct{}, len(allowedRegions))
	regionIDs := make([]string, 0, len(allowedRegions))
	for _, region := range allowedRegions {
		if region == "" {
			continue
		}
		ar[region] = struct{}{}
		regionIDs = append(regionIDs, region)
	}
	return &entity{
		name:               entityName,
		log:                log,
		mqt:                mqt,
		smq:                smq,
		cache:              cache,
		attachTTL:          attachTTLSeconds,
		contentHosts:       append([]string(nil), contentHosts...),
		authSecret:         authSecret,
		defaultVoiceRegion: defaultVoiceRegion,
		disco:              disco,
		streamDisco:        streamDisco,
		allowedRegions:     ar,
		allowedRegionIDs:   regionIDs,
		voiceSelector:      newDMVoiceSelector(),
		streamSelector:     newDMVoiceSelector(),
		user:               user.New(pg.Conn()),
		member:             member.New(pg.Conn()),
		guild:              guild.New(pg.Conn()),
		gd:                 guilddiscovery.New(pg.Conn()),
		urole:              userrole.New(pg.Conn()),
		ch:                 channel.New(pg.Conn()),
		dm:                 dmchannel.New(pg.Conn()),
		gdm:                groupdmchannel.New(pg.Conn()),
		disc:               discriminator.New(pg.Conn()),
		fr:                 friend.New(pg.Conn()),
		uset:               usersettings.New(pg.Conn()),
		rs:                 readstates.New(cql),
		gclm:               guildchannelmessages.New(cql),
		dmlm:               dmchannelmessages.New(cql),
		av:                 avatar.New(cql),
		icon:               icon.New(cql),
		mention:            mention.New(cql),
		gc:                 guildchannels.New(pg.Conn()),
		emoji:              emojirepo.New(pg.Conn()),
		tm:                 threadmember.New(pg.Conn()),
	}
}
