package analitycs

import (
	"context"

	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/models"
)

type AnalyticsWriter interface {
	SaveAggregated(ctx context.Context, agg models.AggregatedMetric) error
}
