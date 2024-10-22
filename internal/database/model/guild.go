package model

import "time"

type Guild struct {
	Id        int64
	Name      string
	OwnerId   int64
	Icon      *int64
	Public    bool
	CreatedAt time.Time
}
