package api

import (
	"context"
	"net/http"

	"github.com/inmydream21/chirpy/internal/auth"
)

func (cfg *APIConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *APIConfig) middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid bearer token")
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid bearer token")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "userID", userID)))
	})
}
