package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	FrontendURL string
	SpotifyClientID string
	SpotifyClientSecret string
	SpotifyRedirectURI string
}

func Load() (*Config, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("No .env file found, reading configuration from environment variables")
	}

	config := &Config{
		DatabaseURL: os.Getenv("DB_URL"),
		FrontendURL: os.Getenv("FRONTEND_URL"),
		SpotifyClientID: os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI: os.Getenv("SPOTIFY_REDIRECT_URI"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if config.FrontendURL == "" {
		return nil, fmt.Errorf("FRONTEND_URL is required")
	}

	if config.SpotifyClientID == "" {
		log.Println("Warning: SPOTIFY_CLIENT_ID is not set")
	}

	if config.SpotifyClientSecret == "" {
		log.Println("Warning: SPOTIFY_CLIENT_SECRET is not set")
	}

	if config.SpotifyRedirectURI == "" {
		log.Println("Warning: SPOTIFY_REDIRECT_URI is not set")
	}

	return config, nil
}
