package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string
	RedisURL            string
	FrontendURL         string
	SpotifyClientID     string
	SpotifyClientSecret string
	Log                 LogConfig
}

type LogConfig struct {
	EnableDebug bool   `json:"enable_debug"`
	EnableInfo  bool   `json:"enable_info"`
	ErrorOutput string `json:"error_output"`
	InfoOutput  string `json:"info_output"`
	DebugOutput string `json:"debug_output"`
}

func Load() (*Config, error) {
	err := godotenv.Load("./.env")
	if err != nil {
		fmt.Println("No .env file found, reading configuration from environment variables")
	}

	// Load log config from config.json
	file, err := os.Open("./config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer file.Close()

	var jsonConfig struct {
		Log LogConfig `json:"log"`
	}
	if err := json.NewDecoder(file).Decode(&jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	config := &Config{
		DatabaseURL:         os.Getenv("DB_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		FrontendURL:         os.Getenv("FRONTEND_URL"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		Log:                 jsonConfig.Log,
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
	return config, nil
}
