package upload

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type MediaProcessor interface {
	CreateWebPPreview(ctx context.Context, source string, maxDimension int) ([]byte, error)
	ConvertToWebP(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64) ([]byte, error)
	ProbeDimensions(ctx context.Context, source string) (int64, int64, error)
}

type FFmpegProcessor struct {
	ffmpegPath   string
	ffprobePath  string
	mediaTimeout time.Duration
	probeTimeout time.Duration
}

func NewFFmpegProcessor() *FFmpegProcessor {
	return &FFmpegProcessor{
		ffmpegPath:   "ffmpeg",
		ffprobePath:  "ffprobe",
		mediaTimeout: 20 * time.Second,
		probeTimeout: 10 * time.Second,
	}
}

func (p *FFmpegProcessor) CreateWebPPreview(ctx context.Context, source string, maxDimension int) ([]byte, error) {
	return p.runFFmpeg(ctx, nil,
		"-v", "error",
		"-y",
		"-i", source,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", maxDimension, maxDimension),
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-",
	)
}

func (p *FFmpegProcessor) ConvertToWebP(ctx context.Context, source io.Reader, maxDimension int, sizeLimit int64) ([]byte, error) {
	return p.runFFmpeg(ctx, source,
		"-v", "error",
		"-y",
		"-i", "pipe:0",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", maxDimension, maxDimension),
		"-f", "image2pipe",
		"-vcodec", "webp",
		"-fs", strconv.FormatInt(sizeLimit, 10),
		"-",
	)
}

func (p *FFmpegProcessor) ProbeDimensions(ctx context.Context, source string) (int64, int64, error) {
	cmdCtx, cancel := p.withProbeTimeout(ctx)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, p.ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		source,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("%w: ffprobe failed: %s", ErrMediaProcess, stderr.String())
	}

	s := strings.TrimSpace(out.String())
	if s == "" {
		return 0, 0, fmt.Errorf("%w: ffprobe returned empty output", ErrMediaProcess)
	}
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("%w: unexpected ffprobe output %q", ErrMediaProcess, s)
	}

	width, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: parse width: %v", ErrMediaProcess, err)
	}
	height, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: parse height: %v", ErrMediaProcess, err)
	}
	return width, height, nil
}

func (p *FFmpegProcessor) runFFmpeg(ctx context.Context, input io.Reader, args ...string) ([]byte, error) {
	cmdCtx, cancel := p.withMediaTimeout(ctx)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, p.ffmpegPath, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Stdin = input
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: ffmpeg failed: %s", ErrMediaProcess, stderr.String())
	}
	return out.Bytes(), nil
}

func (p *FFmpegProcessor) withMediaTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, p.mediaTimeout)
}

func (p *FFmpegProcessor) withProbeTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, p.probeTimeout)
}
