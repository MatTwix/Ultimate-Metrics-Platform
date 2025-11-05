package processor

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/models"
)

type Consumer struct {
	messageConsumer consumer.MessageConsumer
	repo            repository.MetricRepository
	log             logger.Logger
	metrics         *metrics.Metrics
}

func NewConsumer(messageConsumer consumer.MessageConsumer, repo repository.MetricRepository, log logger.Logger, metrics *metrics.Metrics) *Consumer {
	return &Consumer{
		messageConsumer: messageConsumer,
		repo:            repo,
		log:             log,
		metrics:         metrics,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.log.Info("starting consumer")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer stopped")
			return
		default:
			metric, err := c.messageConsumer.ConsumeMetric(ctx)
			if err != nil {
				c.log.Error("failed to consume metric", "error", err)
				time.Sleep(time.Second)
				continue
			}

			c.metrics.MetricsConsumedTotal.Inc()

			if err := c.repo.StoreBranch(ctx, []models.Metric{metric}); err != nil {
				if batchErr, ok := err.(*repository.BatchInsertError); ok {
					c.log.Warn("some metrics failed to be stored", "error", batchErr)
					continue
				}

				c.log.Error("failed to store metrics", "error", err)
				continue
			}

			c.log.Info("successfully stored metric", "source", metric.Source, "name", metric.Name, "value", metric.Value)
		}
	}
}
