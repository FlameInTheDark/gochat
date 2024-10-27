package discriminator

import "github.com/FlameInTheDark/gochat/internal/database/db"

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
