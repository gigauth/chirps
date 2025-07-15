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

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	type Response struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
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

	response := Response{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}
