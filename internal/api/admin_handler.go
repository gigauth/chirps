package api

import (
	"fmt"
	"net/http"
)

func (cfg *Api) handleHitsCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")

	w.WriteHeader((http.StatusOK))
	hits := cfg.FileserverHits.Load()

	html := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, hits)

	w.Write([]byte(html))
}

func (cfg *Api) handleResetHitsCount(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		http.Error(w, "This endpoint is only available in development mode", http.StatusForbidden)
		return
	}

	cfg.Db.Reset(r.Context())
	w.WriteHeader((http.StatusOK))
	cfg.FileserverHits.Store(0)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText((http.StatusOK))))
}
