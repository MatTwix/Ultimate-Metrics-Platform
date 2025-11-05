package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RequestsTotal       *prometheus.CounterVec
	RequestsFailedTotal *prometheus.CounterVec
	RequestDuration     *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)
	return &Metrics{
		RequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "api_service_requests_total",
			Help: "Total number of gRPC requests",
		}, []string{"method"}),
		RequestsFailedTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "api_service_requests_failed_total",
			Help: "Total number of failed gRPC requests",
		}, []string{"method"}),
		RequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "api_service_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
	}
}
