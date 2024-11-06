package dto

type Message struct {
	Id          int64        `json:"id"`
	ChannelId   int64        `json:"channel_id"`
	Author      User         `json:"author_id"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	ContentType *string `json:"content_type"`
	Filename    string  `json:"filename"`
	Height      *int64  `json:"height"`
	Width       *int64  `json:"width"`
	URL         string  `json:"url"`
	Size        int64   `json:"size"`
}
