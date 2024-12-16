package dto

import "time"

type Message struct {
	Id          int64        `json:"id"`
	ChannelId   int64        `json:"channel_id"`
	Author      User         `json:"author"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
}

type Attachment struct {
	ContentType *string `json:"content_type,omitempty"`
	Filename    string  `json:"filename"`
	Height      *int64  `json:"height,omitempty"`
	Width       *int64  `json:"width,omitempty"`
	URL         string  `json:"url"`
	Size        int64   `json:"size"`
}
