package scanner

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_errors_total",
		Help: "Total number of errors in the API.",
	}, []string{"component"})
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
