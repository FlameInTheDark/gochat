package upload

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image/gif"
	"io"
	"strconv"
)

func (p *FFmpegProcessor) convertEmojiVariant(ctx context.Context, source io.Reader, maxDimension int64, sizeLimit int64, animated bool) ([]byte, error) {
	args := []string{"-v", "error", "-y", "-i", "pipe:0"}
	if maxDimension > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", maxDimension, maxDimension))
	}
	if animated {
		args = append(args,
			"-loop", "0",
			"-an",
			"-vsync", "0",
			"-c:v", "libwebp_anim",
			"-f", "webp",
			"-fs", strconv.FormatInt(sizeLimit, 10),
			"-",
		)
	} else {
		args = append(args,
			"-f", "image2pipe",
			"-vcodec", "webp",
			"-fs", strconv.FormatInt(sizeLimit, 10),
			"-",
		)
	}
	payload, err := p.runFFmpeg(ctx, source, args...)
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > sizeLimit {
		return nil, ErrTooLarge
	}
	return payload, nil
}

func DetectAnimated(data []byte) (bool, error) {
	if len(data) < 12 {
		return false, nil
	}
	if isGIF(data) {
		g, err := gif.DecodeAll(bytes.NewReader(data))
		if err != nil {
			return false, err
		}
		return len(g.Image) > 1, nil
	}
	if isPNG(data) {
		return isAPNG(data), nil
	}
	if isWEBP(data) {
		return isAnimatedWEBP(data), nil
	}
	return false, nil
}

func isGIF(data []byte) bool {
	return len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a")
}

func isPNG(data []byte) bool {
	return len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
}

func isWEBP(data []byte) bool {
	return len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

func isAPNG(data []byte) bool {
	for i := 8; i+8 <= len(data); {
		if i+8 > len(data) {
			break
		}
		chunkLen := int(binary.BigEndian.Uint32(data[i : i+4]))
		chunkTypeStart := i + 4
		chunkDataStart := i + 8
		chunkEnd := chunkDataStart + chunkLen
		if chunkEnd+4 > len(data) {
			return false
		}
		if string(data[chunkTypeStart:chunkDataStart]) == "acTL" {
			return true
		}
		i = chunkEnd + 4
	}
	return false
}

func isAnimatedWEBP(data []byte) bool {
	if len(data) < 16 {
		return false
	}
	for i := 12; i+8 <= len(data); {
		chunkType := string(data[i : i+4])
		chunkLen := int(binary.LittleEndian.Uint32(data[i+4 : i+8]))
		chunkDataStart := i + 8
		chunkEnd := chunkDataStart + chunkLen
		if chunkEnd > len(data) {
			return false
		}
		switch chunkType {
		case "ANIM", "ANMF":
			return true
		case "VP8X":
			if chunkLen > 0 && chunkDataStart < len(data) && data[chunkDataStart]&0x02 != 0 {
				return true
			}
		}
		i = chunkEnd
		if chunkLen%2 == 1 {
			i++
		}
	}
	return false
}
