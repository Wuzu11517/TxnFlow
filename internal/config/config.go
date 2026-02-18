package config

import "os"

type Config struct {
	DatabaseURL   string
	Port          string
	InfuraAPIKey  string
}

func Load() Config {
	return Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost/txnflow?sslmode=disable"),
		Port:         getEnv("PORT", "8080"),
		InfuraAPIKey: getEnv("INFURA_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
	