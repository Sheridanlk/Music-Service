package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Sheridanlk/Music-Service/internal/storage"
)

var (
	ErrTrackNotReady = errors.New("track is not ready")
	ErrBadStreamFile = errors.New("bad stream file")
)

type StreamService struct {
	log *slog.Logger

	trackProvider TrackProvider
	mediaProvider MediaProvider
}

type TrackProvider interface {
	GetHLS(ctx context.Context, id int64) (string, string, error)
}

type MediaProvider interface {
	GetObject(ctx context.Context, bucketName, objectName string, byteRange *storage.ByteRange) (io.ReadCloser, string, int64, error)
}

func New(log *slog.Logger, trackProvider TrackProvider, mediaProvider MediaProvider) *StreamService {
	return &StreamService{
		log:           log,
		trackProvider: trackProvider,
		mediaProvider: mediaProvider,
	}
}

func (s *StreamService) GetStreamObject(ctx context.Context, trackID int64, file string, br *storage.ByteRange) (io.ReadCloser, string, int64, error) {
	const op = "stream.GetStreamObject"

	log := s.log.With(
		slog.String("op", op),
		slog.Int64("track_id", trackID),
		slog.String("file", file),
	)

	log.Info("getting file")

	if file == "" || strings.Contains(file, "..") || strings.ContainsAny(file, `\/`) {
		return nil, "", 0, fmt.Errorf("%s: %w", op, ErrBadStreamFile)
	}

	bucket, prefix, err := s.trackProvider.GetHLS(ctx, trackID)
	if err != nil {
		return nil, "", 0, fmt.Errorf("%s: failed to get hls info: %w", op, err)
	}

	if prefix == "" {
		return nil, "", 0, fmt.Errorf("%s: %w", op, ErrTrackNotReady)
	}

	key := prefix + file

	rc, ct, size, err := s.mediaProvider.GetObject(ctx, bucket, key, br)
	if err != nil {
		return nil, "", 0, fmt.Errorf("%s: failed to get object: %w", op, err)
	}

	log.Info("file getted")

	return rc, ct, size, nil
}
