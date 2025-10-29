package model

type Icon struct {
	Id          int64
	GuildId     int64
	URL         *string
	ContentType *string
	Width       *int64
	Height      *int64
	FileSize    int64
	Done        bool
}
