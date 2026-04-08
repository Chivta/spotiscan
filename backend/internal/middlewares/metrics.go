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
	registry          *prometheus.Registry
	requestCount      *prometheus.CounterVec
	requestDur        *prometheus.HistogramVec
	userRegistrations prometheus.Counter
	userLogins        prometheus.Counter
	scansTotal        prometheus.Counter
	scanDuration      prometheus.Histogram
	errorsTotal       *prometheus.CounterVec
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

	userRegistrations := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: appName,
		Name:      "user_registrations_total",
		Help:      "Total number of successful user registrations.",
	})

	userLogins := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: appName,
		Name:      "user_logins_total",
		Help:      "Total number of successful user logins.",
	})

	scansTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: appName,
		Name:      "scans_total",
		Help:      "Total number of successful playlist scan completions.",
	})

	scanDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: appName,
		Name:      "scan_duration_seconds",
		Help:      "Full playlist scan duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	})

	errorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appName,
			Name:      "errors_total",
			Help:      "Total number of application errors by type.",
		},
		[]string{"type"},
	)

	registry.MustRegister(
		requestCount, requestDur,
		userRegistrations, userLogins,
		scansTotal, scanDuration,
		errorsTotal,
	)

	return &Metrics{
		registry:          registry,
		requestCount:      requestCount,
		requestDur:        requestDur,
		userRegistrations: userRegistrations,
		userLogins:        userLogins,
		scansTotal:        scansTotal,
		scanDuration:      scanDuration,
		errorsTotal:       errorsTotal,
	}
}

func (m *Metrics) IncUserRegistrations()          { m.userRegistrations.Inc() }
func (m *Metrics) IncUserLogins()                  { m.userLogins.Inc() }
func (m *Metrics) IncScans()                       { m.scansTotal.Inc() }
func (m *Metrics) ObserveScanDuration(s float64)   { m.scanDuration.Observe(s) }
func (m *Metrics) IncErrors(errorType string)      { m.errorsTotal.WithLabelValues(errorType).Inc() }

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
