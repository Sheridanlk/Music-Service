package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
)

type TrackLister interface {
	ListTracks(ctx context.Context, count int, offset int) ([]models.TrackListItem, error)
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

func (s *ListService) GetTracksList(ctx context.Context, count int, offset int) ([]models.TrackListItem, error) {
	const op = "list.GetTracksList"

	log := s.log.With(
		slog.String("op", op),
	)

	if count <= 0 {
		count = 20
	}
	if count > 200 {
		count = 200
	}
	if offset < 0 {
		offset = 0
	}

	log.Info("getting tracks list")

	tracks, err := s.trackLister.ListTracks(ctx, count, offset)
	if err != nil {
		log.Error("failed to get tracks list", slog.String("error", err.Error()))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tracks geted")

	return tracks, nil
}
