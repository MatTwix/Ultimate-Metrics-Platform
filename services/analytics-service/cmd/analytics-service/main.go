package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/internal/aggregator"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/internal/database"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/internal/processor"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/grpc"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	mongoConfig := cfg.Mongo
	metrics := cfg.Metrics
	urls := cfg.Urls
	cfgMutex.RUnlock()

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoConfig.URI()))
	if err != nil {
		log.Error("failed to connect to MongoDB", "uri", mongoConfig.URI(), "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		log.Error("failed to ping MpngoDB", "error", err)
		os.Exit(1)
	}
	log.Info("MongoDB connected successfully")

	writer := database.NewMongoAnalyticsWriter(mongoClient, mongoConfig.DBName, mongoConfig.Collection)

	apiClient, err := grpc.NewMetricsClient(urls.ApiService)
	if err != nil {
		log.Error("failed to connect to API service", "error", err)
		os.Exit(1)
	}
	defer apiClient.Close()

	aggregator := aggregator.NewAggregator(apiClient, writer)

	processor := processor.NewProcessor(aggregator, log, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("starting analytics-service")
	processor.Start(ctx, metrics)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()

	log.Info("service gracefully stopped")
}
