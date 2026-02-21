package stream

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sheridanlk/Music-Service/internal/lib/response"
	"github.com/Sheridanlk/Music-Service/internal/storage"
	chigo "github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Streamer interface {
	GetStreamObject(ctx context.Context, trackID int64, file string, br *storage.ByteRange) (io.ReadCloser, string, int64, error)
}

func New(log *slog.Logger, streamer Streamer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.track.stream.New"

		log := log.With(
			slog.String("op", op),
		)

		idStr := chigo.URLParam(r, "id")
		file := chigo.URLParam(r, "file")

		trackID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || trackID <= 0 {
			log.Error("invalif track id", slog.String("id", idStr))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid trcak id"))
			return
		}

		var br *storage.ByteRange
		if rng := r.Header.Get("Range"); rng != "" {
			if parsed, ok := parseRangeHeader(rng); ok {
				br = parsed
				w.Header().Set("Accept-Ranges", "bytes")
			}
		}

		rc, ct, _, err := streamer.GetStreamObject(r.Context(), trackID, file, br)
		if err != nil {

			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, response.Error("not found"))

			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", ct)

		if strings.EqualFold(filepath.Ext(file), ".m3u8") {
			w.Header().Set("Cache-Control", "no-store")
		}

		if _, err := io.Copy(w, rc); err != nil {
			log.Info("stream interrupted", slog.String("error", err.Error()))

			return
		}
	}
}

func parseRangeHeader(h string) (*storage.ByteRange, bool) {
	h = strings.TrimSpace(h)
	if !strings.HasPrefix(h, "bytes=") {
		return nil, false
	}
	v := strings.TrimPrefix(h, "bytes=")
	parts := strings.SplitN(v, "-", 2)
	if len(parts) != 2 {
		return nil, false
	}

	start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil || start < 0 {
		return nil, false
	}

	endStr := strings.TrimSpace(parts[1])
	if endStr == "" {
		return &storage.ByteRange{Start: start, End: -1}, true
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil || end < start {
		return nil, false
	}

	return &storage.ByteRange{Start: start, End: end}, true
}
