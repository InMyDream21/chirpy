package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/inmydream21/chirpy/internal/auth"
)

type PolkaWebhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *APIConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if apiKey != cfg.PolkaKey {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var webhook PolkaWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if webhook.Event == "" {
		respondError(w, http.StatusBadRequest, "event is required")
		return
	}
	if webhook.Event != "user.upgraded" {
		respondJSON(w, http.StatusNoContent, nil)
		return
	}
	_, err = cfg.DB.UpgradeUser(r.Context(), webhook.Data.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}
