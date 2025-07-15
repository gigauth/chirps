package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

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
	serveMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleHitsCount)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleResetHitsCount)
	serveMux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)

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

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	type Request struct {
		Email string `json:"email"`
	}

	type UserResponse struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	var request Request

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := request.Email

	user, err := cfg.db.CreateUser(r.Context(), email)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	err = json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusCreated)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Body string `json:"body"`
	}
	type Err struct {
		Error string `json:"error"`
	}
	type Result struct {
		ClanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	request := Request{}
	errReturn := Err{}

	err := decoder.Decode(&request)
	if err != nil {
		errReturn.Error = "Something went wrong"
		errEncoded, _ := json.Marshal(errReturn)

		w.WriteHeader(400)
		w.Write([]byte(errEncoded))
		return
	}

	if len(request.Body) > 140 {
		result, _ := json.Marshal(Err{Error: "Chirp is too long"})
		w.WriteHeader(400)
		w.Write([]byte(result))
		return
	}

	profanedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(request.Body)
	for i, word := range words {
		if _, ok := profanedWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	result := Result{
		ClanedBody: strings.Join(words, " "),
	}

	resultEncoded, _ := json.Marshal(result)

	w.WriteHeader(200)
	w.Write([]byte(resultEncoded))
}
