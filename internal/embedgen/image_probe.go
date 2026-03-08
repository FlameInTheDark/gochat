package embedgen

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/embed"
)

const (
	imageProbeMaxBytes         int64 = 128 << 10
	imageProbeRangeHeaderValue       = "bytes=0-131071"
)

func init() {
	image.RegisterFormat("webp", "RIFF????WEBP", decodeUnsupportedWEBP, decodeWEBPConfig)
}

func (g *Generator) populateMissingEmbedMediaDimensions(ctx context.Context, result *embed.Embed) {
	if result == nil {
		return
	}
	g.populateMissingMediaDimensions(ctx, result.Image)
	g.populateMissingMediaDimensions(ctx, result.Thumbnail)
}

func (g *Generator) populateMissingMediaDimensions(ctx context.Context, media *embed.EmbedMedia) {
	if media == nil || media.URL == "" || (media.Width != nil && media.Height != nil) {
		return
	}

	parsedURL, err := url.Parse(strings.TrimSpace(media.URL))
	if err != nil || !isHTTPURL(parsedURL) {
		return
	}
	normalizedURL := parsedURL.String()
	if g.shouldExcludeURL(media.URL, normalizedURL) {
		return
	}

	width, height, contentType, err := g.probeImageDimensions(ctx, normalizedURL)
	if err != nil {
		return
	}
	if media.Width == nil {
		media.Width = width
	}
	if media.Height == nil {
		media.Height = height
	}
	if media.ContentType == "" && contentType != "" {
		media.ContentType = contentType
	}
}

func (g *Generator) probeImageDimensions(ctx context.Context, rawURL string) (*int64, *int64, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, "", err
	}
	request.Header.Set("User-Agent", g.userAgent)
	request.Header.Set("Range", imageProbeRangeHeaderValue)
	request.Header.Set("Accept", "image/*")

	response, err := g.client.Do(request)
	if err != nil {
		return nil, nil, "", err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, "", fmt.Errorf("unexpected status code %d for %s", response.StatusCode, rawURL)
	}

	contentType := normalizeContentType(response.Header.Get("Content-Type"))
	data, err := io.ReadAll(io.LimitReader(response.Body, imageProbeMaxBytes))
	if err != nil {
		return nil, nil, contentType, err
	}
	if len(data) == 0 {
		return nil, nil, contentType, fmt.Errorf("empty image probe response for %s", rawURL)
	}

	widthValue, heightValue, err := decodeImageDimensions(data)
	if err != nil {
		return nil, nil, contentType, err
	}
	return &widthValue, &heightValue, contentType, nil
}

func decodeUnsupportedWEBP(r io.Reader) (image.Image, error) {
	return nil, fmt.Errorf("full webp decode is not supported")
}

func decodeWEBPConfig(r io.Reader) (image.Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return image.Config{}, err
	}
	width, height, err := parseWEBPDimensions(data)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{Width: width, Height: height}, nil
}

func decodeImageDimensions(data []byte) (int64, int64, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	return int64(cfg.Width), int64(cfg.Height), nil
}

func parseWEBPDimensions(data []byte) (int, int, error) {
	if len(data) < 30 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return 0, 0, fmt.Errorf("invalid webp header")
	}

	chunkType := string(data[12:16])
	chunk := data[20:]

	switch chunkType {
	case "VP8 ":
		if len(chunk) < 10 {
			return 0, 0, fmt.Errorf("invalid VP8 chunk")
		}
		if chunk[3] != 0x9d || chunk[4] != 0x01 || chunk[5] != 0x2a {
			return 0, 0, fmt.Errorf("invalid VP8 frame header")
		}
		width := int(binary.LittleEndian.Uint16(chunk[6:8]) & 0x3fff)
		height := int(binary.LittleEndian.Uint16(chunk[8:10]) & 0x3fff)
		return width, height, nil
	case "VP8L":
		if len(chunk) < 5 || chunk[0] != 0x2f {
			return 0, 0, fmt.Errorf("invalid VP8L chunk")
		}
		bits := uint32(chunk[1]) | uint32(chunk[2])<<8 | uint32(chunk[3])<<16 | uint32(chunk[4])<<24
		width := int(bits&0x3fff) + 1
		height := int((bits>>14)&0x3fff) + 1
		return width, height, nil
	case "VP8X":
		if len(chunk) < 10 {
			return 0, 0, fmt.Errorf("invalid VP8X chunk")
		}
		width := 1 + int(uint32(chunk[4])|uint32(chunk[5])<<8|uint32(chunk[6])<<16)
		height := 1 + int(uint32(chunk[7])|uint32(chunk[8])<<8|uint32(chunk[9])<<16)
		return width, height, nil
	default:
		return 0, 0, fmt.Errorf("unsupported webp chunk %q", chunkType)
	}
}
