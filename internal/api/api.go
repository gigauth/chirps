package api

import (
	"net/http"
	"sync/atomic"

	"github.com/joaogiacometti/goserver/internal/database"
	_ "github.com/lib/pq"
)

type Api struct {
	FileserverHits atomic.Int32
	Db             *database.Queries
	Platform       string
	JwtTokenSecret string
	PolkaKey       string
}

func (cfg *Api) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
