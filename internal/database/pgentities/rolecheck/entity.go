package rolecheck

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type RoleCheck interface {
	getUserRoleIDs(ctx context.Context, guildID, userID int64) ([]int64, error)
	ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error)
	GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error)
	GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error)
}

type Entity struct {
	role role.Role
	chrp channelroleperm.ChannelRolePerm
	chup channeluserperm.ChannelUserPerm
	ur   userrole.UserRole
	g    guild.Guild
	gc   guildchannels.GuildChannels
	ch   channel.Channel
	m    member.Member
	dm   dmchannel.DmChannel
	gdm  groupdmchannel.GroupDMChannel
}

func New(pg *pgdb.DB) RoleCheck {
	return &Entity{
		role: role.New(pg.Conn()),
		chrp: channelroleperm.New(pg.Conn()),
		chup: channeluserperm.New(pg.Conn()),
		ur:   userrole.New(pg.Conn()),
		g:    guild.New(pg.Conn()),
		gc:   guildchannels.New(pg.Conn()),
		ch:   channel.New(pg.Conn()),
		m:    member.New(pg.Conn()),
		dm:   dmchannel.New(pg.Conn()),
		gdm:  groupdmchannel.New(pg.Conn()),
	}
}
