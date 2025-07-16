package api

import "net/http"

func (apiCfg *Api) BindRoutes() http.Handler {
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleResetHitsCount)
	serveMux.HandleFunc("GET /api/healthz", handleHealth)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleHitsCount)

	serveMux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	serveMux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)

	serveMux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirp)

	serveMux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	return serveMux
}
