package message

import (
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

	// Validation error messages
	ErrMessageContentRequired = "message content is required"
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
)

type SendMessageRequest struct {
	Content     string  `json:"content"`
	Attachments []int64 `json:"attachments"`
	Mentions    []int64 `json:"mentions"`
}

func (r SendMessageRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Content,
			validation.When(len(r.Attachments) == 0, validation.Required.Error(ErrMessageContentRequired)),
			validation.RuneLength(0, 2000).Error(ErrMessageContentTooLong),
		),
		validation.Field(&r.Attachments,
			validation.Each(validation.Min(int64(1)).Error(ErrAttachmentIdInvalid)),
		),
		validation.Field(&r.Mentions,
			validation.Each(validation.Min(int64(1)).Error(ErrMentionIdInvalid)),
		),
	)
}

type UpdateMessageRequest struct {
	Content string `json:"content"`
}

func (r UpdateMessageRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Content,
			validation.Required.Error(ErrMessageContentRequired),
			validation.RuneLength(1, 2000).Error(ErrMessageContentTooLong),
		),
	)
}

type UploadAttachmentRequest struct {
	Filename string `json:"filename"`
	FileSize int64  `json:"file_size"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
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
)

const (
	DefaultLimit = int(50)
)

type GetMessagesRequest struct {
	From      *int64     `query:"from" json:"from"`
	Limit     *int       `query:"limit" json:"limit"`
	Direction *Direction `query:"direction" json:"direction"`
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
				validation.In(DirectionBefore, DirectionAfter).Error(ErrDirectionInvalid),
			),
		),
	)
}
