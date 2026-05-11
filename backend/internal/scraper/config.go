package scraper

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL                           string `validate:"required,uri"`
	ScrapeLastFMTopArtistsForAllTags      bool
	ScrapePhonkersDBArtists               bool
	ScrapeMusicBrainzArtistsForAllRegions bool
	LastFMAPIKey                          string
}

func LoadConfig() (*Config, error) {
	godotenv.Load("./.env")

	parseBool := func(key string) bool {
		v, err := strconv.ParseBool(os.Getenv(key))
		if err != nil {
			panic(fmt.Sprintf("failed to parse bool from env var %s: %v", key, err))
		}
		return v
	}

	cfg := &Config{
		DatabaseURL:                           os.Getenv("DATABASE_URL"),
		ScrapeLastFMTopArtistsForAllTags:      parseBool("SCRAPE_LASTFM_TOP_ARTISTS_FOR_ALL_TAGS"),
		ScrapePhonkersDBArtists:               parseBool("SCRAPE_PHONKERS_DB_ARTISTS"),
		ScrapeMusicBrainzArtistsForAllRegions: parseBool("SCRAPE_MUSICBRAINZ_ARTISTS_FOR_ALL_REGIONS"),
		LastFMAPIKey:                          os.Getenv("LASTFM_API_KEY"),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		var msgs []string
		for _, e := range err.(validator.ValidationErrors) {
			msgs = append(msgs, fmt.Sprintf("field '%s' failed validation: %s", e.Field(), e.Tag()))
		}
		return nil, fmt.Errorf("config validation failed: %v", msgs)
	}

	if cfg.ScrapeLastFMTopArtistsForAllTags && cfg.LastFMAPIKey == "" {
		return nil, fmt.Errorf("config validation failed: LASTFM_API_KEY is required when SCRAPE_LASTFM_TOP_ARTISTS_FOR_ALL_TAGS is enabled")
	}

	return cfg, nil
}
