package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	NotificationsConsumedTotal prometheus.Counter
	NotificationsSentTotal     *prometheus.CounterVec
	NotificationsErrorsTotal   *prometheus.CounterVec
	NotificationsSendDuration  *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)
	return &Metrics{
		NotificationsConsumedTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "notification_service_notifications_consumed_total",
			Help: "Total number of notifications consumed",
		}),
		NotificationsSentTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "notification_service_notifications_sent_total",
			Help: "Total number of notifications successfully sent",
		}, []string{"notifier_type"}),
		NotificationsErrorsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "notification_service_notification_errors_total",
			Help: "Total number of errors during notification processing or sending",
		}, []string{"stage", "notifier_type"}),
		NotificationsSendDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "notification_service_notification_send_duration_seconds",
			Help:    "Duration of sending notifications",
			Buckets: prometheus.DefBuckets,
		}, []string{"notifier_type"}),
	}
}
