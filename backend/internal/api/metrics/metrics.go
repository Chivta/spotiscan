package metrics

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()

	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ruscan",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ruscan",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"method", "path", "status"},
	)

	UserRegistrations = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ruscan",
		Name:      "user_registrations_total",
		Help:      "Total number of successful user registrations.",
	})

	UserLogins = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ruscan",
		Name:      "user_logins_total",
		Help:      "Total number of successful user logins.",
	})

	ScansTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ruscan",
			Name:      "scans_total",
			Help:      "Total number of successful scan completions by type.",
		},
		[]string{"type"},
	)

	ScanDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "ruscan",
		Name:      "scan_duration_seconds",
		Help:      "Scan duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	})

	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ruscan",
			Name:      "errors_total",
			Help:      "Total number of application errors by type.",
		},
		[]string{"type"},
	)

	AnonSessions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ruscan",
		Name:      "anon_sessions_total",
		Help:      "Total number of anonymous sessions created.",
	})

	AnonScansTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ruscan",
			Name:      "anon_scans_total",
			Help:      "Total number of successful scans by anonymous users by type.",
		},
		[]string{"type"},
	)

	AnonScanDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "ruscan",
		Name:      "anon_scan_duration_seconds",
		Help:      "Scan duration in seconds for anonymous users.",
		Buckets:   prometheus.DefBuckets,
	})
)

func init() {
	registry.MustRegister(
		HTTPRequestsTotal, HTTPRequestDuration,
		UserRegistrations, UserLogins,
		ScansTotal, ScanDuration,
		ErrorsTotal,
		AnonSessions, AnonScansTotal, AnonScanDuration,
	)
}

// Handler returns the Prometheus metrics HTTP handler.
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// Middleware records HTTP request counts and durations, skipping ignoredPaths.
func Middleware(ignoredPaths ...string) gin.HandlerFunc {
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

		HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	}
}

// ErrorTypeLabel maps an application error to a label value for ErrorsTotal.
func ErrorTypeLabel(err error) string {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		return "internal"
	}
	switch appErr.Code {
	case "SPOTIFY_API_ERROR":
		return "spotify_api"
	case "DATABASE_ERROR":
		return "db"
	case "UNAUTHORIZED", "INVALID_CREDENTIALS", "FORBIDDEN":
		return "auth"
	case "BAD_REQUEST":
		return "bad_request"
	case "TOO_MANY_REQUESTS":
		return "rate_limit"
	case "PLAYLIST_NOT_FOUND":
		return "playlist_not_found"
	case "EMAIL_EXISTS":
		return "email_exists"
	case "ANON_QUOTA_EXCEEDED":
		return "anon_quota_exceeded"
	default:
		return "internal"
	}
}
