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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/scanner"
	"github.com/chivta/ruscan/internal/shared/metrics"
	"github.com/chivta/ruscan/internal/shared/queue"
	"github.com/chivta/ruscan/internal/shared/repository"
)

func main() {
	os.Exit(run())
}

func run() int {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Hook(metrics.MetricsHook{Component: "scanner", Counter: metrics.ErrorsTotalCounter})
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

	artistRepo := repository.NewArtistRepo(db)
	jobRepo := repository.NewJobRepo(redisClient)
	svc := scanner.NewSpotifyService(artistRepo)

	worker := scanner.NewScannerWorker(queueClient, svc, jobRepo)

	r := gin.New()
	r.GET("/metrics", gin.WrapH(scanner.MetricsHandler()))
	r.GET("/health", func(c *gin.Context) { c.Status(204) })
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info().Msg("metrics server starting on :PORT")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("metrics server exited with error")
			errCh <- err
		}
	}()

	go func() {
		log.Info().Msg("scan-worker starting")
		err = worker.Start(ctx)
		if err != nil {
			log.Err(err).Msg("scan-worker exited with error")
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down server")
	case err := <-errCh:
		log.Err(err).Msg("server error")
		return 1
	}

	<-ctx.Done()
	log.Info().Msg("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("server shutdown error")
		return 1
	}

	return 0
}
