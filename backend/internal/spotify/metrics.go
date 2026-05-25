package spotify

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	SpotifyAPIDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "spotify_gateway_api_duration_seconds",
		Help:    "Spotify API call duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
