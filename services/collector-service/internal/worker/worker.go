package worker

import (
	"context"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/client"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/models"
)

type Worker struct {
	repo         repository.MetricRepository
	log          logger.Logger
	pollInterval time.Duration

	githubClient *client.GithubClient
	githubRepo   string

	weatherClient *client.OpenWeatherClient
	weatherCity   string
}

func New(
	repo repository.MetricRepository,
	log logger.Logger,
	pollInterval time.Duration,

	githubClient *client.GithubClient,
	githubRepo string,

	weatherClient *client.OpenWeatherClient,
	weatherCity string,
) *Worker {
	return &Worker{
		repo:         repo,
		log:          log,
		pollInterval: pollInterval,

		githubClient: githubClient,
		githubRepo:   githubRepo,

		weatherClient: weatherClient,
		weatherCity:   weatherCity,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.log.Info("starting worker", "poll_interval", w.pollInterval)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.collectAllMetrics(ctx)

	for {
		select {
		case <-ticker.C:
			w.collectAllMetrics(ctx)
		case <-ctx.Done():
			w.log.Info("worker stopped")
			return
		}
	}
}

func (w *Worker) collectAllMetrics(ctx context.Context) {
	w.collectUptimeMetrics(ctx)
	w.collectGithubMetrics(ctx)
	w.collectOpenWeatherMetrics(ctx)
}

func (w *Worker) collectUptimeMetrics(ctx context.Context) {
	w.log.Info("collecting uptime metrics...")

	//TODO: replace placeholder with real data
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

func (w *Worker) collectGithubMetrics(ctx context.Context) {
	w.log.Info("collecting github metrics...", "repo", w.githubRepo)

	info, err := w.githubClient.GetRepoInfo(ctx, w.githubRepo)
	if err != nil {
		w.log.Error("failed to get github repo info", "error", err)
		return
	}

	metric := models.Metric{
		Source:      "GitHub",
		Name:        "stargazers_count",
		Value:       float64(info.StargazersCount),
		Labels:      map[string]any{"repository": w.githubRepo},
		CollectedAt: time.Now(),
	}

	if err := w.repo.StoreBranch(ctx, []models.Metric{metric}); err != nil {
		w.log.Error("failed to store github metric", "error", err)
	} else {
		w.log.Info("successfully collected github metric", "stars", info.StargazersCount)
	}
}

func (w *Worker) collectOpenWeatherMetrics(ctx context.Context) {
	w.log.Info("collecting open weather metrics...", "city", w.weatherCity)

	data, err := w.weatherClient.GetCurrentTemperature(ctx, w.weatherCity)
	if err != nil {
		w.log.Error("failed to get weather data", "error", err)
		return
	}

	metric := models.Metric{
		Source:      "OpenWeatherMap",
		Name:        "temperature_celsius",
		Value:       data.Main.Temp,
		Labels:      map[string]any{"city": w.weatherCity},
		CollectedAt: time.Now(),
	}

	if err := w.repo.StoreBranch(ctx, []models.Metric{metric}); err != nil {
		w.log.Error("failed to store weather metric", "error", err)
	} else {
		w.log.Info("successfully collected weather metric", "temperature", data.Main.Temp)
	}
}
