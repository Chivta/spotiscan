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
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("No .env file found, reading configuration from environment variables")
	}

	config := &Config{
		DatabaseURL: os.Getenv("DB_URL"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return config, nil
}