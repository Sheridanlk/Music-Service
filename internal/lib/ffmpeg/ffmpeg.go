package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

func ToHLS(ctx context.Context, inputPath string, outputDir string, segSeconds int) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	playlist := filepath.Join(outputDir, "index.m3u8")
	segPattern := filepath.Join(outputDir, "seg_%05d.aac")

	args := []string{
		"-y",
		"-i", inputPath,
		"-vn",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-ac", "2",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", segSeconds),
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segPattern,
		playlist,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %w: %s", err, stderr.String())
	}
	return nil
}
