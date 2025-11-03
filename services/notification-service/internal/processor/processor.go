package processor

import (
	"context"
	"time"

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

func New(consumer consumer.MessageConsumer, notifier notifier.Notifier, log logger.Logger) *Processor {
	return &Processor{
		consumer: consumer,
		notifier: notifier,
		log:      log,
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
				time.Sleep(time.Second)
				continue
			}

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
