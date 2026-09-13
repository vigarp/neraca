package config

import (
	"os"
)

type Config struct {
	Port   string
	DBPath string
	Env    string
}

func Load() *Config {
	port := getEnv("PORT", "8088")
	dbPath := getEnv("DB_PATH", "./data/neraca.db")
	env := getEnv("ENV", "development")

	return &Config{
		Port:   port,
		DBPath: dbPath,
		Env:    env,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
