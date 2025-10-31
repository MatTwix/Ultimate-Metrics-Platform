package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/internal/aggregator"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/logger"
)

type Processor struct {
	aggregator *aggregator.Aggregator
	log        logger.Logger
	interval   time.Duration
}

type aggregationResult struct {
	source string
	name   string
	err    error
}

func NewProcessor(aggregator *aggregator.Aggregator, log logger.Logger, interval time.Duration) *Processor {
	return &Processor{
		aggregator: aggregator,
		log:        log,
		interval:   interval,
	}
}

func (p *Processor) Start(ctx context.Context, metrics []config.MetricInfo) {
	p.log.Info("starting processor", "interval", p.interval)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.aggregateAll(ctx, metrics)
		case <-ctx.Done():
			p.log.Info("processor stopped")
			return
		}
	}
}

func (p *Processor) aggregateAll(ctx context.Context, metrics []config.MetricInfo) {
	resultCh := make(chan aggregationResult, len(metrics)*10)

	var wg sync.WaitGroup
	for _, m := range metrics {
		for _, name := range m.Names {
			wg.Add(1)
			go func(source, name string) {
				defer wg.Done()
				err := p.aggregator.AggregateHourly(ctx, source, name)
				resultCh <- aggregationResult{source: source, name: name, err: err}
			}(m.Source, name)
		}
	}

	wg.Wait()
	close(resultCh)

	var errors []string
	var successfullAggregations []string

	for res := range resultCh {
		if res.err != nil {
			errors = append(errors, fmt.Sprintf("%s/%s: %v", res.source, res.name, res.err))
		} else {
			successfullAggregations = append(successfullAggregations, fmt.Sprintf("%s/%s", res.source, res.name))
		}
	}

	if len(errors) > 0 {
		p.log.Warn("aggregations completed with some errors", "errors", errors, "count", len(errors))
	}

	if len(successfullAggregations) > 0 {
		p.log.Info("aggregated metrics", "metrics", successfullAggregations)
	}
}
