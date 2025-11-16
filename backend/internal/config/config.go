package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	FrontendURL string
}

func Load() (*Config, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("No .env file found, reading configuration from environment variables")
	}

	config := &Config{
		DatabaseURL: os.Getenv("DB_URL"),
		FrontendURL: os.Getenv("FRONTEND_URL"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if config.FrontendURL == "" {
		return nil, fmt.Errorf("FRONTEND_URL is required")
	}

	return config, nil
}
