package update

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
)

type TrackUpdater interface {
	EditTrack(ctx context.Context, id int64, updateParams models.UpdateTrackParams) error
}

type UpdateService struct {
	log *slog.Logger

	trackEditor TrackUpdater
}

func New(log *slog.Logger, trackEditor TrackUpdater) *UpdateService {
	return &UpdateService{
		log:         log,
		trackEditor: trackEditor,
	}
}

func (s *UpdateService) UpdateTrack(ctx context.Context, id int64, updateParams models.UpdateTrackParams) error {
	const op = "service.update.UpdateTrack"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("updating track")

	err := s.trackEditor.EditTrack(ctx, id, updateParams)

	if err != nil {
		return fmt.Errorf("%s: failed to update track: %w", op, err)
	}

	log.Info("track updated")

	return nil
}
