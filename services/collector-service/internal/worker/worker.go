package worker

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/models"
)

type Worker struct {
	repo         repository.MetricRepository
	log          logger.Logger
	pollInterval time.Duration
}

func New(repo repository.MetricRepository, log logger.Logger, pollInterval time.Duration) *Worker {
	return &Worker{
		repo:         repo,
		log:          log,
		pollInterval: pollInterval,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.log.Info("starting worker", "poll_interval", w.pollInterval)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.collectUptimeMetrics(ctx)
		case <-ctx.Done():
			w.log.Info("worker stopped")
			return
		}
	}
}

func (w *Worker) collectUptimeMetrics(ctx context.Context) {
	w.log.Info("collecting uptime metrics...")

	// placeholder
	mockMetric := models.Metric{
		Source:      "Source",
		Name:        "availability_percent",
		Value:       99.9,
		Labels:      map[string]any{"site": "google.com"},
		CollectedAt: time.Now(),
	}

	if err := w.repo.StoreBranch(ctx, []models.Metric{mockMetric}); err != nil {
		w.log.Error("failed to store uptime metric", "error", err)
		return
	}

	w.log.Info("successfully collected and stored uptime metric")
}
