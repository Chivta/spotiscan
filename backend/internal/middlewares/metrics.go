package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry     *prometheus.Registry
	requestCount *prometheus.CounterVec
	requestDur   *prometheus.HistogramVec
}

func NewMetrics(appName string) *Metrics {
	registry := prometheus.NewRegistry()

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appName,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	requestDur := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appName,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"method", "path", "status"},
	)

	registry.MustRegister(requestCount, requestDur)

	return &Metrics{
		registry:     registry,
		requestCount: requestCount,
		requestDur:   requestDur,
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) Middleware(ignoredPaths ...string) gin.HandlerFunc {
	ignored := make(map[string]struct{}, len(ignoredPaths))
	for _, p := range ignoredPaths {
		ignored[p] = struct{}{}
	}

	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			c.Next()
			return
		}
		if _, skip := ignored[path]; skip {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		duration := time.Since(start).Seconds()

		m.requestCount.WithLabelValues(method, path, status).Inc()
		m.requestDur.WithLabelValues(method, path, status).Observe(duration)
	}
}
