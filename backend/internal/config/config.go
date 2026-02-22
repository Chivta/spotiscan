package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string          `validate:"required,uri"`
	RedisURL            string          `validate:"required,uri"`
	SpotifyClientID     string          `validate:"required,alphanum,min=1"`
	SpotifyClientSecret string          `validate:"required,min=1"`
	Log                 LogConfig       `json:"log" validate:"required"`
	RateLimit           RateLimitConfig `json:"rate_limit" validate:"required"`
}

type LogConfig struct {
	EnableDebug bool   `json:"enable_debug"`
	EnableInfo  bool   `json:"enable_info"`
	ErrorOutput string `json:"error_output"`
	InfoOutput  string `json:"info_output"`
	DebugOutput string `json:"debug_output"`
}

type RateLimitConfig struct {
	RequestLimit  int `json:"request_limit" validate:"required,gt=0"`
	WindowSeconds int `json:"window_seconds" validate:"required,gt=0"`
}

func Load() (*Config, error) {
	godotenv.Load("./.env")
		
	// Load log config from config.json
	file, err := os.Open("./config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer file.Close()

	config := &Config{
		DatabaseURL:         os.Getenv("DB_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	// Validate the configuration
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, fmt.Sprintf("field '%s' failed validation: %s", err.Field(), err.Tag()))
		}
		return nil, fmt.Errorf("configuration validation failed: %v", validationErrors)
	}

	return config, nil
}
