package consumer

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/models"
)

type MessageConsumer interface {
	ConsumeMetric(ctx context.Context) (models.Metric, error)
	Close() error
}
