package upload

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

func init() {
	image.RegisterFormat("webp", "RIFF????WEBP", decodeUnsupportedWEBP, decodeWEBPConfig)
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

func DecodeImageDimensions(data []byte) (int64, int64, error) {
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
