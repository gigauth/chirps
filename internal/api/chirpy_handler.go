package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joaogiacometti/goserver/internal/auth"
	"github.com/joaogiacometti/goserver/internal/database"
)

type ResponseChrip struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserID    string `json:"user_id"`
}

type RequestChirpy struct {
	Body string `json:"body"`
}

func MapChirpToResponse(chirp database.Chirp) ResponseChrip {
	return ResponseChrip{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	}
}

func (cfg *Api) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	var request RequestChirpy

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JwtTokenSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Body) > 140 {
		http.Error(w, "Chirp body exceeds 140 characters", http.StatusBadRequest)
		w.WriteHeader(400)
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

	cleareDBody := strings.Join(words, " ")

	chirp, err := cfg.Db.CreateChrip(r.Context(), database.CreateChripParams{
		Body:   cleareDBody,
		UserID: userID,
	})
	if err != nil {
		http.Error(w, "Failed to create chirp", http.StatusInternalServerError)
		return
	}

	response := MapChirpToResponse(chirp)

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (cfg *Api) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.Db.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve chirps", http.StatusInternalServerError)
		return
	}

	var response []ResponseChrip
	for _, chirp := range chirps {
		response = append(response, MapChirpToResponse(chirp))
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg *Api) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.Db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		http.Error(w, "Failed to retrieve chirp", http.StatusNotFound)
		return
	}

	response := MapChirpToResponse(chirp)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg *Api) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.JwtTokenSecret)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chirp, err := cfg.Db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		http.Error(w, "Failed to retrieve chirp", http.StatusNotFound)
		return
	}

	if chirp.UserID != userID {
		http.Error(w, "You can only delete your own chirps", http.StatusForbidden)
		return
	}

	err = cfg.Db.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		http.Error(w, "Failed to delete chirp", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
