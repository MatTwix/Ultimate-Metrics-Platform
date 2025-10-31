package metrics

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/models"
)

type MetricsReader interface {
	GetMetricsInRange(ctx context.Context, source, name string, from, to time.Time) ([]models.Metric, error)
}
