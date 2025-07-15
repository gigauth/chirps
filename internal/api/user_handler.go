package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joaogiacometti/goserver/internal/auth"
	"github.com/joaogiacometti/goserver/internal/database"
)

type ResponseLogin struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Token     string    `json:"token"`
}

type ResponseCreateUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RequestUser struct {
	Email            string         `json:"email"`
	Password         string         `json:"password"`
	ExpiresInSeconds *time.Duration `json:"expires_in_seconds"`
}

func mapUserToResponseLogin(user database.User, token string) ResponseLogin {
	return ResponseLogin{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     token,
	}
}

func mapUserToResponseCreateUser(user database.User) ResponseLogin {
	return ResponseLogin{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (cfg *Api) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var request RequestUser

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

	user, err := cfg.Db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := mapUserToResponseCreateUser(user)

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg *Api) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request RequestUser

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := cfg.Db.GetUserByEmail(r.Context(), request.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = auth.CheckPasswordHash(request.Password, user.HashedPassword)
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(
		user.ID,
		cfg.JwtTokenSecret,
		request.ExpiresInSeconds,
	)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	response := mapUserToResponseLogin(user, token)
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
