package cache

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/models"
)

type Cache interface {
	SetMetric(ctx context.Context, metric models.Metric) error
	GetMetric(ctx context.Context, source, name string) (*models.Metric, error)
}
