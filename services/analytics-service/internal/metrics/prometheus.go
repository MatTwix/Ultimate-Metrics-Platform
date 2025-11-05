package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	BatchProcessingDuration prometheus.Histogram
	MetricsSaved            prometheus.Counter
	DatabaseErrors          prometheus.Counter
	BatchSize               prometheus.Histogram
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)
	return &Metrics{
		BatchProcessingDuration: factory.NewHistogram(prometheus.HistogramOpts{
			Name:    "analytics_service_processing_duration_seconds",
			Help:    "Duration of processing a batch of metrics from aggregation",
			Buckets: prometheus.DefBuckets,
		}),
		MetricsSaved: factory.NewCounter(prometheus.CounterOpts{
			Name: "analytics_service_metrics_saved_total",
			Help: "The total number of aggregated metrics successfully saved to the database",
		}),
		DatabaseErrors: factory.NewCounter(prometheus.CounterOpts{
			Name: "analytics_service_database_errors_total",
			Help: "The total number of database write errors",
		}),
		BatchSize: factory.NewHistogram(prometheus.HistogramOpts{
			Name:    "analytics_service_batch_size",
			Help:    "Size of metric batches received for processing",
			Buckets: prometheus.LinearBuckets(10, 10, 10),
		}),
	}
}
