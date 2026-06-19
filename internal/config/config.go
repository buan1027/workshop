package config

import "os"

type Config struct {
	Addr                 string
	DatabaseURL          string
	AdminToken           string
	AuthMode             string
	KeycloakIssuerURL    string
	KeycloakClientID     string
	ResetDatabaseOnStart bool
}

func Load() Config {
	return Config{
		Addr:                 envOrDefault("APP_ADDR", ":3000"),
		DatabaseURL:          envOrDefault("DATABASE_URL", "postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"),
		AdminToken:           os.Getenv("ADMIN_TOKEN"),
		AuthMode:             envOrDefault("AUTH_MODE", "admin-token"),
		KeycloakIssuerURL:    os.Getenv("KEYCLOAK_ISSUER_URL"),
		KeycloakClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ResetDatabaseOnStart: envBoolDefault("RESET_DATABASE_ON_START", true),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envBoolDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value == "true" || value == "1" || value == "yes"
}
