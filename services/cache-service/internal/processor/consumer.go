package processor

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/cache"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/logger"
)

type Consumer struct {
	messageConsumer consumer.MessageConsumer
	cache           cache.Cache
	log             logger.Logger
}

func NewConsumer(messageConsumer consumer.MessageConsumer, cache cache.Cache, log logger.Logger) *Consumer {
	return &Consumer{
		messageConsumer: messageConsumer,
		cache:           cache,
		log:             log,
	}
}

func getTTl(source string) time.Duration {
	switch source {
	case "GitHub":
		return 10 * time.Minute
	case "OpenWeatherMap":
		return 5 * time.Minute
	case "UptimeChecker":
		return 1 * time.Minute
	default:
		return 5 * time.Minute
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.log.Info("starting cache consumer")

	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer stopped")
			return
		default:
			metric, err := c.messageConsumer.ConsumeMetric(ctx)
			if err != nil {
				c.log.Error("failed to consume metric")
				time.Sleep(time.Second)
				continue
			}

			if cached, ok := metric.Labels["cached"].(string); ok && cached == "true" {
				c.log.Debug("skipping cached metric", "source", metric.Source)
				continue
			}

			ttl := getTTl(metric.Source)

			if err := c.cache.SetMetric(ctx, metric, ttl); err != nil {
				c.log.Error("failed to cache metric", "error", err)
				continue
			}

			c.log.Info("successfully cached metric", "source", metric.Source, "name", metric.Name)
		}
	}
}
