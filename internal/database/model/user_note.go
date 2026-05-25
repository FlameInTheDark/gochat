package model

import "time"

type UserNote struct {
	OwnerUserId  int64     `db:"owner_user_id"`
	TargetUserId int64     `db:"target_user_id"`
	Note         string    `db:"note"`
	UpdatedAt    time.Time `db:"updated_at"`
}
