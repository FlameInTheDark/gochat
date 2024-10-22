package member

import "github.com/FlameInTheDark/gochat/internal/database/db"

type Entity struct {
	Name string
	c    *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
