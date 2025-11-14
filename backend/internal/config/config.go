package config

import (
	"os"
	"fmt"
	"github.com/joho/godotenv"
)

type Config struct{
	DatabaseURL string

}

func Load() (*Config, error){
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		DatabaseURL: os.Getenv("DB_URL"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return config, nil
}