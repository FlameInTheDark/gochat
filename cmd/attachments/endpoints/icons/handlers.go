package icons

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
//	@Summary		Upload guild icon
//	@Description	Uploads a guild icon. Resizes to max 128x128 and converts to WebP <= 250KB. Only guild owner can upload. Sets guild icon and emits update.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			icon_id		path		int64	true	"Icon ID"
//	@Param			file		body		[]byte	true	"Binary image payload"
//	@Success		201			{string}	string	"Created"
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		413			{string}	string	"File too large"
//	@failure		415			{string}	string	"Unsupported Media Type"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/upload/icons/{guild_id}/{icon_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	// Parse params
	guildIdStr := c.Params("guild_id")
	guildId, err := strconv.ParseInt(guildIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	iconIdStr := c.Params("icon_id")
	iconId, err := strconv.ParseInt(iconIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectIconID)
	}

	// Auth and ownership check
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	g, err := e.gld.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get guild")
	}
	if g.OwnerId != user.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	}

	// Validate placeholder and ownership via Cassandra row
	ic, err := e.ic.GetIcon(c.UserContext(), iconId, guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get icon")
	}
	if ic.GuildId != guildId {
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	}
	if ic.Done {
		return c.SendStatus(fiber.StatusNoContent)
	}

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

	// Peek to detect content type and basic validation
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
	full := io.MultiReader(bytes.NewReader(peek), stream)

	// Convert to WEBP and limit size
	buf, convErr := ffmpegToWebPStreamLimited(full, iconMaxSizeBytes)
	if convErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessImage)
	}
	if buf.Len() > iconMaxSizeBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	}

	// Upload
	key := fmt.Sprintf("icons/%d/%d.webp", guildId, iconId)
	if err := e.storage.UploadObject(c.UserContext(), key, bytes.NewReader(buf.Bytes()), "image/webp"); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
	}

	// Public URL
	publicURL := pathJoin(strings.TrimRight(e.s3ExternalURL, "/"), key)

	// Try to compute dimensions from the produced WEBP (best-effort)
	var w, h int64
	if cfg, _, derr := image.DecodeConfig(bytes.NewReader(buf.Bytes())); derr == nil {
		w, h = int64(cfg.Width), int64(cfg.Height)
	}
	contentType := "image/webp"
	size := int64(buf.Len())
	if err := e.ic.DoneIcon(c.UserContext(), iconId, guildId, &contentType, &publicURL, &h, &w, &size); err != nil {
		slog.Error("unable to finalize icon metadata", slog.String("error", err.Error()))
	}

	// Set icon on guild (Postgres)
	if err := e.gld.SetGuildIcon(c.UserContext(), guildId, iconId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to set guild icon")
	}

	// Emit WS guild update (best-effort)
	go func() {
		gg, gerr := e.gld.GetGuildById(context.Background(), guildId)
		if gerr != nil {
			slog.Error("unable to fetch guild for icon update", slog.String("error", gerr.Error()))
			return
		}
		icn := dto.Icon{Id: iconId, URL: publicURL, Filesize: size, Width: w, Height: h}
		upd := mqmsg.UpdateGuild{Guild: dto.Guild{Id: gg.Id, Name: gg.Name, Icon: &icn, Owner: gg.OwnerId, Public: gg.Public, Permissions: gg.Permissions}}
		if err := e.mqt.SendGuildUpdate(guildId, &upd); err != nil {
			slog.Error("unable to send guild update after icon upload", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusCreated)
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

func pathJoin(base, key string) string {
	if base == "" {
		return key
	}
	b := strings.TrimRight(base, "/")
	k := strings.TrimLeft(key, "/")
	return b + "/" + k
}
