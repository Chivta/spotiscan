package scanner

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// this includes metrics declared in shared/metrics 
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

