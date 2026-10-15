package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/inmydream21/chirpy/internal/api"
	"github.com/inmydream21/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env loaded: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	cfg := &api.APIConfig{
		DB:        database.New(db),
		Platform:  os.Getenv("PLATFORM"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		PolkaKey:  os.Getenv("POLKA_KEY"),
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: cfg.Routes(),
	}
	log.Printf("listening on %s", server.Addr)
	return server.ListenAndServe()
}
