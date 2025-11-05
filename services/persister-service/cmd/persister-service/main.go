package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/database"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/processor"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/kafka"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/persister-service/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	dbConfig := cfg.Postgres
	brokerConfig := cfg.Broker
	serverConfig := cfg.Server
	cfgMutex.RUnlock()

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	db, err := database.New(dbConfig, m)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connected successfully")

	if err := db.RunMigrations(); err != nil {
		log.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations applied successfully")
	var metricsRepo repository.MetricRepository = db

	var msgCons consumer.MessageConsumer

	switch brokerConfig.Type {
	case "kafka":
		msgCons = kafka.NewKafkaConsumer(brokerConfig.Kafka.Brokers, brokerConfig.Kafka.Topic, brokerConfig.Kafka.GroupID)
	default:
		log.Error("unsupported consumer type", "type", brokerConfig.Type)
		os.Exit(1)
	}
	defer msgCons.Close()

	proc := processor.NewConsumer(msgCons, metricsRepo, log, m)

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		log.Info("metrics server listening", "port", serverConfig.Port)
		if err := http.ListenAndServe(":"+serverConfig.Port, nil); err != nil {
			log.Error("failed to start metrics server", "error", err)
			os.Exit(1)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("starting persister-service")
	proc.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()

	log.Info("service gracefully stopped")
}
