package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	TotalRequests  prometheus.Counter
	ResponseStatus prometheus.CounterVec
	HttpDuration   prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	return &Metrics{
		TotalRequests: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of get requests",
		}),
		ResponseStatus: *promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "http_responce_status",
			Help: "Status of HTTP response",
		}, []string{"status"}),
		HttpDuration: *promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_responce_seconds",
			Help: "Duration of HTTP requests",
		}, []string{"path"}),
	}
}
