package upload

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sheridanlk/Music-Service/internal/lib/ffmpeg"
)

type UploadService struct {
	log *slog.Logger

	trackSaver TrackSaver
	mediaSaver MediaSaver

	originalBucket string
	hlsBucket      string
}

type TrackSaver interface {
	SaveTrack(ctx context.Context, title, originBucket string) (int64, error)
	SetOrginKey(ctx context.Context, id int64, originKey string) error
	SetHLS(ctx context.Context, id int64, hlsBucket string, hlsPrefix string) error
}

type MediaSaver interface {
	PutObject(ctx context.Context, bucketName, objectName string, r io.Reader, size int64, contentType string) error
}

func New(log *slog.Logger, trackSaver TrackSaver, mediaSaver MediaSaver, originalBucket, hlsBucket string) *UploadService {
	return &UploadService{
		log:            log,
		trackSaver:     trackSaver,
		mediaSaver:     mediaSaver,
		originalBucket: originalBucket,
		hlsBucket:      hlsBucket,
	}
}

func (s *UploadService) UploadTrack(ctx context.Context, title string, filename string, reader io.Reader, size int64) (int64, error) {
	const op = "tracks.UploadTrack"

	log := s.log.With(
		slog.String("op", op),
		slog.String("filename", filename),
	)

	title = strings.TrimSpace(title)
	if title == "" {
		title = filename
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}

	log.Info("uloading track")

	id, err := s.trackSaver.SaveTrack(ctx, title, s.originalBucket)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to save track: %w", op, err)
	}

	originKey := fmt.Sprintf("tracks/%d/source/original%s", id, ext)
	if err := s.trackSaver.SetOrginKey(ctx, id, originKey); err != nil {
		return 0, fmt.Errorf("%s: failed to save origin key: %w", op, err)
	}

	tmpDir, err := os.MkdirTemp("", "track-*")
	if err != nil {
		return 0, fmt.Errorf("%s: failed to create temp dir: %w", op, err)
	}
	defer os.RemoveAll(tmpDir)

	originalLocal := filepath.Join(tmpDir, "original"+ext)
	if err := writeToFile(originalLocal, reader); err != nil {
		return 0, fmt.Errorf("%s: failed to copy original file: %w", op, err)
	}

	if err := putFile(ctx, s.mediaSaver, s.originalBucket, originKey, originalLocal, detectContentType(ext)); err != nil {
		return 0, fmt.Errorf("%s: failed to upload original file: %w", op, err)
	}

	log.Info("original file uploaded successfully")
	log.Info("starting HLS conversion")

	hlsLocal := filepath.Join(tmpDir, "hls")
	if err := os.Mkdir(hlsLocal, 0755); err != nil {
		return 0, fmt.Errorf("%s: failed to create local hls dir: %w", op, err)
	}

	if err := ffmpeg.ToHLS(ctx, originalLocal, hlsLocal, 4); err != nil {
		return 0, fmt.Errorf("%s: failed to convert to hls: %w", op, err)
	}

	log.Info("HLS conversion completed successfully")
	log.Info("uploading HLS files")

	hlsPrefix := fmt.Sprintf("tracks/%d/hls/aac_128/", id)
	if err := putDir(ctx, s.mediaSaver, s.hlsBucket, hlsPrefix, hlsLocal); err != nil {
		return 0, fmt.Errorf("%s: failed to upload hls files: %w", op, err)

	}

	if err := s.trackSaver.SetHLS(ctx, id, s.hlsBucket, hlsPrefix); err != nil {
		return 0, fmt.Errorf("%s: failed to save hls info: %w", op, err)
	}

	log.Info("HLS files uploaded successfully")
	log.Info("the track has been loaded")

	return id, nil
}

func writeToFile(path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

func putFile(ctx context.Context, ms MediaSaver, bucketName, objectName, filePath, ct string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	return ms.PutObject(ctx, bucketName, objectName, f, stat.Size(), ct)
}

func putDir(ctx context.Context, ms MediaSaver, bucketName, prefix, dirPath string) error {
	enttries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range enttries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		local := filepath.Join(dirPath, name)
		objectName := prefix + name
		ct := detectContentType(filepath.Ext(name))

		if err := putFile(ctx, ms, bucketName, objectName, local, ct); err != nil {
			return err
		}
	}
	return nil
}

func detectContentType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".aac":
		return "audio/aac"
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	default:
		return "application/octet-stream"
	}
}
