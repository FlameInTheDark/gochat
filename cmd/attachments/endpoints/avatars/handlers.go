package avatars

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
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

// Upload
//
//	@Summary		Upload user avatar
//	@Description	Uploads an avatar image. Resizes to max 128x128 and converts to WebP <= 250KB. Finalizes avatar metadata.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			user_id		path		int64	true	"User ID"
//	@Param			avatar_id	path		int64	true	"Avatar ID"
//	@Param			file		body		[]byte	true	"Binary image payload"
//	@Success		201			{string}	string	"Created"
//	@Success		204			{string}	string	"No Content (already uploaded)"
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		404			{string}	string	"Avatar not found"
//	@failure		413			{string}	string	"File too large"
//	@failure		415			{string}	string	"Unsupported Media Type"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/upload/avatars/{user_id}/{avatar_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	// Parse params
	userIdStr := c.Params("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectUserID)
	}
	avatarIdStr := c.Params("avatar_id")
	avatarId, err := strconv.ParseInt(avatarIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectAvatarID)
	}

	// Auth and ownership
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	av, err := e.av.GetAvatar(c.UserContext(), avatarId, userId)
	if err != nil {
		slog.Error("unable to get avatar", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetAvatar)
	}
	if av.UserId != user.Id || user.Id != userId {
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	}
	if av.Done {
		return c.SendStatus(fiber.StatusNoContent)
	}

	// Read body
	// Streaming body reader (fallback to buffered)
	var stream io.Reader
	if r := c.Context().RequestBodyStream(); r != nil {
		stream = r
	} else {
		b := c.Body()
		if len(b) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
		}
		stream = bytes.NewReader(b)
	}

	// Peek to detect content type and enforce basic validation
	peek := make([]byte, 512)
	n, _ := io.ReadFull(stream, peek)
	peek = peek[:n]
	if n == 0 {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	}
	ct := http.DetectContentType(peek)
	if !strings.HasPrefix(ct, "image/") {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, ErrUnsupportedContentType)
	}
	// Recompose full reader with the peeked bytes
	full := io.MultiReader(bytes.NewReader(peek), stream)

	// Convert+resize to WEBP via ffmpeg from stdin; limit output size directly
	buf, convErr := ffmpegToWebPStreamLimited(full, avatarMaxSizeBytes)
	if convErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessImage)
	}
	if buf.Len() > avatarMaxSizeBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	}

	// Upload result
	key := fmt.Sprintf("avatars/%d/%d.webp", userId, avatarId)
	if err := e.storage.UploadObject(c.UserContext(), key, bytes.NewReader(buf.Bytes()), "image/webp"); err != nil {
		_ = e.av.RemoveAvatar(c.UserContext(), avatarId, userId)
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
	}

	// Probe dimensions via public URL
	publicURL := pathJoin(strings.TrimRight(e.s3ExternalURL, "/"), key)
	var w, h int64
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(buf.Bytes())); err == nil {
		w, h = int64(cfg.Width), int64(cfg.Height)
	} else {
		if ww, hh, derr := ffprobeDimensions(publicURL); derr == nil {
			w, h = ww, hh
		} else {
			w, h = 0, 0
		}
	}

	// Finalize
	contentType := "image/webp"
	size := int64(buf.Len())
	if err := e.av.DoneAvatar(c.UserContext(), avatarId, userId, &contentType, &publicURL, &h, &w, &size); err != nil {
		_ = e.av.RemoveAvatar(c.UserContext(), avatarId, userId)
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeAvatar)
	}

	if err := e.usr.SetUserAvatar(c.UserContext(), userId, avatarId); err != nil {
		slog.Error("failed to set user avatar", slog.String("error", err.Error()))
		// Do not fail the upload; activation can be retried separately from the settings page in the app
	}

	go func() {
		ad := dto.AvatarData{URL: publicURL, ContentType: &contentType, Width: &w, Height: &h, Size: size}

		u, uerr := e.usr.GetUserById(c.UserContext(), userId)
		if uerr != nil {
			slog.Error("unable to fetch user for avatar update", slog.String("error", uerr.Error()))
			return
		}

		upd := mqmsg.UpdateUser{User: dto.User{Id: u.Id, Name: u.Name, Discriminator: "", Avatar: &ad}}
		if err := e.mqt.SendUserUpdate(userId, &upd); err != nil {
			slog.Error("unable to send user update event after avatar upload", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusCreated)
}

// ffmpegToWebP converts input bytes into 128x128 webp, best effort
func ffmpegToWebP(in []byte) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-y",
		"-i", "pipe:0",
		"-vf", "scale=128:128:force_original_aspect_ratio=decrease",
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-",
	)
	cmd.Stdin = bytes.NewReader(in)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w: %s", err, stderr.String())
	}
	return &out, nil
}

func ffmpegToWebPLimited(in []byte) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-y",
		"-i", "pipe:0",
		"-vf", "scale=128:128:force_original_aspect_ratio=decrease",
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-fs", fmt.Sprintf("%d", avatarMaxSizeBytes),
		"-",
	)
	cmd.Stdin = bytes.NewReader(in)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w: %s", err, stderr.String())
	}
	return &out, nil
}

// ffmpegToWebPStreamLimited converts stdin stream to webp with max size limit.
func ffmpegToWebPStreamLimited(r io.Reader, sizeLimit int64) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-y",
		"-i", "pipe:0",
		"-vf", "scale=128:128:force_original_aspect_ratio=decrease",
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-fs", fmt.Sprintf("%d", sizeLimit),
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

func pathJoin(base, key string) string {
	if base == "" {
		return key
	}
	b := strings.TrimRight(base, "/")
	k := strings.TrimLeft(key, "/")
	return b + "/" + k
}
