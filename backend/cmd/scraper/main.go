package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"


	"github.com/chivta/ruscan/internal/shared/repository"
	"github.com/chivta/ruscan/internal/scraper"
)

func main() {
	os.Exit(runApp())
}

func runApp() int {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := scraper.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config")
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

	if cfg.ScrapeLastFMTopArtistsForAllTags {
		wg.Go(func() {
			err := scraper.ScrapeLastFMTopArtistsForAllTags(ctx, repo, cfg.LastFMAPIKey)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scrape LastFM artists")
			}
		})
	}

	if cfg.ScrapeMusicBrainzArtistsForAllRegions {
		wg.Go(func() {
			err := scraper.ScrapeMusicBrainzArtistsForAllRegions(ctx, repo)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scrape MusicBrainz artists by regions")
			}
		})
	}

	if cfg.ScrapePhonkersDBArtists {
		wg.Go(func() {
			err := scraper.ScrapePhonkersDB(ctx, repo)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scrape PhonkersDB artists")
			}
		})
	}

	wg.Wait()

	repo.LoadRussianArtistsToRedis(ctx)

	return 0
}
