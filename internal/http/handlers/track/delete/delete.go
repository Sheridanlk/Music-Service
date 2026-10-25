package delete

import (
	"context"
	"errors"
	"strconv"

	"log/slog"
	"net/http"

	"github.com/Sheridanlk/Music-Service/internal/lib/response"
	"github.com/Sheridanlk/Music-Service/internal/logger"
	"github.com/Sheridanlk/Music-Service/internal/storage"
	chigo "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Deleter interface {
	DeleteTrack(ctx context.Context, id int64) error
}

func New(log *slog.Logger, deleter Deleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.track.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("req_id", middleware.GetReqID(r.Context())),
		)

		idStr := chigo.URLParam(r, "id")

		trackID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || trackID <= 0 {
			log.Error("invalif track id", slog.String("id", idStr))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid trcak id"))
			return
		}

		err = deleter.DeleteTrack(r.Context(), trackID)

		if errors.Is(err, storage.ErrTrackNotFound) {
			log.Error("track not found", logger.Err(err), slog.Int64("track_id", trackID))

			w.WriteHeader(http.StatusNotFound)

			return
		}

		if err != nil {
			log.Error("failed to delete track", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
