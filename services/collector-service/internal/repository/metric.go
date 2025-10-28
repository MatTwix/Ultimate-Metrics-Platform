package repository

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/models"
)

type MetricRepository interface {
	StoreBranch(ctx context.Context, metrics []models.Metric) error
}
