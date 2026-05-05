package search

import (
	"strings"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrPermissionsRequired  = "permissions required"
	ErrIncorrectGuildID     = "incorrect guild ID"
	ErrUnableToGetUserToken = "unable to get user token"
	ErrUnableToParseBody    = "unable to parse body"
	ErrUnableToFindMessages = "unable to find messages"
	ErrUnableToGetMessages  = "unable to get messages"
	ErrUnableToGetUsers     = "unable to get users"
	ErrUnableToGetGuildByID = "unable to get guild by id"
	ErrUnsupportedChannel   = "search is only available in text channels"

	// Validation error messages
	ErrMentionIdInvalid   = "mention ID must be positive"
	ErrIncorrectChannelID = "incorrect channel ID"
	ErrChannelIDRequired  = "channel ID is required"
	ErrPageInvalid        = "page must be non-negative"
	ErrContentTooLong     = "content must be 2000 characters or fewer"
	ErrHasInvalid         = "has values must be one of: url, image, video, file"
	ErrSortInvalid        = "sort must be one of: best_match, popularity, alphabetical"
)

const maxSearchContentLength = 2000

type MessageSearchRequest struct {
	ChannelId int64                   `json:"channel_id,string" example:"2230469276416868352"` // Channel ID to search in. Required.
	Mentions  helper.StringInt64Array `json:"mentions" example:"2230469276416868352"`          // Mentions contains a list of int64 user IDs.
	AuthorId  *int64                  `json:"author_id,string" example:"2230469276416868352"`  // Author ID to search by.
	Content   *string                 `json:"content" example:"Hello world!"`                  // Content contains a string to search for. Might be empty if need to search by other parameters.
	Has       []string                `json:"has" enums:"url,image,video,file"`                // List of specific features to search for.
	Page      int                     `json:"page" default:"0"`                                // Page number to get. Starts from 0.
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
		validation.Field(&r.Page,
			validation.Min(0).Error(ErrPageInvalid),
		),
		validation.Field(&r.Content,
			validation.When(r.Content != nil,
				validation.RuneLength(0, maxSearchContentLength).Error(ErrContentTooLong),
			),
		),
		validation.Field(&r.Has,
			validation.Each(validation.In("url", "image", "video", "file").Error(ErrHasInvalid)),
		),
	)
}

func normalizeGuildDiscoveryTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}
