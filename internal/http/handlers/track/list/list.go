package list

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
	"github.com/Sheridanlk/Music-Service/internal/lib/response"
	"github.com/Sheridanlk/Music-Service/internal/logger"
	"github.com/go-chi/render"
)

const (
	defaultOffset = 0
	defaultLimit  = 20
	maxLimit      = 200
	streamBaseURL = "/stream/%d/index.m3u8"
)

type Response struct {
	response.Response
	Items []TrackListResponse `json:"items,omitempty"`
}

type TrackListResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	StreamURL string    `json:"stream_url"`
}

type Lister interface {
	GetTracksList(ctx context.Context, count int, offset int) ([]models.TrackListItem, error)
}

func New(log *slog.Logger, lister Lister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.track.list.New"

		log := log.With(slog.String("op", op))

		limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
		offsetRaw := strings.TrimSpace(r.URL.Query().Get("offset"))

		limit := defaultLimit
		offset := defaultOffset

		if limitRaw != "" {
			n, err := strconv.Atoi(limitRaw)
			if err != nil {
				log.Error("invalid limit, not a number", slog.String("limit", limitRaw))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("invalid limit: must be integer"))

				return
			}
			switch {
			case n <= 0:
				limit = defaultLimit
			case n > maxLimit:
				limit = maxLimit
			default:
				limit = n
			}
		}

		if offsetRaw != "" {
			n, err := strconv.Atoi(offsetRaw)
			if err != nil {
				log.Error("invalid offset, not a number", slog.String("offset", offsetRaw))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("invalid offset: must be integer"))

				return
			}
			if n < defaultOffset {
				offset = defaultOffset
			} else {
				offset = n
			}
		}

		list, err := lister.GetTracksList(r.Context(), limit, offset)
		if err != nil {
			log.Error("failed to get tracks list", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("ailed to get tracks list"))

			return
		}

		respList := MapTracksToResponse(list, streamBaseURL)

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, Response{
			Items: respList,
		})
	}
}

func MapTrackToResponse(t models.TrackListItem, streamBaseURL string) TrackListResponse {
	return TrackListResponse{
		ID:        t.ID,
		Title:     t.Title,
		CreatedAt: t.CreatedAt,
		StreamURL: fmt.Sprintf(streamBaseURL, t.ID),
	}
}

func MapTracksToResponse(tracks []models.TrackListItem, streamBaseURL string) []TrackListResponse {
	result := make([]TrackListResponse, len(tracks))

	for i, t := range tracks {
		result[i] = MapTrackToResponse(t, streamBaseURL)
	}

	return result
}
