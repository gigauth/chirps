package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joaogiacometti/goserver/internal/auth"
	"github.com/joaogiacometti/goserver/internal/database"
)

type ResponseCreateUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RequestUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
