package avatars

const (
	ErrUnableToGetUserToken    = "unable to get user token"
	ErrIncorrectUserID         = "incorrect user ID"
	ErrIncorrectAvatarID       = "incorrect avatar ID"
	ErrUnableToGetAvatar       = "unable to get avatar"
	ErrForbiddenToUpload       = "forbidden to upload"
	ErrUnableToReadBody        = "unable to read request body"
	ErrFileIsTooBig            = "file is too big"
	ErrUnsupportedContentType  = "unsupported content type"
	ErrUnableToProcessImage    = "unable to process image"
	ErrUnableToUploadToStorage = "unable to upload to storage"
	ErrUnableToFinalizeAvatar  = "unable to finalize avatar"
)

const (
	avatarMaxSizeBytes = 250 * 1024
	avatarMaxDim       = 128
)
