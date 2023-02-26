package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBHost         string        `envconfig:"DB_HOST" default:"localhost"`
	DBPort         string        `envconfig:"DB_PORT" default:"26257"`
	DBName         string        `envconfig:"DB_NAME" default:"authentication"`
	DBUser         string        `envconfig:"DB_USER" default:"root"`
	DBPassword     string        `envconfig:"DB_PASSWORD" default:""`
	DBSSLMode      string        `envconfig:"DB_SSL_MODE" default:"disable"`
	JWTSecretKey   string        `envconfig:"JWT_SECRET_KEY" default:"123"`
	TokenTTL       time.Duration `envconfig:"TOKEN_TTL" default:"30m"`
	PORT           string        `envconfig:"PORT" default:"8080"`
	MIGRATION_PATH string        `envconfig:"MIGRATION_PATH" default:"/app/migrations"`
}

func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	return &cfg, nil
}
