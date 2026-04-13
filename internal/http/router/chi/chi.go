package chi

import (
	"log/slog"
	"net/http"

	"github.com/Sheridanlk/Music-Service/internal/http/handlers/player"
	"github.com/Sheridanlk/Music-Service/internal/http/handlers/track/list"
	"github.com/Sheridanlk/Music-Service/internal/http/handlers/track/stream"
	"github.com/Sheridanlk/Music-Service/internal/http/handlers/track/upload"
	"github.com/Sheridanlk/Music-Service/internal/http/middleware/logger"
	chigo "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Setup(log *slog.Logger, trackUploader upload.TrackUploader, streamer stream.Streamer, lister list.Lister) http.Handler {
	router := chigo.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)

	router.Get("/player", player.New())

	router.Post("/tracks", upload.New(log, trackUploader))
	router.Get("/tracks", list.New(log, lister))

	router.Get("/stream/{id}/{file}", stream.New(log, streamer))

	return router
}
