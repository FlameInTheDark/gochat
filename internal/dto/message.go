package dto

import (
	"time"

	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/helper"
)

type Message struct {
	Id                 int64                `json:"id" example:"2230469276416868352"`         // Message ID
	ChannelId          int64                `json:"channel_id" example:"2230469276416868352"` // Channel id the message was sent to
	Author             User                 `json:"author"`
	Content            string               `json:"content" example:"Hello world!"`
	Position           *int64               `json:"position,omitempty" example:"512"`                       // Monotonic channel-local message position used for navigation.
	Nonce              *helper.MessageNonce `json:"nonce,omitempty" swaggertype:"string" example:"draft-1"` // Ephemeral client correlation token echoed only to the author.
	Attachments        []Attachment         `json:"attachments,omitempty"`
	Embeds             []embed.Embed        `json:"embeds,omitempty"`
	Flags              int                  `json:"flags,omitempty"` // Bitmask. Includes suppress-embeds and banned-author markers in API responses.
	Type               int                  `json:"type" example:"0"`
	Reference          *int64               `json:"reference,omitempty" example:"2230469276416868352"`            // Referenced source message id.
	ReferenceChannelId *int64               `json:"reference_channel_id,omitempty" example:"2230469276416868352"` // Channel id of the referenced source message.
	ThreadId           *int64               `json:"thread_id,omitempty" example:"2230469276416868352"`            // Thread linked from this message.
	Thread             *Channel             `json:"thread,omitempty"`                                             // Thread metadata when the message is linked to a thread.
	UpdatedAt          *time.Time           `json:"updated_at,omitempty"`                                         // Timestamp of the last message edit
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
