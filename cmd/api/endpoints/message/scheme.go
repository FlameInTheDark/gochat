package message

const (
	ErrUnableToGetUserToken         = "unable to get user token"
	ErrUnableToParseBody            = "unable to parse body"
	ErrPermissionsRequired          = "permissions required"
	ErrUnableToCreateAttachment     = "unable to create attachment"
	ErrUnableToCreateUploadURL      = "unable to create upload url"
	ErrIncorrectChannelID           = "incorrect channel ID"
	ErrFileIsTooBig                 = "file is too big"
	ErrUnableToSendMessage          = "unable to send message"
	ErrUnableToGetUser              = "unable to get user"
	ErrUnableToGetUserDiscriminator = "unable to get discriminator"
	ErrUnableToGetAttachements      = "unable to get attachments"
	ErrUnableToSentToThisChannel    = "unable to send to this channel"
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
