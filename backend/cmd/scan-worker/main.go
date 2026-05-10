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

	cfg, err := scanner.LoadConfig()
	if err != nil {
		log.Err(err).Msg("failed to load config")
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := repository.InitializeDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize database")
		return 1
	}
	defer db.Close()

	redisClient, err := repository.InitializeRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize redis")
		return 1
	}
	defer redisClient.Close()

	queueClient, err := queue.NewClient(ctx, cfg.RabbitMQURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize rabbitmq")
		return 1
	}
	defer queueClient.Close()

	artistRepo := repository.NewArtistRepo(db, redisClient)
	jobRepo := repository.NewJobRepo(redisClient)
	svc := scanner.NewSpotifyService(artistRepo)

	worker := scanner.NewScannerWorker(queueClient, svc, jobRepo)
	log.Info().Msg("scan-worker starting")

	err = worker.Start(ctx)
	if err != nil {
		log.Err(err).Msg("scan-worker exited with error")
		return 1
	}
	return 0
}
