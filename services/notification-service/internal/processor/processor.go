package processor

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/notifier"
)

type Processor struct {
	consumer consumer.MessageConsumer
	notifier notifier.Notifier
	stars    map[string]int
	log      logger.Logger
}

func New(consumer consumer.MessageConsumer, notifier notifier.Notifier) *Processor {
	return &Processor{
		consumer: consumer,
		notifier: notifier,
		stars:    make(map[string]int),
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
				p.log.Error("failed to consume metric", "error", err)
				continue
			}

			if metric.Source == "GitHub" && metric.Name == "stargazers_count" {
				repo := "golang/go"
				newStars := int(metric.Value)
				oldStars, exists := p.stars[repo]

				if exists && newStars > oldStars {
					p.notifier.NotifyStarInrcease(repo, oldStars, newStars)
				}

				p.stars[repo] = newStars
			}
		}
	}
}
