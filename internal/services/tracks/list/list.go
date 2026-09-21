package list

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
)

type TrackLister interface {
	ListTracks(ctx context.Context, status string, count int, offset int) ([]models.TrackListItem, error)
}

type ListService struct {
	log *slog.Logger

	readyTrackLister TrackLister
}

func New(log *slog.Logger, TrackLister TrackLister) *ListService {
	return &ListService{
		log:              log,
		readyTrackLister: TrackLister,
	}
}

func (s *ListService) GetTracksList(ctx context.Context, status string, limit int, offset int) ([]models.TrackListItem, error) {
	const op = "list.GetTracksList"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("getting tracks list")

	tracks, err := s.readyTrackLister.ListTracks(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tracks list: %w", op, err)
	}

	log.Info("tracks geted")

	return tracks, nil
}
