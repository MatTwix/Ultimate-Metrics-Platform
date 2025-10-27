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

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/configs"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

func main() {
	var cfgMutex sync.RWMutex

	cfg, err := configs.LoadConfig("./config.yaml")
	if err != nil {
		slog.Error("failed to load initial configuration, shutting down", "error", err)
		os.Exit(1)
	}

	var log logger.Logger = logger.New(cfg.Env)

	log.Info("starting collector-service", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	go configs.WatchConfig(func(e fsnotify.Event) {
		log.Info("config gile changed, reloading...", "file", e.Name)

		cfgMutex.Lock()
		reloadedCfg, err := configs.LoadConfig("./config.yaml")
		if err != nil {
			log.Error("error updating config", "error", err)
			return
		}

		cfg = reloadedCfg
		cfgMutex.Unlock()

		log = logger.New(reloadedCfg.Env)
		log.Info("config reloaded successfully")
	})

	mux := http.NewServeMux()

	cfgMutex.RLock()
	serverConfig := cfg.Server
	cfgMutex.RUnlock()

	server := &http.Server{
		Addr:         ":" + serverConfig.Port,
		Handler:      mux,
		ReadTimeout:  serverConfig.Timeout,
		WriteTimeout: serverConfig.Timeout,
		IdleTimeout:  serverConfig.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "addr", server.Addr, "error", err)
			os.Exit(1)
		}
		log.Info("server is ready to handle requests", "addr", server.Addr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", "error", err)
	}

	log.Info("server gracefully stopped")
}
