package attachments

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

// Upload
//
//	@Summary		Upload attachment
//	@Description	Uploads a file for an existing attachment. Stores the original as-is and generates a WebP preview for images/videos. Finalizes the attachment metadata.
//	@Tags			Attachments
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
//	@Router			/attachments/{channel_id}/{attachment_id} [post]
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

	body := c.Body()
	if len(body) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	}

	if at.FileSize > 0 && int64(len(body)) > at.FileSize {
		_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	}

	// Sniff content type
	head := body
	if len(head) > 512 {
		head = head[:512]
	}
	ct := http.DetectContentType(head)

	base := strings.TrimRight(e.s3ExternalURL, "/")

	origName := at.Name

	var finalURL, previewURL *string
	var heightPtr, widthPtr *int64

	switch {
	case strings.HasPrefix(ct, "image/"):
		fileKey := fmt.Sprintf("media/%d/%d/%s", channelId, attachmentId, origName)
		if err := e.storage.UploadObject(c.UserContext(), fileKey, bytes.NewReader(body), ct); err != nil {
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload object for image type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}

		publicURL := pathJoin(base, fileKey)
		if w, h, perr := ffprobeDimensions(publicURL); perr == nil {
			widthPtr, heightPtr = &w, &h
		} else {
			if iw, ih, ok := imageDimensionsFromBody(body); ok {
				wv, hv := int64(iw), int64(ih)
				widthPtr, heightPtr = &wv, &hv
			}
		}
		prevBuf, err := ffmpegExtractWebP(publicURL)
		if err != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), fileKey)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to process image to create preview", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessImage)
		}

		prevKey := fmt.Sprintf("media/%d/%d/preview.webp", channelId, attachmentId)
		if err := e.storage.UploadObject(c.UserContext(), prevKey, bytes.NewReader(prevBuf.Bytes()), "image/webp"); err != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), fileKey)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload preview for image type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		f := publicURL
		p := pathJoin(base, prevKey)
		finalURL, previewURL = &f, &p

	case strings.HasPrefix(ct, "video/"):
		videoKey := fmt.Sprintf("media/%d/%d/%s", channelId, attachmentId, origName)
		if err := e.storage.UploadObject(c.UserContext(), videoKey, bytes.NewReader(body), ct); err != nil {
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload object for video type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}

		publicURL := pathJoin(base, videoKey)

		if w, h, perr := ffprobeDimensions(publicURL); perr == nil {
			widthPtr, heightPtr = &w, &h
		}
		prevBuf, err := ffmpegExtractWebP(publicURL)
		if err != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), videoKey)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to process video to create preview", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessVideo)
		}
		prevKey := fmt.Sprintf("media/%d/%d/preview.webp", channelId, attachmentId)
		if err := e.storage.UploadObject(c.UserContext(), prevKey, bytes.NewReader(prevBuf.Bytes()), "image/webp"); err != nil {
			_ = e.storage.RemoveAttachment(c.UserContext(), videoKey)
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload preview for video type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		f := publicURL
		p := pathJoin(base, prevKey)
		finalURL, previewURL = &f, &p

	default:
		// Unsupported for preview; upload as-is without preview
		fileKey := fmt.Sprintf("media/%d/%d/%s", channelId, attachmentId, origName)
		if err := e.storage.UploadObject(c.UserContext(), fileKey, bytes.NewReader(body), ct); err != nil {
			_ = e.at.RemoveAttachment(c.UserContext(), attachmentId, channelId)
			slog.Error("unable to upload object for unsupported type", slog.String("error", err.Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
		}
		f := pathJoin(base, fileKey)
		finalURL = &f
	}

	// Finalize attachment in DB
	if err := e.at.DoneAttachment(c.UserContext(), attachmentId, channelId, &ct, finalURL, previewURL, heightPtr, widthPtr); err != nil {
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

	return c.SendStatus(fiber.StatusCreated)
}

// Generate a webp preview that fits 350x350 preserving aspect ratio
func ffmpegExtractWebP(path string) (*bytes.Buffer, error) {
	cmd := exec.Command("ffmpeg",
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

// ffprobeDimensions returns width and height using ffprobe for the first video stream
func ffprobeDimensions(path string) (width, height int64, err error) {
	cmd := exec.Command("ffprobe",
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
