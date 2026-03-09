package upload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type PreparedBody struct {
	Reader      io.Reader
	Size        int64
	ContentType string
}

type BufferedBody struct {
	Data        []byte
	Size        int64
	ContentType string
}

func PrepareBody(reader io.Reader, expectedSize int64) (*PreparedBody, error) {
	if expectedSize <= 0 {
		return nil, fmt.Errorf("%w: invalid expected size", ErrSizeMismatch)
	}

	sniffSize := 512
	if expectedSize < int64(sniffSize) {
		sniffSize = int(expectedSize)
	}

	sniff := make([]byte, sniffSize)
	n, err := io.ReadFull(reader, sniff)
	switch err {
	case nil:
	case io.EOF:
		if n == 0 {
			return nil, ErrEmptyBody
		}
	case io.ErrUnexpectedEOF:
	default:
		return nil, err
	}

	if n == 0 {
		return nil, ErrEmptyBody
	}

	sniff = sniff[:n]

	return &PreparedBody{
		Reader:      &exactSizeReader{reader: io.MultiReader(bytes.NewReader(sniff), reader), remaining: expectedSize},
		Size:        expectedSize,
		ContentType: http.DetectContentType(sniff),
	}, nil
}

func ReadBodyToMemory(reader io.Reader, expectedSize int64) (*BufferedBody, error) {
	prepared, err := PrepareBody(reader, expectedSize)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(prepared.Reader)
	if err != nil {
		return nil, err
	}

	return &BufferedBody{
		Data:        data,
		Size:        int64(len(data)),
		ContentType: prepared.ContentType,
	}, nil
}

type exactSizeReader struct {
	reader    io.Reader
	remaining int64
	eof       bool
}

func (r *exactSizeReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.eof {
		return 0, io.EOF
	}

	if r.remaining == 0 {
		var extra [1]byte
		for {
			n, err := r.reader.Read(extra[:])
			if n > 0 {
				return 0, ErrTooLarge
			}
			if err == io.EOF {
				r.eof = true
				return 0, io.EOF
			}
			if err != nil {
				return 0, err
			}
		}
	}

	if int64(len(p)) > r.remaining {
		p = p[:int(r.remaining)]
	}

	n, err := r.reader.Read(p)
	if n > 0 {
		r.remaining -= int64(n)
	}
	if err == io.EOF {
		if r.remaining > 0 {
			return n, ErrSizeMismatch
		}
		r.eof = true
	}
	return n, err
}
