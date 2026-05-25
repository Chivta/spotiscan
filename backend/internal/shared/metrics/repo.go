package metrics

// Defines shared metrics for repos (postgres and redis latencies)

import (
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
