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

	"github.com/chivta/ruscan/internal/shared/queue"
	"github.com/chivta/ruscan/internal/shared/repository"
	"github.com/chivta/ruscan/internal/scanner"
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

	dbURL := os.Getenv("DB_URL")
	redisURL := os.Getenv("REDIS_URL")
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if dbURL == "" || redisURL == "" || rabbitURL == "" {
		log.Error().Msg("DB_URL, REDIS_URL, and RABBITMQ_URL are required")
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := repository.InitializeDatabase(ctx, dbURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize database")
		return 1
	}
	defer db.Close()

	redisClient, err := repository.InitializeRedis(ctx, redisURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize redis")
		return 1
	}
	defer redisClient.Close()

	queueClient, err := queue.NewClient(ctx, rabbitURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize rabbitmq")
		return 1
	}
	defer queueClient.Close()

	artistRepo := repository.NewArtistRepo(db, redisClient)
	jobRepo := repository.NewJobRepo(redisClient)
	svc := scanner.NewSpotifyService(artistRepo, jobRepo)

	worker := scanner.NewScannerWorker(queueClient, svc)
	log.Info().Msg("scan-worker starting")

	err = worker.Start(ctx)
	if err != nil {
		log.Err(err).Msg("scan-worker exited with error")
		return 1
	}
	return 0
}
