package worker

import (
	"context"
	"net/http"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/internal/client"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/broker"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/grpc"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/models"
)

type Worker struct {
	broker       broker.MessageBroker
	cacheClient  *grpc.CacheClient
	log          logger.Logger
	pollInterval time.Duration

	githubClient *client.GithubClient
	githubRepo   string

	weatherClient *client.OpenWeatherClient
	weatherCity   string
}

func New(
	broker broker.MessageBroker,
	cacheClient *grpc.CacheClient,
	log logger.Logger,
	pollInterval time.Duration,

	githubClient *client.GithubClient,
	githubRepo string,

	weatherClient *client.OpenWeatherClient,
	weatherCity string,
) *Worker {
	return &Worker{
		broker:       broker,
		cacheClient:  cacheClient,
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

func checkUptime(url string, timeout time.Duration) (float64, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return 100, nil
	}

	return 0, nil
}

func (w *Worker) collectUptimeMetrics(ctx context.Context) {
	w.log.Info("collecting uptime metrics...")

	cachedMetric, err := w.cacheClient.GetCachedMetric(ctx, "UptimeChecker", "availability_percent")
	if err != nil {
		w.log.Warn("failed to get from cache, falling back to API", "error", err)
	} else if cachedMetric != nil {
		w.log.Info("using cached uptimechecher metric", "percent", cachedMetric.Value)
		metric := *cachedMetric

		metric.CollectedAt = time.Now()
		if metric.Labels == nil {
			metric.Labels = map[string]any{}
		}
		metric.Labels["cached"] = "true"

		if err := w.broker.SendMetrics(ctx, []models.Metric{metric}); err != nil {
			w.log.Error("failed to send cached metric")
		}
		return
	}

	uptimePercent, err := checkUptime("https://www.google.com", 5*time.Second)
	if err != nil {
		w.log.Error("failed to check uptime", "error", err)
		uptimePercent = 0
	}

	metric := models.Metric{
		Source:      "UptimeChecker",
		Name:        "availability_percent",
		Value:       uptimePercent,
		Labels:      map[string]any{"site": "google.com"},
		CollectedAt: time.Now(),
	}

	if err := w.broker.SendMetrics(ctx, []models.Metric{metric}); err != nil {
		w.log.Error("failed to send uptime metric", "error", err)
		return
	}

	w.log.Info("successfully collected and sended uptime metric")
}

func (w *Worker) collectGithubMetrics(ctx context.Context) {
	w.log.Info("collecting github metrics...", "repo", w.githubRepo)

	cachedMetric, err := w.cacheClient.GetCachedMetric(ctx, "GitHub", "stargazers_count")
	if err != nil {
		w.log.Warn("failed to get from cache, falling back to API", "error", err)
	} else if cachedMetric != nil {
		w.log.Info("using cached github metric", "stars", cachedMetric.Value)
		metric := *cachedMetric

		metric.CollectedAt = time.Now()
		if metric.Labels == nil {
			metric.Labels = map[string]any{}
		}
		metric.Labels["cached"] = "true"

		if err := w.broker.SendMetrics(ctx, []models.Metric{metric}); err != nil {
			w.log.Error("failed to send cached metric")
		}
		return
	}

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

	if err := w.broker.SendMetrics(ctx, []models.Metric{metric}); err != nil {
		w.log.Error("failed to send github metric", "error", err)
	} else {
		w.log.Info("successfully collected github metric", "stars", info.StargazersCount)
	}
}

func (w *Worker) collectOpenWeatherMetrics(ctx context.Context) {
	w.log.Info("collecting open weather metrics...", "city", w.weatherCity)

	cachedMetric, err := w.cacheClient.GetCachedMetric(ctx, "OpenWeatherMap", "temperature_celsius")
	if err != nil {
		w.log.Warn("failed to get from cache, falling back to API", "error", err)
	} else if cachedMetric != nil {
		w.log.Info("using cached openweather metric", "temperature", cachedMetric.Value)
		metric := *cachedMetric

		metric.CollectedAt = time.Now()
		if metric.Labels == nil {
			metric.Labels = map[string]any{}
		}
		metric.Labels["cached"] = "true"

		if err := w.broker.SendMetrics(ctx, []models.Metric{metric}); err != nil {
			w.log.Error("failed to send cached metric")
		}
		return
	}

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

	if err := w.broker.SendMetrics(ctx, []models.Metric{metric}); err != nil {
		w.log.Error("failed to send weather metric", "error", err)
	} else {
		w.log.Info("successfully collected weather metric", "temperature", data.Main.Temp)
	}
}
