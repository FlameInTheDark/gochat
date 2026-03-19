package message

import (
	"errors"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrUnableToGetUserToken         = "unable to get user token"
	ErrUnableToParseBody            = "unable to parse body"
	ErrPermissionsRequired          = "permissions required"
	ErrUnableToCreateAttachment     = "unable to create attachment"
	ErrUnableToCreateUploadURL      = "unable to create upload url"
	ErrIncorrectChannelID           = "incorrect channel ID"
	ErrIncorrectMessageID           = "incorrect message ID"
	ErrFileIsTooBig                 = "file is too big"
	ErrUnableToSendMessage          = "unable to send message"
	ErrUnableToUpdateMessage        = "unable to update message"
	ErrUnableToGetUser              = "unable to get user"
	ErrUnableToGetGuild             = "unable to get guild"
	ErrUnableToGetUserDiscriminator = "unable to get discriminator"
	ErrUnableToGetAttachements      = "unable to get attachments"
	ErrUnableToSentToThisChannel    = "unable to send to this channel"
	ErrUnableToReadFromThisChannel  = "unable to read from this channel"
	ErrUnableToGetMessage           = "unable to get message"
	ErrUnableToSetReadState         = "unable to set read state"
	ErrUnableToSendTypingEvent      = "unable to send typing event"
	ErrInvalidAttachments           = "invalid attachments"
	ErrUnableToCreateThread         = "unable to create thread"
	ErrThreadClosed                 = "thread is closed"
	ErrThreadSourceInvalid          = "threads can only be created from guild text channel messages"
	ErrThreadNestingForbidden       = "cannot create a thread inside a thread"
	ErrThreadAlreadyExists          = "thread already exists for this message"
	ErrThreadNameTooLong            = "thread name must be 256 characters or fewer"
	ErrMessageNotEditable           = "message type cannot be edited"
	ErrNonceRequiredWithEnforce     = "nonce is required when enforce_nonce is true"
	ErrReferenceIdInvalid           = "reference ID must be positive"

	// Validation error messages
	ErrMessagePayloadRequired = "message content, attachments, or embeds are required"
	ErrMessageUpdateRequired  = "message content, embeds, or flags update is required"
	ErrMessageContentTooLong  = "message content must be less than 2000 characters"
	ErrAttachmentIdInvalid    = "attachment ID must be positive"
	ErrMentionIdInvalid       = "mention ID must be positive"
	ErrFilenameRequired       = "filename is required"
	ErrFilenameTooLong        = "filename must be less than 255 characters"
	ErrFileSizeInvalid        = "file size must be positive"
	ErrFileSizeTooBig         = "file size must be less than 100MB"
	ErrDimensionsInvalid      = "width and height must be non-negative"
	ErrLimitInvalid           = "limit must be between 1 and 100"
	ErrFromIdInvalid          = "from ID must be positive"
	ErrDirectionInvalid       = "direction must be 'before' or 'after'"
	ErrFlagsInvalid           = "flags must be non-negative"
)

const (
	maxThreadNameLength = 256
)

type SendMessageRequest struct {
	Content      string                  `json:"content" example:"Hello world!"`                                          // Message content
	Nonce        *helper.MessageNonce    `json:"nonce,omitempty" swaggertype:"string" example:"draft-1"`                  // Optional client correlation token echoed back to the author.
	EnforceNonce bool                    `json:"enforce_nonce,omitempty"`                                                 // When true, deduplicates sends with the same nonce in the same channel for a short window.
	Reference    *helper.StringInt64     `json:"reference,omitempty" swaggertype:"integer" example:"2230469276416868352"` // Referenced message ID in the same channel. When set, the new message is stored as type 1 (Reply).
	Attachments  helper.StringInt64Array `json:"attachments" example:"2230469276416868352"`                               // IDs of attached files
	Mentions     helper.StringInt64Array `json:"mentions" example:"2230469276416868352"`                                  // IDs of mentioned users
	Embeds       []embed.Embed           `json:"embeds,omitempty"`                                                        // Manual embeds supplied by the client. These are stored separately from generated URL embeds.
}

func (r SendMessageRequest) Validate() error {
	if r.EnforceNonce && (r.Nonce == nil || r.Nonce.IsZero()) {
		return errors.New(ErrNonceRequiredWithEnforce)
	}
	if r.Reference != nil && int64(*r.Reference) < 1 {
		return errors.New(ErrReferenceIdInvalid)
	}
	return validateMessagePayload(r.Content, r.Attachments, r.Mentions, r.Embeds)
}

type CreateThreadRequest struct {
	Name        string                  `json:"name,omitempty" example:"Thread title"`                  // Optional explicit thread name.
	Content     string                  `json:"content" example:"Hello from thread!"`                   // First thread message content.
	Nonce       *helper.MessageNonce    `json:"nonce,omitempty" swaggertype:"string" example:"draft-1"` // Optional client correlation token for the starter message event.
	Attachments helper.StringInt64Array `json:"attachments" example:"2230469276416868352"`              // IDs of attached files uploaded to the parent channel before thread creation.
	Mentions    helper.StringInt64Array `json:"mentions" example:"2230469276416868352"`                 // IDs of mentioned users.
	Embeds      []embed.Embed           `json:"embeds,omitempty"`                                       // Manual embeds for the first thread message.
}

