package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/internal/client"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/internal/server"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/internal/worker"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/broker"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/grpc"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/kafka"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

func main() {
	var cfgMutex sync.RWMutex

	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		slog.Error("failed to load initial configuration, shutting down", "error", err)
		os.Exit(1)
	}

	var log logger.Logger = logger.New(cfg.Env)

	log.Info("starting collector-service", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	go config.WatchConfig(func(e fsnotify.Event) {
		log.Info("config gile changed, reloading...", "file", e.Name)

		cfgMutex.Lock()
		reloadedCfg, err := config.LoadConfig("./config.yaml")
		if err != nil {
			log.Error("error updating config", "error", err)
			return
		}

		cfg = reloadedCfg
		cfgMutex.Unlock()

		log = logger.New(reloadedCfg.Env)
		log.Info("config reloaded successfully")
	})

	cfgMutex.RLock()
	brokerConfig := cfg.Broker
	serverConfig := cfg.Server
	githubConfig := cfg.Github
	weatherConfig := cfg.OpenWeather
	urlsConfig := cfg.Urls
	cfgMutex.RUnlock()

	var msgBroker broker.MessageBroker
	switch brokerConfig.Type {
	case "kafka":
		producer := kafka.NewProducer(cfg.Broker.Kafka.Brokers, cfg.Broker.Kafka.Topic)
		msgBroker = kafka.NewKafkaBroker(producer)
		log.Info("kafka broker created", "brokers", cfg.Broker.Kafka.Brokers, "topic", cfg.Broker.Kafka.Topic)
	default:
		log.Error("unsupported broker type", "type", cfg.Broker.Type)
		os.Exit(1)
	}

	defer func() {
		if err := msgBroker.Close(); err != nil {
			log.Error("failed to close broker", "error", err)
		}
	}()

	srv := server.New(serverConfig, log, msgBroker)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	cacheClient, err := grpc.NewCacheClient(urlsConfig.CacheService)
	if err != nil {
		log.Warn("cache service unavailable, proceeding without cache", "error", err)
		cacheClient = nil
	}

	githubClient := client.NewGithubClient(githubConfig.Token)
	weatherClient := client.NewOpenWeatherClient(weatherConfig.APIKey)

	wrk := worker.New(
		msgBroker,
		cacheClient,
		log,
		cfg.Worker.PollInterval,

		githubClient,
		githubConfig.Repository,

		weatherClient,
		weatherConfig.City,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wrk.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()

	log.Info("server is shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", "error", err)
	}

	log.Info("server gracefully stopped")
}
