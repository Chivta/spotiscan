package main

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/repository"
	"github.com/chivta/ruscan/internal/shared/models"
	"github.com/chivta/ruscan/internal/shared/queue"
	"github.com/chivta/ruscan/internal/spotify"
)

func main() {
	os.Exit(run())
}

func run() int {
	godotenv.Load("./.env")

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	stdlog.SetOutput(log.Logger)
	stdlog.SetFlags(0)

	redisURL := os.Getenv("REDIS_URL")
	rabbitURL := os.Getenv("RABBITMQ_URL")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if rabbitURL == "" || spotifyClientID == "" || spotifyClientSecret == "" {
		log.Error().Msg("RABBITMQ_URL, SPOTIFY_CLIENT_ID, and SPOTIFY_CLIENT_SECRET are required")
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queueClient, err := queue.NewClient(ctx, rabbitURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize rabbitmq")
		return 1
	}
	defer queueClient.Close()

	// declare the scanner queue so publish calls don't silently drop messages
	if err := queueClient.DeclareQueue(models.ScannerQueueName); err != nil {
		log.Err(err).Msg("failed to declare scanner queue")
		return 1
	}

	redisClient, err := repository.InitializeRedis(ctx, redisURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize redis")
		return 1
	}
	defer redisClient.Close()


	jobRepo := repository.NewJobRepo(redisClient)

	spotifyClient := spotify.NewSpotifyClient(spotifyClientID, spotifyClientSecret)
	worker := spotify.NewSpotifyGatewayWorker(jobRepo, queueClient, spotifyClient)

	log.Info().Msg("spotify-gateway starting")
	if err := worker.Start(ctx); err != nil {
		log.Err(err).Msg("spotify-gateway exited with error")
		return 1
	}
	return 0
}
