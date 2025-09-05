package search

import validation "github.com/go-ozzo/ozzo-validation/v4"

const (
	ErrPermissionsRequired = "permissions required"
	ErrIncorrectGuildID    = "incorrect guild ID"
	ErrUnableToParseBody   = "unable to parse body"

	// Validation error messages
	ErrMentionIdInvalid   = "mention ID must be positive"
	ErrIncorrectChannelID = "incorrect channel ID"
	ErrChannelIDRequired  = "channel ID is required"
)

type MessageSearchRequest struct {
	ChannelId *int64   `json:"channel_id"`
	Mentions  []int64  `json:"mentions"`
	AuthorId  *int64   `json:"author_id"`
	Content   *string  `json:"content"`
	Has       []string `json:"has"`
	Page      int      `json:"page"`
}

type MessageSearchResponse struct {
	Ids   []int64 `json:"ids"`
	Pages int     `json:"pages"`
}

func (r MessageSearchRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChannelId,
			validation.Required.Error(ErrChannelIDRequired),
			validation.Min(int64(1)).Error(ErrIncorrectChannelID),
		),
		validation.Field(&r.Mentions,
			validation.Each(validation.Min(int64(1)).Error(ErrMentionIdInvalid)),
		),
	)
}
