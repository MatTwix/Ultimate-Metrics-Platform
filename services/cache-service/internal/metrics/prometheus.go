package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	CacheRequests          prometheus.CounterVec
	CacheOperationDuration prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	return &Metrics{
		CacheRequests: *promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "cache_requests_total",
			Help: "Total number of cache requests",
		}, []string{"operation", "status"}),
		CacheOperationDuration: *promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Duration of cache operations",
			Buckets: prometheus.DefBuckets,
		}, []string{"operation"}),
	}
}
