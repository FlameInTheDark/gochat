package dto

import "time"

type Message struct {
	Id          int64        `json:"id" example:"2230469276416868352"`         // Message ID
	ChannelId   int64        `json:"channel_id" example:"2230469276416868352"` // Channel id the message was sent to
	Author      User         `json:"author"`
	Content     string       `json:"content" example:"Hello world!"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Type        int          `json:"type" example:"0"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"` // Timestamp of the last message edit
}

type Attachment struct {
	ContentType *string `json:"content_type,omitempty" example:"image/png"`                             // File mime type
	Filename    string  `json:"filename" example:"image.png"`                                           // File name
	Height      *int64  `json:"height,omitempty" example:"600"`                                         // Image dimensions in pixels
	Width       *int64  `json:"width,omitempty" example:"800"`                                          // Image dimensions in pixels
	URL         string  `json:"url" example:"https://example.com/image.png"`                            // URL to download the file
	PreviewURL  *string `json:"preview_url,omitempty" example:"https://example.com/image_preview.webp"` // Preview URL for image/video
	Size        int64   `json:"size" example:"1000000"`                                                 // FileSize in bytes
}
