package event

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventPushCounter      prometheus.Counter
	eventPushErrorCounter prometheus.Counter
	eventPushTimeHistMs   prometheus.Histogram
)

func init() {
	eventPushCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "events_push_total",
		},
	)
	eventPushErrorCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "events_push_error_total",
		},
	)
	eventPushTimeHistMs = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "event_push_time_ms",
			Buckets: []float64{2, 3, 5, 8, 13, 21, 44, 65, 109, 174, 283, 457, 740, 1197, 1937, 2984, 4581, 7165},
		},
	)
}