func (r CreateThreadRequest) Validate() error {
	if err := validateMessagePayload(r.Content, r.Attachments, r.Mentions, r.Embeds); err != nil {
		return err
	}

	name := strings.TrimSpace(r.Name)
	if name == "" {
		return nil
	}

	return validation.Validate(name,
		validation.RuneLength(1, maxThreadNameLength).Error(ErrThreadNameTooLong),
	)
}

func (r CreateThreadRequest) MessageRequest(attachmentIDs []int64) *SendMessageRequest {
	var nonce *helper.MessageNonce
	if r.Nonce != nil {
		nonce = r.Nonce.Clone()
	}
	return &SendMessageRequest{
		Content:     r.Content,
		Nonce:       nonce,
		Attachments: helper.StringInt64Array(attachmentIDs),
		Mentions:    r.Mentions,
		Embeds:      append([]embed.Embed(nil), r.Embeds...),
	}
}

type UpdateMessageRequest struct {
	Content *string        `json:"content,omitempty" example:"Hello world!"` // Message content
	Embeds  *[]embed.Embed `json:"embeds,omitempty"`                         // Full replacement for the manual embed array. Generated embeds are managed by the embedder service.
	Flags   *int           `json:"flags,omitempty" example:"4"`              // Message flags bitmask. Use 4 to suppress URL embed generation and clear generated embeds.
}

func (r UpdateMessageRequest) Validate() error {
	if r.Content == nil && r.Embeds == nil && r.Flags == nil {
		return errors.New(ErrMessageUpdateRequired)
	}

	if r.Content != nil {
		if err := validation.Validate(*r.Content,
			validation.RuneLength(0, 2000).Error(ErrMessageContentTooLong),
		); err != nil {
			return err
		}
	}

	if r.Embeds != nil {
		if err := embed.ValidateEmbeds(*r.Embeds); err != nil {
			return err
		}
	}

	if r.Flags != nil {
		if err := validation.Validate(*r.Flags,
			validation.Min(0).Error(ErrFlagsInvalid),
		); err != nil {
			return err
		}
	}

	return nil
}

func validateMessagePayload(content string, attachments, mentions helper.StringInt64Array, embeds []embed.Embed) error {
	payload := struct {
		Content     string
		Attachments helper.StringInt64Array
		Mentions    helper.StringInt64Array
	}{
		Content:     content,
		Attachments: attachments,
		Mentions:    mentions,
	}

	if err := validation.ValidateStruct(&payload,
		validation.Field(&payload.Content,
			validation.When(len(payload.Attachments) == 0 && len(embeds) == 0, validation.Required.Error(ErrMessagePayloadRequired)),
			validation.RuneLength(0, 2000).Error(ErrMessageContentTooLong),
		),
		validation.Field(&payload.Attachments,
			validation.Each(validation.Min(int64(1)).Error(ErrAttachmentIdInvalid)),
		),
		validation.Field(&payload.Mentions,
			validation.Each(validation.Min(int64(1)).Error(ErrMentionIdInvalid)),
		),
	); err != nil {
		return err
	}

	return embed.ValidateEmbeds(embeds)
}

type UploadAttachmentRequest struct {
	Filename    string `json:"filename" example:"image.png"`     // File name
	FileSize    int64  `json:"file_size" example:"100000"`       // File size in bytes
	Width       int64  `json:"width" example:"800"`              // Image width in pixels
	Height      int64  `json:"height" example:"600"`             // Image height in pixels
	ContentType string `json:"content_type" example:"image/png"` // File content-type meta data
}

func (r UploadAttachmentRequest) Validate() error {
	const maxFileSize = 100 * 1024 * 1024 // 100MB in bytes

	return validation.ValidateStruct(&r,
		validation.Field(&r.Filename,
			validation.Required.Error(ErrFilenameRequired),
			validation.RuneLength(1, 255).Error(ErrFilenameTooLong),
		),
		validation.Field(&r.FileSize,
			validation.Required.Error(ErrFileSizeInvalid),
			validation.Min(int64(1)).Error(ErrFileSizeInvalid),
			validation.Max(int64(maxFileSize)).Error(ErrFileSizeTooBig),
		),
		validation.Field(&r.Width,
			validation.Min(int64(0)).Error(ErrDimensionsInvalid),
		),
		validation.Field(&r.Height,
			validation.Min(int64(0)).Error(ErrDimensionsInvalid),
		),
	)
}

type Direction string

const (
	DirectionBefore = Direction("before")
	DirectionAfter  = Direction("after")
	DirectionAround = Direction("around")
)

const (
	DefaultLimit = int(50)
)

type GetMessagesRequest struct {
	From      *int64     `query:"from" json:"from" example:"2230469276416868352"`                          // ID of the message whe start to look from
	Limit     *int       `query:"limit" json:"limit" example:"30"`                                         // Number of messages to return.
	Direction *Direction `query:"direction" json:"direction" enums:"before,after,around" example:"before"` // Direction to look for messages
}

func (r GetMessagesRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.From,
			validation.When(r.From != nil, validation.Min(int64(1)).Error(ErrFromIdInvalid)),
		),
		validation.Field(&r.Limit,
			validation.When(r.Limit != nil,
				validation.Min(1).Error(ErrLimitInvalid),
				validation.Max(100).Error(ErrLimitInvalid),
			),
		),
		validation.Field(&r.Direction,
			validation.When(r.Direction != nil,
				validation.In(DirectionBefore, DirectionAfter, DirectionAround).Error(ErrDirectionInvalid),
			),
		),
	)
}
