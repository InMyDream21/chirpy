package api

import "net/http"

func (cfg *APIConfig) Routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir("static")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /api/healthz", handleHealth)

	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.Handle("POST /api/refresh", http.HandlerFunc(cfg.handleRefresh))
	mux.Handle("POST /api/revoke", http.HandlerFunc(cfg.handleRevokeJWT))
	mux.Handle("PUT /api/users", cfg.middlewareAuth(http.HandlerFunc(cfg.handleUpdateUser)))

	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.getChirp)
	mux.Handle("POST /api/chirps", cfg.middlewareAuth(http.HandlerFunc(cfg.handleAddChirp)))
	mux.Handle("DELETE /api/chirps/{id}", cfg.middlewareAuth(http.HandlerFunc(cfg.handleDeleteChirp)))

	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	mux.Handle("POST /api/polka/webhooks", http.HandlerFunc(cfg.handlePolkaWebhook))

	return mux
}
