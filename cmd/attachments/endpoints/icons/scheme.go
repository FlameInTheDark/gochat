package icons

const (
	ErrUnableToGetUserToken    = "unable to get user token"
	ErrIncorrectGuildID        = "incorrect guild ID"
	ErrIncorrectIconID         = "incorrect icon ID"
	ErrForbiddenToUpload       = "forbidden to upload"
	ErrUnableToReadBody        = "unable to read request body"
	ErrFileIsTooBig            = "file is too big"
	ErrUnsupportedContentType  = "unsupported content type"
	ErrUnableToProcessImage    = "unable to process image"
	ErrUnableToUploadToStorage = "unable to upload to storage"
)

const (
	iconMaxSizeBytes = 250 * 1024
	iconMaxDim       = 128
)
