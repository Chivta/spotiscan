package metrics

// Defines shared metrics and hooks for zerolog logging, triggered on log.Error() calls

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

var (
	ErrorsTotalCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total number of errors in components.",
	}, []string{"component"})
)

type MetricsHook struct {
	Component string
	Counter   *prometheus.CounterVec
}

func (h MetricsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.ErrorLevel {
		h.Counter.WithLabelValues(h.Component).Inc()
	}
}
