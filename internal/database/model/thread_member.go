package model

import "time"

type ThreadMember struct {
	ThreadId int64     `db:"thread_id"`
	UserId   int64     `db:"user_id"`
	Flags    int       `db:"flags"`
	JoinAt   time.Time `db:"join_at"`
}
