package main

import (
	"chirpy/internal/database"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserversHits atomic.Int32
	db              *database.Queries
	platform        string
	jwtSecret       string
	polkaKey        string
}

type errResp struct {
	Err string `json:"error"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserversHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleWriteHitCount(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserversHits.Load())

	w.Write([]byte(msg))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserversHits.Store(0)
	cfg.resetUsers(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
