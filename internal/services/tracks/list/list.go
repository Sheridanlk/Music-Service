package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
)

type TrackLister interface {
	ListReadyTracks(ctx context.Context, count int, offset int) ([]models.TrackListItem, error)
}

type ListService struct {
	log *slog.Logger

	trackLister TrackLister
}

func New(log *slog.Logger, trackLister TrackLister) *ListService {
	return &ListService{
		log:         log,
		trackLister: trackLister,
	}
}

func (s *ListService) GetTracksList(ctx context.Context, limit int, offset int) ([]models.TrackListItem, error) {
	const op = "list.GetTracksList"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("getting tracks list")

	tracks, err := s.trackLister.ListReadyTracks(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tracks list: %w", op, err)
	}

	log.Info("tracks geted")

	return tracks, nil
}
