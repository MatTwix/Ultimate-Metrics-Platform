package kafka

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/models"
)

type KafkaBroker struct {
	producer *Producer
}

func NewKafkaBroker(producer *Producer) *KafkaBroker {
	return &KafkaBroker{
		producer: producer,
	}
}

func (k *KafkaBroker) SendMetrics(ctx context.Context, metrics []models.Metric) error {
	return k.producer.ProduceMetrics(ctx, metrics)
}

func (k *KafkaBroker) Close() error {
	return k.producer.Close()
}
