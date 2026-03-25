package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string `validate:"required,uri"`
	RedisURL            string `validate:"required,uri"`
	JWTSecret           string `validate:"required,min=32"`
	SpotifyClientID     string `validate:"required,alphanum,min=1"`
	SpotifyClientSecret string `validate:"required,min=1"`
	SecureCookies       bool
	Log                 LogConfig       `json:"log" validate:"required"`
	RateLimit           RateLimitConfig `json:"rate_limit" validate:"required"`
	ScraperConfig       ScraperConfig   `json:"scraper_config" validate:"required"`
}

type ScraperConfig struct {
	ScrapeLastFMTopArtistsForAllTags      bool
	ScrapePhonkersDBArtists               bool
	ScrapeMusicBrainzArtistsForAllRegions bool
	LastFMAPIKey                          string `validate:"required,min=1"`
	LastFMSharedSecret                    string `validate:"min=1"`
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
		JWTSecret:           os.Getenv("JWT_SECRET"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SecureCookies:       func() bool { v, _ := strconv.ParseBool(os.Getenv("SECURE_COOKIES")); return v }(),
		ScraperConfig: ScraperConfig{
			ScrapeLastFMTopArtistsForAllTags:      func() bool { v, _ := strconv.ParseBool(os.Getenv("SCRAPE_LASTFM_TOP_ARTISTS_FOR_ALL_TAGS")); return v }(),
			ScrapePhonkersDBArtists:               func() bool { v, _ := strconv.ParseBool(os.Getenv("SCRAPE_PHONKERS_DB_ARTISTS")); return v }(),
			ScrapeMusicBrainzArtistsForAllRegions: func() bool { v, _ := strconv.ParseBool(os.Getenv("SCRAPE_MUSICBRAINZ_ARTISTS_FOR_ALL_REGIONS")); return v }(),
			LastFMAPIKey:                          os.Getenv("LASTFM_API_KEY"),
			LastFMSharedSecret:                    os.Getenv("LASTFM_SHARED_SECRET"),
		},
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
