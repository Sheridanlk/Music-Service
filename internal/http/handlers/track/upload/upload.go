package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Sheridanlk/Music-Service/internal/lib/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const maxUploadSize = int64(512 << 20) // 512 MB

type Request struct {
	Title string `json:"title" validate:"max=200"`
}

type Response struct {
	response.Response
	ID     int64  `json:"id"`
	Title  string `json:"title,omitempty"`
	Stream string `json:"stream,omitempty"`
}

type TrackUploader interface {
	UploadTrack(ctx context.Context, title string, filename string, reader io.Reader, size int64) (int64, error)
}

func New(log *slog.Logger, uploader TrackUploader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "uploadHandler"

		log := log.With(
			slog.String("op", op),
		)

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			log.Error("failed to parse multipart form", slog.String("error", err.Error()))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to parse multipart form"))

			return
		}
		defer r.MultipartForm.RemoveAll()

		req := Request{
			Title: r.FormValue("title"),
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", slog.String("error", err.Error()))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		file, hdr, err := r.FormFile("file")
		if errors.Is(err, http.ErrMissingFile) || err != nil {
			log.Error("missing file", slog.String("error", err.Error()))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("missing file"))

			return
		}
		defer file.Close()

		filename := hdr.Filename
		if strings.TrimSpace(filename) == "" {
			filename = "track" + filepath.Ext(hdr.Filename)
		}

		size := hdr.Size

		id, err := uploader.UploadTrack(r.Context(), req.Title, filename, file, size)
		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to upload track"))

			return
		}

		stream := fmt.Sprintf("/stream/%d/index.m3u8", id)

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, Response{
			ID:     id,
			Title:  req.Title,
			Stream: stream,
		})
	}
}
