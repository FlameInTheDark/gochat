package model

type Discriminator struct {
	UserId        int64  `db:"user_id"`
	Discriminator string `db:"discriminator"`
}
