package rolecheck

import (
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/entities/member"
	"github.com/FlameInTheDark/gochat/internal/database/entities/role"
	"github.com/FlameInTheDark/gochat/internal/database/entities/userrole"
)

type Entity struct {
	c    *db.CQLCon
	role *role.Entity
	chrp *channelroleperm.Entity
	chup *channeluserperm.Entity
	ur   *userrole.Entity
	g    *guild.Entity
	gc   *guildchannels.Entity
	ch   *channel.Entity
	m    *member.Entity
}

func New(c *db.CQLCon) *Entity {
	return &Entity{
		c:    c,
		role: role.New(c),
		chrp: channelroleperm.New(c),
		chup: channeluserperm.New(c),
		ur:   userrole.New(c),
		g:    guild.New(c),
		gc:   guildchannels.New(c),
		ch:   channel.New(c),
		m:    member.New(c),
	}
}
