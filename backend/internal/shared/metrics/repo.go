package metrics

// Defines shared metrics for repos (postgres and redis latencies)

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RedisLatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "redis_latency_seconds",
		Help:    "Latency of Redis operations in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	PostgresLatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "postgres_latency_seconds",
		Help:    "Latency of Postgres operations in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})
)

func PgxTracer() pgx.QueryTracer {
	return pgxTracer{}
}

type pgxTracer struct{}

type pgxStartKey struct{}

func (t pgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, pgxStartKey{}, time.Now())
}

func (t pgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value(pgxStartKey{}).(time.Time)
	if !ok {
		return
	}
	var operation string
	if data.Err != nil {
		operation = "error"
	} else {
		operation = strings.SplitN(data.CommandTag.String(), " ", 2)[0]
		if operation == "" {
			operation = "unknown"
		}
	}
	PostgresLatencyHistogram.WithLabelValues(operation).Observe(time.Since(start).Seconds())
}

func RedisHook() redis.Hook {
	return redisHook{}
}

type redisHook struct{}

func (h redisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		operation := cmd.Name()
		err := next(ctx, cmd)
		RedisLatencyHistogram.WithLabelValues(operation).Observe(time.Since(start).Seconds())
		return err
	}
}

func (h redisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		if err != nil {
			return err
		}
		RedisLatencyHistogram.WithLabelValues("pipeline").Observe(time.Since(start).Seconds())

		return nil
	}
}

func (h redisHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}
