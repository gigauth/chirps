package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/joaogiacometti/goserver/internal/auth"
)

type RequestPolkaWebook struct {
	Event string `json:"event"`
	Data  struct {
		UserId string `json:"user_id"`
	} `json:"data"`
}

const (
	UserUpgraded = "user.upgraded"
)

func (cfg *Api) handlePolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if apiKey != cfg.PolkaKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request RequestPolkaWebook

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := uuid.MustParse(request.Data.UserId)

	if request.Event != UserUpgraded {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.Db.UpgradeUserToRed(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to upgrade user", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
