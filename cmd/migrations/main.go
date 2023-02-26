package main

import (
	"fmt"
	"os"

	"github.com/cvele/authentication-service/internal/config"
	_ "github.com/cvele/authentication-service/migrations"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

func main() {

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Connect to the database
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := goose.OpenDBWithDriver("postgres", connStr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Obtain the directory where the migration files are located
	migrationsDir := os.Getenv("MIGRATIONS_DIR")

	// Obtain the location of the custom binary for the Goose migrations
	binaryPath := os.Getenv("GOOSE_CUSTOM_BINARY")

	// Run the migrations using the custom binary
	if err := goose.Run("up", db, migrationsDir, binaryPath); err != nil {
		log.Fatal().Err(err).Msg("failed to apply migrations")
	}

	log.Info().Msg("Migrations applied successfully!")
}
