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

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/client"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/database"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/server"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/worker"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
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
	dbConfig := cfg.Postgres
	cfgMutex.RUnlock()

	db, err := database.New(dbConfig)
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

	cfgMutex.RLock()
	serverConfig := cfg.Server
	cfgMutex.RUnlock()

	srv := server.New(serverConfig, log, metricsRepo)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	cfgMutex.RLock()
	githubConfig := cfg.Github
	weatherConfig := cfg.OpenWeather
	cfgMutex.RUnlock()

	githubClient := client.NewGithubClient(githubConfig.Token)
	weatherClient := client.NewOpenWeatherClient(weatherConfig.APIKey)

	wrk := worker.New(
		metricsRepo,
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
