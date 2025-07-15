package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joaogiacometti/goserver/internal/auth"
	"github.com/joaogiacometti/goserver/internal/database"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func mapUserToResponse(user database.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var request Request

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := request.Email
	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := mapUserToResponse(user)

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request Request

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), request.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = auth.CheckPasswordHash(request.Password, user.HashedPassword)
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	response := mapUserToResponse(user)
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
