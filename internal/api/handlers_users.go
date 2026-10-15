package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/inmydream21/chirpy/internal/auth"
	"github.com/inmydream21/chirpy/internal/database"
)

type userResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type loginResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type userRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *APIConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body userRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if body.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}
	if body.Password == "" {
		respondError(w, http.StatusBadRequest, "password is required")
		return
	}

	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("error creating user: %v", err)
		respondError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondJSON(w, http.StatusCreated, userResponse{
		ID:          user.ID,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

func (cfg *APIConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if body.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}
	if body.Password == "" {
		respondError(w, http.StatusBadRequest, "password is required")
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "Couldn't get user by email")
		return
	}
	valid, err := auth.CheckPasswordHash(user.HashedPassword, body.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if !valid {
		respondError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if body.ExpiresInSeconds == 0 || body.ExpiresInSeconds > 3600 {
		body.ExpiresInSeconds = 3600
	}
	expiresIn := time.Duration(body.ExpiresInSeconds) * time.Second
	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, expiresIn)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Couldn't create token")
		return
	}
	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}
	respondJSON(w, http.StatusOK, loginResponse{
		ID:           user.ID,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	})
}

func (cfg *APIConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "No refresh token")
		return
	}
	refreshTokenData, err := cfg.DB.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}
	if refreshTokenData.RevokedAt != (sql.NullTime{}) || refreshTokenData.ExpiresAt.Before(time.Now()) {
		respondError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}
	token, err := auth.MakeJWT(refreshTokenData.UserID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Couldn't create token")
		return
	}
	respondJSON(w, http.StatusOK, loginResponse{
		Token: token,
	})
}

func (cfg *APIConfig) handleRevokeJWT(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "No refresh token")
		return
	}
	err = cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (cfg *APIConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var body userRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if body.Email == "" || body.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
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

	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}
	updatedUser, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}
	respondJSON(w, http.StatusOK, userResponse{
		ID:          updatedUser.ID,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
	})
}
