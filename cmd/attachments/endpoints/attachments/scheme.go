package attachments

const (
	ErrUnableToGetUserToken       = "unable to get user token"
	ErrIncorrectChannelID         = "incorrect channel ID"
	ErrIncorrectAttachmentID      = "incorrect attachment ID"
	ErrUnableToGetAttachment      = "unable to get attachment"
	ErrForbiddenToUpload          = "forbidden to upload"
	ErrUnableToReadBody           = "unable to read request body"
	ErrFileIsTooBig               = "file is too big"
	ErrUnsupportedContentType     = "unsupported content type"
	ErrUnableToProcessImage       = "unable to process image"
	ErrUnableToProcessVideo       = "unable to process video"
	ErrUnableToUploadToStorage    = "unable to upload to storage"
	ErrUnableToFinalizeAttachment = "unable to finalize attachment"
)

const (
	previewMaxSize = 350
)
