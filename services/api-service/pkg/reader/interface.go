package reader

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/api-service/pkg/models"
)

type MetricsReader interface {
	GetMetrics(ctx context.Context, source, name string, limit int) ([]models.Metric, error)
	GetMetric(ctx context.Context, source, name string) (*models.Metric, error)
}
