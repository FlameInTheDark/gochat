package emoji

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

const entityName = "emoji"

type entity struct {
	name       string
	publicBase string
}

func (e *entity) Name() string {
	return e.name
}

func New(publicBase string, _ *slog.Logger) server.Entity {
	return &entity{name: entityName, publicBase: publicBase}
}

func (e *entity) Init(router fiber.Router) {
	router.Get("/:emoji_id", e.Redirect)
}

// Redirect
//
//	@Summary	Redirect to public emoji asset
//	@Produce	plain
//	@Tags		Emoji
//	@Param		emoji_id	path		string	true	"Emoji filename ending in .webp"
//	@Param		size		query		int		false	"Preferred rendered size"
//	@Success	307			{string}	string	"Temporary Redirect"
//	@failure	404			{string}	string	"Not found"
//	@Router		/emoji/{emoji_id} [get]
func (e *entity) Redirect(c *fiber.Ctx) error {
	rawID := c.Params("emoji_id")
	if !strings.HasSuffix(rawID, ".webp") {
		return fiber.ErrNotFound
	}
	rawID = strings.TrimSuffix(rawID, ".webp")
	emojiID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || emojiID <= 0 {
		return fiber.ErrNotFound
	}

	size, _ := strconv.Atoi(c.Query("size"))
	variant := emojiutil.SelectClosestVariant(size)
	return c.Redirect(upload.PublicURL(e.publicBase, upload.EmojiVariantKey(emojiID, variant)), fiber.StatusTemporaryRedirect)
}
