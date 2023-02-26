package main

import (
	"net/http"
	"os"

	"github.com/cvele/authentication-service/internal/authentication"
	"github.com/cvele/authentication-service/internal/config"
	"github.com/cvele/authentication-service/internal/db"
	"github.com/cvele/authentication-service/internal/router"
	"github.com/rs/zerolog/log"
)

func main() {

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	db, err := db.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// Create API
	api, err := authentication.NewAPI(cfg, *db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create API")
	}

	// Create router and endpoints
	r := router.New()

	r.HandleFunc("/login", api.LoginHandler).Methods("POST")
	r.HandleFunc("/refresh", api.RefreshHandler).Methods("POST")
	r.HandleFunc("/validate", api.ValidateHandler).Methods("POST")
	r.HandleFunc("/register", api.RegisterHandler).Methods("POST")
	r.HandleFunc("/change-password", api.ChangePasswordHandler).Methods("POST")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create server
	srv := &http.Server{
		Addr:    ":" + cfg.PORT,
		Handler: r.Router(),
	}

	// Start server
	log.Info().Msg("starting server...")
	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("server stopped")
		os.Exit(1)
	}
}
