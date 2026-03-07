package upload

import "errors"

var (
	ErrPlaceholderNotFound = errors.New("upload placeholder not found")
	ErrForbidden           = errors.New("upload forbidden")
	ErrAlreadyDone         = errors.New("upload already completed")
	ErrEmptyBody           = errors.New("upload body is empty")
	ErrSizeMismatch        = errors.New("upload size mismatch")
	ErrTooLarge            = errors.New("upload body exceeds declared size")
	ErrUnsupportedMedia    = errors.New("unsupported media type")
	ErrMediaProcess        = errors.New("media processing failed")
	ErrStorage             = errors.New("storage operation failed")
	ErrFinalize            = errors.New("finalize operation failed")
)
