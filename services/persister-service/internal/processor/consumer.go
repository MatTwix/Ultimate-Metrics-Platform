package processor

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/persister-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/persister-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/persister-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/persister-service/pkg/models"
)

type Consumer struct {
	messageConsumer consumer.MessageConsumer
	repo            repository.MetricRepository
	log             logger.Logger
}

func NewConsumer(messageConsumer consumer.MessageConsumer, repo repository.MetricRepository, log logger.Logger) *Consumer {
	return &Consumer{
		messageConsumer: messageConsumer,
		repo:            repo,
		log:             log,
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
				c.log.Error("failed to consumer metric", "error", err)
				time.Sleep(time.Second)
				continue
			}

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
