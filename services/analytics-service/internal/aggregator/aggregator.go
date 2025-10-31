package aggregator

import (
	"context"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/analitycs"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/grpc"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/models"
)

type Aggregator struct {
	apiClient *grpc.MetricsClient
	writer    analitycs.AnalyticsWriter
}

func NewAggregator(apiClient *grpc.MetricsClient, writer analitycs.AnalyticsWriter) *Aggregator {
	return &Aggregator{apiClient: apiClient, writer: writer}
}

func (a *Aggregator) AggregateHourly(ctx context.Context, source, name string) error {
	metrics, err := a.apiClient.GetMetrics(ctx, source, name, 60)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if len(metrics) == 0 {
		return nil
	}

	sum := 0.0
	mn := metrics[0].Value
	mx := metrics[0].Value
	for _, m := range metrics {
		sum += m.Value
		mn = min(mn, m.Value)
		mx = max(mx, m.Value)
	}

	avg := sum / float64(len(metrics))

	agg := models.AggregatedMetric{
		Source:    source,
		Name:      name,
		AvgValue:  avg,
		MinValue:  mn,
		MaxValue:  mx,
		Count:     len(metrics),
		TimeRange: "hour",
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
	}

	if err := a.writer.SaveAggregated(ctx, agg); err != nil {
		return fmt.Errorf("failed to save aggregated metric: %w", err)
	}

	return nil
}
