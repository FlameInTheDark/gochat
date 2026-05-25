package model

type Banner struct {
	Id          int64
	UserId      int64
	URL         *string
	ContentType *string
	Width       *int64
	Height      *int64
	FileSize    int64
	Done        bool
}
