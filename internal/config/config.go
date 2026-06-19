package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
	AdminToken  string
}

func Load() Config {
	return Config{
		Addr:        envOrDefault("APP_ADDR", ":3000"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"),
		AdminToken:  os.Getenv("ADMIN_TOKEN"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
