package upload

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sheridanlk/Music-Service/internal/lib/media"
)

type UploadService struct {
	log *slog.Logger

	trackSaver   TrackProvider
	mediaSaver   MediaSaver
	taskProducer TaskProducer

	originalBucket string
	hlsBucket      string
}

type TrackProvider interface {
	SaveTrack(ctx context.Context, title, originBucket string) (int64, error)
	SetOrginKey(ctx context.Context, id int64, originKey string) error
	SetStatusPending(ctx context.Context, id int64) error
	SetStatusError(ctx context.Context, id int64) error
}

type MediaSaver interface {
	PutObject(ctx context.Context, bucketName, objectName string, r io.Reader, size int64, contentType string) error
}

type TaskProducer interface {
	SendTrackTask(ctx context.Context, trackId string) error
}

func New(log *slog.Logger, trackProvider TrackProvider, mediaSaver MediaSaver, taskProducer TaskProducer, originalBucket, hlsBucket string) *UploadService {
	return &UploadService{
		log:            log,
		trackSaver:     trackProvider,
		mediaSaver:     mediaSaver,
		taskProducer:   taskProducer,
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

	log.Info("starting track upload")

	id, err := s.trackSaver.SaveTrack(ctx, title, s.originalBucket)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to save track: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = s.trackSaver.SetStatusError(ctx, id) != nil
		}
	}()

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
	if err := media.WriteToFile(originalLocal, reader); err != nil {
		return 0, fmt.Errorf("%s: failed to copy original file: %w", op, err)
	}

	file, size, err := media.OpenFile(originalLocal)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to open original file: %w", op, err)
	}
	defer file.Close()

	ct := media.DetectContentType(ext)

	if err := s.mediaSaver.PutObject(ctx, s.originalBucket, originKey, file, size, ct); err != nil {
		return 0, fmt.Errorf("%s: failed to upload original file: %w", op, err)
	}

	log.Info("original file uploaded successfully")

	if err := s.trackSaver.SetStatusPending(ctx, id); err != nil {
		return 0, fmt.Errorf("%s: failed to set track status pending: %w", op, err)
	}

	log.Info("sending track processing task")
	if err := s.taskProducer.SendTrackTask(ctx, fmt.Sprintf("%d", id)); err != nil {
		return 0, fmt.Errorf("%s: failed to send task: %w", op, err)
	}

	log.Info("track processing task sent successfully")

	return id, nil
}
