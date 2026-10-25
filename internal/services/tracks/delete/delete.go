package delete

import (
	"context"
	"fmt"
	"log/slog"
)

type TrackDeleter interface {
	DeleteTrack(ctx context.Context, id int64) error
}

type MediaDeleter interface {
	DeleteObject(ctx context.Context, bucketName, objectName string) error
}

type DeleteService struct {
	log *slog.Logger

	trackDeleter TrackDeleter
	mediaDeleter MediaDeleter

	originalBucket string
	hlsBucket      string
}

func New(log *slog.Logger, trackDeleter TrackDeleter, mediaDeleter MediaDeleter, originalBucket, hlsBucket string) *DeleteService {
	return &DeleteService{
		log:            log,
		trackDeleter:   trackDeleter,
		mediaDeleter:   mediaDeleter,
		originalBucket: originalBucket,
		hlsBucket:      hlsBucket,
	}
}

func (d *DeleteService) DeleteTrack(ctx context.Context, id int64) error {
	const op = "delete.DeleteService.DeleteTrack"

	log := d.log.With(
		slog.String("op", op),
		slog.Int64("track_id", id),
	)

	log.Info("deleting track")

	err := d.trackDeleter.DeleteTrack(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete track: %w", op, err)
	}

	log.Info("metadata deleted")

	err = d.mediaDeleter.DeleteObject(ctx, d.hlsBucket, fmt.Sprintf("tracks/%d/", id))
	if err != nil {
		log.Error("failed to delete HLS media", slog.Any("error", err))
	}

	log.Info("HLS media deleted")

	err = d.mediaDeleter.DeleteObject(ctx, d.originalBucket, fmt.Sprintf("tracks/%d/", id))
	if err != nil {
		log.Error("failed to delete HLS media", slog.Any("error", err))

	}

	log.Info("original media deleted")

	return nil
}
