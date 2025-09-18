package model

import "time"

type Guild struct {
	Id          int64     `db:"id"`
	Name        string    `db:"name"`
	OwnerId     int64     `db:"owner_id"`
	Icon        *int64    `db:"icon"`
	Public      bool      `db:"public"`
	Permissions int64     `db:"permissions"`
	CreatedAt   time.Time `db:"created_at"`
}
