package channelroleperm

import (
	"github.com/jmoiron/sqlx"
)

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) *Entity {
	return &Entity{c: c}
}
