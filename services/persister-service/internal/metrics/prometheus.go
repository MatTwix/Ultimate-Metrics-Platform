package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	MetricsConsumedTotal prometheus.Counter
	MetricsSavedTotal    prometheus.Counter
	DatabaseErrorsTotal  prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)
	return &Metrics{
		MetricsConsumedTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "persister_service_metrics_consumed_total",
			Help: "Total number of consumed metrics",
		}),
		MetricsSavedTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "persister_service_metrics_saved_total",
			Help: "Total number of metrics successfully saved to the database",
		}),
		DatabaseErrorsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "persister_service_database_errors_total",
			Help: "Total number of database write errors",
		}),
	}
}
