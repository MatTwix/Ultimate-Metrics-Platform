package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/broker"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/models"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	return &Producer{writer: writer}
}

func (p *Producer) ProduceMetrics(ctx context.Context, metrics []models.Metric) error {
	kafkaMessages := make([]kafka.Message, 0, len(metrics))
	var serErr broker.SerialisationError

	for i, metric := range metrics {
		msgBytes, err := json.Marshal(metric)
		if err != nil {
			serErr.FailedCount++
			serErr.Errors = append(serErr.Errors, fmt.Errorf("metric %d %s/%s: failed to marshal metric: %w",
				i,
				metric.Source,
				metric.Name,
				err,
			))
			continue
		}

		kafkaMessages = append(kafkaMessages, kafka.Message{
			Value: msgBytes,
		})
		serErr.SuccessfullCount++
	}

	if len(kafkaMessages) > 0 {
		if err := p.writer.WriteMessages(ctx, kafkaMessages...); err != nil {
			return fmt.Errorf("failed to sent %d messages to Kafka: %w", len(kafkaMessages), err)
		}
	}

	if serErr.FailedCount > 0 {
		return &serErr
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
