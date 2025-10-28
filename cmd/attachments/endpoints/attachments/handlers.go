package attachments

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

const defaultUploadLimit int64 = 50 * 1000 * 1000 // 50 MB in bytes

// Upload
//
//	@Summary		Upload attachment
//	@Description	Uploads a file for an existing attachment. Stores the original as-is and generates a WebP preview for images/videos. Finalizes the attachment metadata.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			channel_id		path		int64	true	"Channel ID"
//	@Param			attachment_id	path		int64	true	"Attachment ID"
//	@Param			file			body		[]byte	true	"Binary file to upload"
//	@Success		201				{string}	string	"Created"
//	@Success		204				{string}	string	"No Content (already uploaded)"
//	@failure		400				{string}	string	"Bad request"
//	@failure		401				{string}	string	"Unauthorized"
//	@failure		403				{string}	string	"Forbidden"
//	@failure		404				{string}	string	"Attachment not found"
//	@failure		413				{string}	string	"File too large"
//	@failure		500				{string}	string	"Internal server error"
//	@Router			/upload/attachments/{channel_id}/{attachment_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	attachmentIdStr := c.Params("attachment_id")
	attachmentId, err := strconv.ParseInt(attachmentIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectAttachmentID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	at, err := e.at.GetAttachment(c.UserContext(), attachmentId, channelId)
	if err != nil {
		slog.Error(err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetAttachment)
	}
	if at.AuthorId == nil || *at.AuthorId != user.Id || at.ChannelId != channelId || at.Id != attachmentId {
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	}
	if at.Done {
		return c.SendStatus(fiber.StatusNoContent)
	}

	// Build a streaming reader and detect content type
	full, ct, err := streamingBodyWithContentType(c)
	if err != nil {
		return err
	}

	base := strings.TrimRight(e.s3ExternalURL, "/")

	origName := at.Name

	var finalURL, previewURL *string
	var heightPtr, widthPtr *int64
	var actualSize int64

	kind := inferAttachmentKind(ct, origName)
	switch kind {
	case "image":
		key, publicURL, size, uerr := e.putToS3ViaPresign(c.UserContext(), channelId, attachmentId, origName, ct, at.FileSize, full, 2*time.Minute)
		if uerr != nil {
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		actualSize = size
		// Enforce user upload limit using actual stored size
		if limit := e.getUserUploadLimit(c.UserContext(), user.Id); limit > 0 && actualSize > limit {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
		}
		prevBuf, perr := generatePreviewWithRetry(publicURL, 3, 300*time.Millisecond)
		if perr != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to process image to create preview", slog.String("error", perr.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessImage)
		}
		if w, h, derr := ffprobeDimensions(publicURL); derr == nil {
			widthPtr, heightPtr = &w, &h
		}
		prevKey := fmt.Sprintf("media/%d/%d/preview.webp", channelId, attachmentId)
		if err := e.storage.UploadObject(c.UserContext(), prevKey, bytes.NewReader(prevBuf.Bytes()), "image/webp"); err != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload preview for image type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		f := publicURL
		p := pathJoin(base, prevKey)
		finalURL, previewURL = &f, &p

	case "video":
		key, publicURL, size, uerr := e.putToS3ViaPresign(c.UserContext(), channelId, attachmentId, origName, ct, at.FileSize, full, 5*time.Minute)
		if uerr != nil {
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		actualSize = size
		// Enforce user upload limit using actual stored size
		if limit := e.getUserUploadLimit(c.UserContext(), user.Id); limit > 0 && actualSize > limit {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
		}
		prevBuf, perr := generatePreviewWithRetry(publicURL, 3, 300*time.Millisecond)
		if perr != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to process video to create preview", slog.String("error", perr.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessVideo)
		}
		if w, h, derr := ffprobeDimensions(publicURL); derr == nil {
			widthPtr, heightPtr = &w, &h
		}
		prevKey := fmt.Sprintf("media/%d/%d/preview.webp", channelId, attachmentId)
		if err := e.storage.UploadObject(c.UserContext(), prevKey, bytes.NewReader(prevBuf.Bytes()), "image/webp"); err != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload preview for video type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		f := publicURL
		p := pathJoin(base, prevKey)
		finalURL, previewURL = &f, &p

	default:
		key, publicURL, size, uerr := e.putToS3ViaPresign(c.UserContext(), channelId, attachmentId, origName, ct, at.FileSize, full, 2*time.Minute)
		if uerr != nil {
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		actualSize = size
		// Enforce user upload limit using actual stored size
		if limit := e.getUserUploadLimit(c.UserContext(), user.Id); limit > 0 && actualSize > limit {
			_ = e.storage.RemoveAttachment(c.UserContext(), key)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
		}
		f := publicURL
		finalURL = &f
	}

	// Finalize attachment in DB
	if err := e.at.DoneAttachment(c.UserContext(), attachmentId, channelId, &ct, finalURL, previewURL, heightPtr, widthPtr, &actualSize, &origName, at.AuthorId); err != nil {
		slog.Error("unable to finalize attachment", slog.String("error", err.Error()))
		if finalURL != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), strings.TrimPrefix(*finalURL, strings.TrimRight(e.s3ExternalURL, "/")+"/"))
		}
		if previewURL != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), strings.TrimPrefix(*previewURL, strings.TrimRight(e.s3ExternalURL, "/")+"/"))
		}
		_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeAttachment)
	}

	incTransferred(kind, actualSize)

	return c.SendStatus(fiber.StatusCreated)
}

