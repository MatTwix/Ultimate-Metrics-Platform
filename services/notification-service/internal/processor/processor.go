package processor

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/notifier"
)

type Processor struct {
	consumer consumer.MessageConsumer
	notifier notifier.Notifier
	stars    map[string]int
	log      logger.Logger
	metrics  *metrics.Metrics
}

func New(consumer consumer.MessageConsumer, notifier notifier.Notifier, log logger.Logger, metrics *metrics.Metrics) *Processor {
	return &Processor{
		consumer: consumer,
		notifier: notifier,
		log:      log,
		stars:    make(map[string]int),
		metrics:  metrics,
	}
}

func (p *Processor) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			metric, err := p.consumer.ConsumeMetric(ctx)
			if err != nil {
				p.metrics.NotificationsErrorsTotal.WithLabelValues("consume", "none").Inc()
				p.log.Error("failed to consume metric", "error", err)
				time.Sleep(time.Second)
				continue
			}

			p.metrics.NotificationsConsumedTotal.Inc()

			if metric.Source == "GitHub" && metric.Name == "stargazers_count" {
				repo := "golang/go"
				newStars := int(metric.Value)
				oldStars, exists := p.stars[repo]

				if exists && newStars > oldStars {
					err := p.notifier.NotifyStarInrcease(repo, oldStars, newStars)
					if err != nil {
						p.log.Error("failed to notify", "error", err)
						time.Sleep(time.Second)
					}
					p.log.Info("notification succeccfully sended", "repo", repo, "stars_incrementations", newStars-oldStars)
				}

				p.stars[repo] = newStars
			}
		}
	}
}
