package chi

import (
	"log/slog"
	"net/http"

	"github.com/Sheridanlk/Music-Service/internal/http/handlers/player"
	"github.com/Sheridanlk/Music-Service/internal/http/handlers/track/stream"
	"github.com/Sheridanlk/Music-Service/internal/http/handlers/track/upload"
	chigo "github.com/go-chi/chi/v5"
)

func Setup(log *slog.Logger, trackUploader upload.TrackUploader, streamer stream.Streamer) http.Handler {
	router := chigo.NewRouter()

	router.Get("/player", player.New())

	router.Post("/tracks", upload.New(log, trackUploader))
	router.Get("/stream/{id}/{file}", stream.New(log, streamer))

	return router
}
