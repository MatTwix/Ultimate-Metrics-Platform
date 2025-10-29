package broker

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/models"
)

type MessageBroker interface {
	SendMetrics(ctx context.Context, metrics []models.Metric) error
	Close() error
}
