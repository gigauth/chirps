package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joaogiacometti/goserver/internal/database"
)

type ResponseChrip struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserID    string `json:"user_id"`
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

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	var request Request

	err := json.NewDecoder(r.Body).Decode(&request)
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

	clearedBody := strings.Join(words, " ")
	uuid := uuid.MustParse(request.UserID)

	chirp, err := cfg.db.CreateChrip(r.Context(), database.CreateChripParams{
		Body:   clearedBody,
		UserID: uuid,
	})
	if err != nil {
		fmt.Printf("Error creating chirp: %v\n", err)
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

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAll(r.Context())
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

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
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
