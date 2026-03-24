package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/repository/db_client"
)

func main() {
	os.Exit(runApp())
}

func runApp() int {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return 1
	}

	appLogger := logger.NewLogger(
		cfg.Log.EnableDebug,
		cfg.Log.EnableInfo,
		cfg.Log.ErrorOutput,
		cfg.Log.InfoOutput,
		cfg.Log.DebugOutput,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	db, err := db_client.NewDBClient(cfg.DatabaseURL)
	if err != nil {
		appLogger.Errorf("Failed to initialize postgres: %v", err)
		return 1
	}

	if cfg.ScraperConfig.ScrapeLastFMTopArtistsForAllTags {
		err = scrapeLastFMTopArtistsForAllTags(ctx, appLogger, db, cfg.ScraperConfig.LastFMAPIKey)
		if err != nil {
			appLogger.Errorf("Failed to scrape LastFM artists: %v", err)
		}
	}

	if cfg.ScraperConfig.ScrapeMusicBrainzArtistsForAllRegions {
		err = scrapeMusicBrainzArtistsForAllRegions(ctx, appLogger, db)
		if err != nil {
			appLogger.Errorf("Failed to scrape MusicBrainz artists by regions: %v", err)
		}
	}

	return 0
}
