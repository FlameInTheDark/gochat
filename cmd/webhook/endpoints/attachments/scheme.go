package attachments

import "fmt"

type FinalizeRequest struct {
	ID          int64   `json:"id"`
	ChannelID   int64   `json:"channel_id"`
	AuthorID    *int64  `json:"author_id,omitempty"`
	ContentType *string `json:"content_type,omitempty"`
	URL         *string `json:"url,omitempty"`
	PreviewURL  *string `json:"preview_url,omitempty"`
	Height      *int64  `json:"height,omitempty"`
	Width       *int64  `json:"width,omitempty"`
	FileSize    *int64  `json:"file_size,omitempty"`
	Name        *string `json:"name,omitempty"`
}

func (r FinalizeRequest) Validate() error {
	if r.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if r.ChannelID == 0 {
		return fmt.Errorf("channel_id is required")
	}
	return nil
}
