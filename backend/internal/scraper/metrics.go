package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	ruArtistsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scraper_ru_artists_total",
		Help: "Current total number of Russian artists in the database.",
	})
)

func SetRuArtistsTotal(n int) {
	ruArtistsTotal.Set(float64(n))
}

func PushMetrics(url string) error {
	return push.New(url, "scraper").
		Gatherer(prometheus.DefaultGatherer).
		Push()
}
