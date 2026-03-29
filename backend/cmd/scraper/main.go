package main

import (
	"context"
	"log"
	"os"
	"sync"
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

	// TODO: remove hardcodes
	appLogger := logger.NewLogger(
		false,
		true,
		"stdout",
		"stdout",
		"stdout",
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	db, err := repository.InitializeDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		return 1
	}
	defer db.Close()

	redis, err := repository.InitializeRedis(ctx, cfg.RedisURL)
	if err != nil {
		appLogger.Errorf("Failed to initialize redis: %v", err)
		return 1
	}
	defer redis.Close()

	repo := repository.NewArtistRepo(appLogger, db, redis)

	var wg sync.WaitGroup

	if cfg.ScraperConfig.ScrapeLastFMTopArtistsForAllTags {
		wg.Go(func() {
			err = scrapeLastFMTopArtistsForAllTags(ctx, appLogger, repo, cfg.ScraperConfig.LastFMAPIKey)
			if err != nil {
				appLogger.Errorf("Failed to scrape LastFM artists: %v", err)
			}
		})
	}

	if cfg.ScraperConfig.ScrapeMusicBrainzArtistsForAllRegions {
		wg.Go(func() {
			err = scrapeMusicBrainzArtistsForAllRegions(ctx, appLogger, repo)
			if err != nil {
				appLogger.Errorf("Failed to scrape MusicBrainz artists by regions: %v", err)
			}
		})
	}

	if cfg.ScraperConfig.ScrapePhonkersDBArtists {
		wg.Go(func() {
			err = scrapePhonkersDB(ctx, appLogger, repo)
			if err != nil {
				appLogger.Errorf("Failed to scrape PhonkersDB artists: %v", err)
			}
		})
	}

	wg.Wait()

	repo.LoadRussianArtistsToRedis(ctx)

	return 0
}
