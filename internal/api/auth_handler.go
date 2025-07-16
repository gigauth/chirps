package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/joaogiacometti/goserver/internal/auth"
	"github.com/joaogiacometti/goserver/internal/database"
)

type ResponseLogin struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type ResponseRefreshToken struct {
	Token string `json:"token"`
}

func mapUserToResponseLogin(user database.User, token, refreshToken string) ResponseLogin {
	return ResponseLogin{
		ID:           user.ID.String(),
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
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
	)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
		return
	}

	cfg.Db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})

	response := mapUserToResponseLogin(user, token, refreshToken)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg *Api) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Invalid or missing refresh token", http.StatusUnauthorized)
		return
	}

	token, err := cfg.Db.GetRefreshTokenByToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	if token.RevokedAt.Valid {
		http.Error(w, "Refresh token has been revoked", http.StatusUnauthorized)
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Refresh token has expired", http.StatusUnauthorized)
		return
	}

	newToken, err := auth.MakeJWT(token.UserID, cfg.JwtTokenSecret)
	if err != nil {
		http.Error(w, "Failed to create new token", http.StatusInternalServerError)
		return
	}

	response := ResponseRefreshToken{
		Token: newToken,
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg *Api) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Invalid or missing refresh token", http.StatusUnauthorized)
		return
	}

	err = cfg.Db.Revoke(r.Context(), refreshToken)
	if err != nil {
		fmt.Println("Error revoking refresh token:", err)
		http.Error(w, "Failed to revoke refresh token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
