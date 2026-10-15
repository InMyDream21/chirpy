package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inmydream21/chirpy/internal/database"
)

const maxChirpLength = 140

var forbiddenWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func cleanBody(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if forbiddenWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func (cfg *APIConfig) handleAddChirp(w http.ResponseWriter, r *http.Request) {
	var chirp struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if chirp.Body == "" {
		respondError(w, http.StatusBadRequest, "Chirp is empty")
		return
	}
	if len(chirp.Body) > maxChirpLength {
		respondError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	chirp.Body = cleanBody(chirp.Body)
	userId := r.Context().Value("userID")
	if userId == nil {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID, ok := userId.(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	createdChirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   chirp.Body,
		UserID: userID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}
	respondJSON(w, http.StatusCreated, ChirpResponse{
		ID:        createdChirp.ID,
		Body:      createdChirp.Body,
		UserID:    createdChirp.UserID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
	})
}

func (cfg *APIConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		parsedAuthorID = uuid.Nil
	}
	var chirps []database.Chirp
	if parsedAuthorID != uuid.Nil {
		chirps, err = cfg.DB.GetChirpsByAuthor(r.Context(), parsedAuthorID)
	} else {
		chirps, err = cfg.DB.GetChirps(r.Context())
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get chirps")
		return
	}
	sort := r.URL.Query().Get("sort")
	if sort == "desc" {
		sortChirpsDescending(chirps)
	}
	responses := make([]ChirpResponse, len(chirps))
	for i, chirp := range chirps {
		responses[i] = ChirpResponse{
			ID:        chirp.ID,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}
	}
	respondJSON(w, http.StatusOK, responses)
}

func (cfg *APIConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("id")
	if chirpID == "" {
		respondError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}
	chirpIDUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.DB.GetChirp(r.Context(), chirpIDUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}
	respondJSON(w, http.StatusOK, ChirpResponse{
		ID:        chirp.ID,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (cfg *APIConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("id")
	if chirpID == "" {
		respondError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}
	chirpIDUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), chirpIDUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}
	userId := r.Context().Value("userID")
	if userId == nil {
		respondError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID, ok := userId.(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if chirp.UserID != userID {
		respondError(w, http.StatusForbidden, "You are not authorized to delete this chirp")
		return
	}
	err = cfg.DB.DeleteChirp(r.Context(), chirpIDUUID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func sortChirpsDescending(chirps []database.Chirp) {
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
	})
}
