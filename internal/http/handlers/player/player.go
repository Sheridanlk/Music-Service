package player

import (
	"embed"
	"net/http"
)

//go:embed player.html
var fs embed.FS

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, err := fs.ReadFile("player.html")
		if err != nil {
			http.Error(w, "player not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(tmp)
	}
}
