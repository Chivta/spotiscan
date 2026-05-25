package main

import (
	"context"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/chivta/ruscan/internal/shared/metrics"
	"github.com/chivta/ruscan/internal/shared/queue"
	"github.com/chivta/ruscan/internal/shared/repository"
	"github.com/chivta/ruscan/internal/spotify"
)

func main() {
	os.Exit(run())
}

func run() int {
	godotenv.Load("./.env")

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Hook(metrics.MetricsHook{Component: "spotify-gateway", Counter: metrics.ErrorsTotalCounter})
	stdlog.SetOutput(log.Logger)
	stdlog.SetFlags(0)

	cfg, err := spotify.LoadConfig()
	if err != nil {
		log.Err(err).Msg("failed to load config")
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queueClient, err := queue.NewClient(ctx, cfg.RabbitMQURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize rabbitmq")
		return 1
	}
	defer queueClient.Close()

	// declare the scanner queue so publish calls don't silently drop messages
	if err := queueClient.DeclareQueueWithDLQ(domain.ScannerQueueName); err != nil {
		log.Err(err).Msg("failed to declare scanner queue")
		return 1
	}

	redisClient, err := repository.InitializeRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Err(err).Msg("failed to initialize redis")
		return 1
	}
	defer redisClient.Close()

	jobRepo := repository.NewJobRepo(redisClient)

	spotifyClient := spotify.NewSpotifyClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	worker := spotify.NewSpotifyGatewayWorker(jobRepo, queueClient, spotifyClient)

	r := gin.New() 
	r.GET("/metrics", gin.WrapH(spotify.MetricsHandler()))
	r.GET("/health", func(c *gin.Context) { c.Status(204) })
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info().Msg("metrics server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("metrics server exited with error")
			errCh <- err
		}
	}()

	go func() {
		log.Info().Msg("spotify-gateway starting")
		if err := worker.Start(ctx); err != nil {
			log.Err(err).Msg("spotify-gateway exited with error")
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down")
	case err := <-errCh:
		log.Err(err).Msg("server error")
		return 1
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("server shutdown error")
		return 1
	}

	return 0
}
