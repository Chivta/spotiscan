package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/config"
	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/repository"
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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	db, err := repository.InitializeDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		return 1
	}
	defer db.Close()

	redis, err := repository.InitializeRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize redis")
		return 1
	}
	defer redis.Close()

	repo := repository.NewArtistRepo(db, redis)

	var wg sync.WaitGroup

	if cfg.ScraperConfig.ScrapeLastFMTopArtistsForAllTags {
		wg.Go(func() {
			err := scrapeLastFMTopArtistsForAllTags(ctx, repo, cfg.ScraperConfig.LastFMAPIKey)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scrape LastFM artists")
			}
		})
	}

	if cfg.ScraperConfig.ScrapeMusicBrainzArtistsForAllRegions {
		wg.Go(func() {
			err := scrapeMusicBrainzArtistsForAllRegions(ctx, repo)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scrape MusicBrainz artists by regions")
			}
		})
	}

	if cfg.ScraperConfig.ScrapePhonkersDBArtists {
		wg.Go(func() {
			err := scrapePhonkersDB(ctx, repo)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scrape PhonkersDB artists")
			}
		})
	}

	wg.Wait()

	repo.LoadRussianArtistsToRedis(ctx)

	return 0
}
