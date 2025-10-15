package model

type Attachment struct {
	Id          int64
	ChannelId   int64
	Name        string
	FileSize    int64
	ContentType *string
	Height      *int64
	Width       *int64
	URL         *string
	Done        bool
}
