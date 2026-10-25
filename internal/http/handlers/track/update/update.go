package update

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Sheridanlk/Music-Service/internal/domain/models"
	"github.com/Sheridanlk/Music-Service/internal/lib/response"
	"github.com/Sheridanlk/Music-Service/internal/logger"
	"github.com/Sheridanlk/Music-Service/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type UpdateTrackRequest struct {
	Title          *string `json:"title" validate:"omitempty,min=1,max=255"`
	OriginalAuthor *string `json:"original_author" validate:"omitempty,min=1,max=255"`
}

type Updater interface {
	UpdateTrack(ctx context.Context, id int64, updateParams models.UpdateTrackParams) error
}

func New(log *slog.Logger, updater Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.track.update.New"

		log := log.With(
			slog.String("op", op),
			slog.String("req_id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")

		trackID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || trackID <= 0 {
			log.Error("invalif track id", slog.String("id", idStr))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid trcak id"))

			return
		}

		var req UpdateTrackRequest

		err = render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", logger.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", logger.Err(err))

			validateErr := err.(validator.ValidationErrors)

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		updateParams := models.UpdateTrackParams{
			Title:          req.Title,
			OriginalAuthor: req.OriginalAuthor,
		}

		err = updater.UpdateTrack(r.Context(), trackID, updateParams)
		if errors.Is(err, storage.ErrTrackNotFound) {
			log.Error("track not found", logger.Err(err), slog.Int64("track_id", trackID))

			w.WriteHeader(http.StatusNotFound)

			return
		}
		if err != nil {
			log.Error("failed to update track", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
