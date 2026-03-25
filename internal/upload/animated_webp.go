package upload

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	xwebp "golang.org/x/image/webp"
)

type animatedWEBPFrame struct {
	canvasWidth  int
	canvasHeight int
	background   color.NRGBA
	x            int
	y            int
	noBlend      bool
	webp         []byte
}

func RenderAnimatedWEBPFirstFramePNG(data []byte) ([]byte, int64, int64, error) {
	frame, err := parseAnimatedWEBPFirstFrame(data)
	if err != nil {
		return nil, 0, 0, err
	}

	img, err := xwebp.Decode(bytes.NewReader(frame.webp))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("%w: decode first animated webp frame: %v", ErrMediaProcess, err)
	}

	canvasWidth := frame.canvasWidth
	canvasHeight := frame.canvasHeight
	if canvasWidth <= 0 {
		canvasWidth = frame.x + img.Bounds().Dx()
	}
	if canvasHeight <= 0 {
		canvasHeight = frame.y + img.Bounds().Dy()
	}
	if canvasWidth <= 0 || canvasHeight <= 0 {
		return nil, 0, 0, fmt.Errorf("%w: invalid animated webp canvas", ErrMediaProcess)
	}

	canvas := image.NewNRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: frame.background}, image.Point{}, draw.Src)

	target := image.Rect(frame.x, frame.y, frame.x+img.Bounds().Dx(), frame.y+img.Bounds().Dy())
	op := draw.Over
	if frame.noBlend {
		op = draw.Src
	}
	draw.Draw(canvas, target, img, img.Bounds().Min, op)

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		return nil, 0, 0, fmt.Errorf("%w: encode animated webp frame png: %v", ErrMediaProcess, err)
	}
	return encoded.Bytes(), int64(canvasWidth), int64(canvasHeight), nil
}

func parseAnimatedWEBPFirstFrame(data []byte) (*animatedWEBPFrame, error) {
	if !isWEBP(data) {
		return nil, fmt.Errorf("%w: not a webp file", ErrMediaProcess)
	}

	frame := &animatedWEBPFrame{}
	for i := 12; i+8 <= len(data); {
		chunkType := string(data[i : i+4])
		chunkLen := int(binary.LittleEndian.Uint32(data[i+4 : i+8]))
		chunkDataStart := i + 8
		chunkEnd := chunkDataStart + chunkLen
		if chunkEnd > len(data) {
			return nil, fmt.Errorf("%w: truncated webp chunk %q", ErrMediaProcess, chunkType)
		}

		switch chunkType {
		case "VP8X":
			if chunkLen < 10 {
				return nil, fmt.Errorf("%w: invalid VP8X chunk", ErrMediaProcess)
			}
			frame.canvasWidth = 1 + int(uint32(data[chunkDataStart+4])|uint32(data[chunkDataStart+5])<<8|uint32(data[chunkDataStart+6])<<16)
			frame.canvasHeight = 1 + int(uint32(data[chunkDataStart+7])|uint32(data[chunkDataStart+8])<<8|uint32(data[chunkDataStart+9])<<16)
		case "ANIM":
			if chunkLen < 6 {
				return nil, fmt.Errorf("%w: invalid ANIM chunk", ErrMediaProcess)
			}
			frame.background = color.NRGBA{
				R: data[chunkDataStart+2],
				G: data[chunkDataStart+1],
				B: data[chunkDataStart],
				A: data[chunkDataStart+3],
			}
		case "ANMF":
			return parseAnimatedWEBPFrameChunk(data[chunkDataStart:chunkEnd], frame)
		}

		i = chunkEnd
		if chunkLen%2 == 1 {
			i++
		}
	}

	return nil, fmt.Errorf("%w: animated webp frame chunk not found", ErrMediaProcess)
}

func parseAnimatedWEBPFrameChunk(payload []byte, base *animatedWEBPFrame) (*animatedWEBPFrame, error) {
	if len(payload) < 16 {
		return nil, fmt.Errorf("%w: invalid ANMF chunk", ErrMediaProcess)
	}

	frame := *base
	frame.x = 2 * int(readUint24(payload[0:3]))
	frame.y = 2 * int(readUint24(payload[3:6]))
	frameWidth := 1 + int(readUint24(payload[6:9]))
	frameHeight := 1 + int(readUint24(payload[9:12]))
	frame.noBlend = payload[15]&0x02 != 0

	alphaChunk, bitstreamChunk, err := extractAnimatedWEBPBitstreamChunks(payload[16:])
	if err != nil {
		return nil, err
	}
	frame.webp = buildStillWEBP(frameWidth, frameHeight, alphaChunk, bitstreamChunk)
	if frame.canvasWidth == 0 {
		frame.canvasWidth = frame.x + frameWidth
	}
	if frame.canvasHeight == 0 {
		frame.canvasHeight = frame.y + frameHeight
	}
	return &frame, nil
}

func extractAnimatedWEBPBitstreamChunks(payload []byte) ([]byte, []byte, error) {
	var alphaChunk []byte
	var bitstreamChunk []byte

	for i := 0; i+8 <= len(payload); {
		chunkType := string(payload[i : i+4])
		chunkLen := int(binary.LittleEndian.Uint32(payload[i+4 : i+8]))
		chunkEnd := i + 8 + chunkLen
		if chunkEnd > len(payload) {
			return nil, nil, fmt.Errorf("%w: truncated animated webp frame subchunk %q", ErrMediaProcess, chunkType)
		}

		chunkBytes := append([]byte(nil), payload[i:chunkEnd]...)
		if chunkLen%2 == 1 {
			chunkBytes = append(chunkBytes, 0)
		}

		switch chunkType {
		case "ALPH":
			if alphaChunk == nil {
				alphaChunk = chunkBytes
			}
		case "VP8 ", "VP8L":
			bitstreamChunk = chunkBytes
			return alphaChunk, bitstreamChunk, nil
		}

		i = chunkEnd
		if chunkLen%2 == 1 {
			i++
		}
	}

	return nil, nil, fmt.Errorf("%w: animated webp frame image data not found", ErrMediaProcess)
}

func buildStillWEBP(width, height int, alphaChunk, bitstreamChunk []byte) []byte {
	body := make([]byte, 0, len(alphaChunk)+len(bitstreamChunk)+18)

	vp8xPayload := make([]byte, 10)
	if len(alphaChunk) > 0 {
		vp8xPayload[0] = 0x10
	}
	writeUint24(vp8xPayload[4:7], width-1)
	writeUint24(vp8xPayload[7:10], height-1)
	body = append(body, makeRIFFChunk("VP8X", vp8xPayload)...)

	if len(alphaChunk) > 0 {
		body = append(body, alphaChunk...)
	}
	body = append(body, bitstreamChunk...)

	out := make([]byte, 12, 12+len(body))
	copy(out[:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(out[4:8], uint32(4+len(body)))
	copy(out[8:12], []byte("WEBP"))
	out = append(out, body...)
	return out
}

func makeRIFFChunk(tag string, payload []byte) []byte {
	out := make([]byte, 8, 8+len(payload)+1)
	copy(out[:4], []byte(tag))
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(payload)))
	out = append(out, payload...)
	if len(payload)%2 == 1 {
		out = append(out, 0)
	}
	return out
}

func readUint24(data []byte) uint32 {
	return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
}

func writeUint24(dst []byte, value int) {
	dst[0] = byte(value)
	dst[1] = byte(value >> 8)
	dst[2] = byte(value >> 16)
}
