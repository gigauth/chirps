package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joaogiacometti/goserver/internal/api"
	"github.com/joaogiacometti/goserver/internal/database"
	"github.com/joho/godotenv"
)

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

	jwtTokenSecret := os.Getenv("JWT_TOKEN_SECRET")
	if jwtTokenSecret == "" {
		log.Fatal("JWT_TOKEN_SECRET must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot connect with database: %s", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := api.Api{
		Db:             dbQueries,
		Platform:       platform,
		JwtTokenSecret: jwtTokenSecret,
	}

	serverMux := apiCfg.BindRoutes()

	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
