package message

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
)

type SendMessageRequest struct {
	Content     string  `json:"content"`
	Attachments []int64 `json:"attachments"`
}

type UpdateMessageRequest struct {
	Content string `json:"content"`
}

type UploadAttachmentRequest struct {
	Filename string `json:"filename"`
	FileSize int64  `json:"file_size"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
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
