package rolecheck

import (
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/role"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/userrole"
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

func New(c *db.CQLCon, pg *pgdb.DB) *Entity {
	return &Entity{
		c:    c,
		role: role.New(pg.Conn()),
		chrp: channelroleperm.New(pg.Conn()),
		chup: channeluserperm.New(pg.Conn()),
		ur:   userrole.New(pg.Conn()),
		g:    guild.New(pg.Conn()),
		gc:   guildchannels.New(pg.Conn()),
		ch:   channel.New(pg.Conn()),
		m:    member.New(pg.Conn()),
	}
}
