package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/cache-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/cache-service/pkg/models"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic, groupID string) consumer.MessageConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
	})
	return &KafkaConsumer{reader: reader}
}

func (k *KafkaConsumer) ConsumeMetric(ctx context.Context) (models.Metric, error) {
	msg, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return models.Metric{}, err
	}

	var metric models.Metric
	err = json.Unmarshal(msg.Value, &metric)

	return metric, err
}

func (k *KafkaConsumer) Close() error {
	return k.reader.Close()
}
