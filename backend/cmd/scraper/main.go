package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chivta/spotiscan/internal/config"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repository"
)

func main() {
	os.Exit(runApp())
}

type artistsRepo interface {
	InsertArtists(ctx context.Context, artists []models.Artist) error 
	GetRuTags(ctx context.Context) ([]string, error)
	GetRuRegionIds(ctx context.Context) ([]string, error)
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

	db, err := repository.InitializeDatabase(cfg.DatabaseURL)
	if err != nil {
		return 1
	}
	defer db.Close()

	redis, err := repository.InitializeRedis(cfg.RedisURL)
	if err != nil {
		appLogger.Errorf("Failed to initialize redis (rate limiting and caching disabled): %v", err)
		return 1
	}
	defer redis.Close()

	repo := repository.NewArtistRepo(appLogger, db, redis)

	if cfg.ScraperConfig.ScrapeLastFMTopArtistsForAllTags {
		err = scrapeLastFMTopArtistsForAllTags(ctx, appLogger, repo, cfg.ScraperConfig.LastFMAPIKey)
		if err != nil {
			appLogger.Errorf("Failed to scrape LastFM artists: %v", err)
		}
	}

	if cfg.ScraperConfig.ScrapeMusicBrainzArtistsForAllRegions {
		err = scrapeMusicBrainzArtistsForAllRegions(ctx, appLogger, repo)
		if err != nil {
			appLogger.Errorf("Failed to scrape MusicBrainz artists by regions: %v", err)
		}
	}

	if cfg.ScraperConfig.ScrapePhonkersDBArtists {
		err = scrapePhonkersDB(ctx, appLogger, repo)
		if err != nil {
			appLogger.Errorf("Failed to scrape PhonkersDB artists: %v", err)
		}
	}

	repo.LoadRussianArtistsToRedis(ctx)

	return 0
}
