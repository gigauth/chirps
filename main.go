package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joaogiacometti/goserver/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot connect with database: %s", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		db:       dbQueries,
		platform: platform,
	}

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))
	serveMux.HandleFunc("GET /api/healthz", handleHealth)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleHitsCount)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleResetHitsCount)
	serveMux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleHitsCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")

	w.WriteHeader((http.StatusOK))
	hits := cfg.fileserverHits.Load()

	html := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, hits)

	w.Write([]byte(html))
}

func (cfg *apiConfig) handleResetHitsCount(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		http.Error(w, "This endpoint is only available in development mode", http.StatusForbidden)
		return
	}

	cfg.db.Reset(r.Context())
	w.WriteHeader((http.StatusOK))
	cfg.fileserverHits.Store(0)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText((http.StatusOK))))
}