// Generate a webp preview that fits 350x350 preserving aspect ratio
func ffmpegExtractWebP(path string) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-y",
		"-i", path,
		"-vframes", "1",
		"-vf", "scale=350:350:force_original_aspect_ratio=decrease",
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w: %s", err, stderr.String())
	}
	return &out, nil
}

// ffmpegExtractWebPFromReader reads input from an io.Reader (stdin) and extracts a single webp preview frame.
func ffmpegExtractWebPFromReader(r io.Reader) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-y",
		"-i", "pipe:0",
		"-vframes", "1",
		"-vf", "scale=350:350:force_original_aspect_ratio=decrease",
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-",
	)
	cmd.Stdin = r
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w: %s", err, stderr.String())
	}
	return &out, nil
}

// ffprobeDimensions returns width and height using ffprobe for the first video stream
func ffprobeDimensions(path string) (width, height int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		path,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		return 0, 0, fmt.Errorf("ffprobe failed: %w: %s", runErr, stderr.String())
	}
	s := strings.TrimSpace(out.String())
	if s == "" {
		return 0, 0, fmt.Errorf("ffprobe returned empty output")
	}
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe output: %s", s)
	}
	w, werr := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if werr != nil {
		return 0, 0, fmt.Errorf("parse width: %w", werr)
	}
	h, herr := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if herr != nil {
		return 0, 0, fmt.Errorf("parse height: %w", herr)
	}
	return w, h, nil
}

// inferAttachmentKind classifies the upload as image/video/other using content type and filename extension.
func inferAttachmentKind(contentType, name string) string {
	ct := strings.ToLower(contentType)
	if strings.HasPrefix(ct, "image/") {
		return "image"
	}
	if strings.HasPrefix(ct, "video/") {
		return "video"
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	// Common image extensions
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff":
		return "image"
	// Common video extensions
	case ".mp4", ".m4v", ".mov", ".webm", ".mkv", ".avi", ".wmv", ".flv", ".ogv", ".3gp", ".3g2", ".ts", ".m2ts":
		return "video"
	}
	return ""
}

// generatePreviewWithRetry calls ffmpegExtractWebP with simple retries to mitigate eventual consistency of public URLs.
func generatePreviewWithRetry(url string, attempts int, delay time.Duration) (*bytes.Buffer, error) {
	var buf *bytes.Buffer
	var err error
	for i := 0; i < attempts; i++ {
		buf, err = ffmpegExtractWebP(url)
		if err == nil {
			return buf, nil
		}
		time.Sleep(delay)
		// backoff a bit
		if delay < 2*time.Second {
			delay *= 2
		}
	}
	return nil, err
}

// imageDimensionsFromBody tries to decode image config from the in-memory bytes
func imageDimensionsFromBody(body []byte) (w, h int, ok bool) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

func pathJoin(base, key string) string {
	if base == "" {
		return key
	}
	b := strings.TrimRight(base, "/")
	k := strings.TrimLeft(key, "/")
	return b + "/" + k
}

// getUserUploadLimit fetches user's custom limit from Postgres or returns the default.
func (e *entity) getUserUploadLimit(ctx context.Context, userId int64) int64 {
	limit := defaultUploadLimit

	if e.usr == nil {
		return limit
	}
	u, err := e.usr.GetUserById(ctx, userId)
	if err != nil || u.UploadLimit == nil {
		return limit
	}
	if *u.UploadLimit <= 0 {
		return limit
	}
	return *u.UploadLimit
}

// streamingBodyWithContentType returns a reader that includes the initial peek for content type detection.
func streamingBodyWithContentType(c *fiber.Ctx) (io.Reader, string, error) {
	var stream io.Reader
	if r := c.Context().RequestBodyStream(); r != nil {
		stream = r
	} else {
		b := c.Body()
		if len(b) == 0 {
			return nil, "", fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
		}
		stream = bytes.NewReader(b)
	}
	peek := make([]byte, 512)
	n, _ := io.ReadFull(stream, peek)
	if n == 0 {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	}
	peek = peek[:n]
	ct := http.DetectContentType(peek)
	full := io.MultiReader(bytes.NewReader(peek), stream)
	return full, ct, nil
}

// putToS3ViaPresign uploads the body to S3 using a presigned PUT and returns key, public URL and actual stored size.
func (e *entity) putToS3ViaPresign(ctx context.Context, channelId, attachmentId int64, name, contentType string, sizeHint int64, body io.Reader, timeout time.Duration) (key, publicURL string, actualSize int64, err error) {
	key = fmt.Sprintf("media/%d/%d/%s", channelId, attachmentId, name)
	putURL, perr := e.storage.MakeUploadAttachment(ctx, channelId, attachmentId, sizeHint, name)
	if perr != nil {
		return "", "", 0, perr
	}
	req, rerr := http.NewRequestWithContext(ctx, http.MethodPut, putURL, body)
	if rerr != nil {
		return "", "", 0, rerr
	}
	req.Header.Set("Content-Type", contentType)
	if sizeHint > 0 {
		req.ContentLength = sizeHint
	}
	httpClient := &http.Client{Timeout: timeout}
	resp, herr := httpClient.Do(req)
	if herr != nil {
		return "", "", 0, herr
	}
	_ = resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", "", 0, fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	base := strings.TrimRight(e.s3ExternalURL, "/")
	publicURL = pathJoin(base, key)

	if s, _, serr := e.storage.StatObject(ctx, key); serr == nil {
		actualSize = s
	}
	return key, publicURL, actualSize, nil
}
