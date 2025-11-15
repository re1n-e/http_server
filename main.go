package main

import (
	"chirpy/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	AuthToken    string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	polka := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		log.Fatalf("failed to connecct to db: %v", err)
	}

	dbQueries := database.New(db)

	const filepathRoot = "."
	const port = "8080"
	cfg := apiConfig{
		db:        dbQueries,
		platform:  platform,
		jwtSecret: secret,
		polkaKey:  polka,
	}
	mux := http.NewServeMux()

	mux.Handle("/app/",
		cfg.middlewareMetricsInc(
			http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))),
		),
	)

	mux.HandleFunc("GET /api/healthz", handleReadinessEndpoint)
	mux.HandleFunc("GET /admin/metrics", cfg.handleWriteHitCount)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)
	mux.HandleFunc("POST /api/users", cfg.createUser)
	mux.HandleFunc("PUT /api/users", cfg.updateUser)
	mux.HandleFunc("POST /api/login", cfg.loginUser)
	mux.HandleFunc("POST /api/chirps", cfg.createChirp)
	mux.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevoke)
	mux.HandleFunc("GET /api/chirps", cfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpById)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.HandleUpdateChirpyRed)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
