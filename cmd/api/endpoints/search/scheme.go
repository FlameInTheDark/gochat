package search

import (
	"github.com/FlameInTheDark/gochat/internal/dto"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrPermissionsRequired  = "permissions required"
	ErrIncorrectGuildID     = "incorrect guild ID"
	ErrUnableToParseBody    = "unable to parse body"
	ErrUnableToFindMessages = "unable to find messages"
	ErrUnableToGetMessages  = "unable to get messages"
	ErrUnableToGetUsers     = "unable to get users"

	// Validation error messages
	ErrMentionIdInvalid   = "mention ID must be positive"
	ErrIncorrectChannelID = "incorrect channel ID"
	ErrChannelIDRequired  = "channel ID is required"
)

type MessageSearchRequest struct {
	ChannelId int64    `json:"channel_id" example:"2230469276416868352"` // Channel ID to search in. Required.
	Mentions  []int64  `json:"mentions" example:"2230469276416868352"`   // Mentions contains a list of int64 user IDs.
	AuthorId  *int64   `json:"author_id" example:"2230469276416868352"`  // Author ID to search by.
	Content   *string  `json:"content" example:"Hello world!"`           // Content contains a string to search for. Might be empty if need to search by other parameters.
	Has       []string `json:"has" enums:"url,image,video,file"`         // List of specific features to search for.
	Page      int      `json:"page" default:"0"`                         // Page number to get. Starts from 0.
}

type MessageSearchResponse struct {
	Messages []dto.Message `json:"messages"` // List of messages
	Pages    int           `json:"pages"`    // Total number of pages with current search parameters
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
