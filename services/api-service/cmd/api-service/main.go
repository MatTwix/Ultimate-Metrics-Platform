package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/internal/database"
	grpcInternal "github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/internal/grpc"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/proto"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
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
	grpcConfig := cfg.GRPC
	serverConfig := cfg.Server
	cfgMutex.RUnlock()

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	db, err := database.New(dbConfig)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connected successfully")

	reader := database.NewPostgresMetricsReeader(db)

	grpcServer := grpc.NewServer()
	apiServer := grpcInternal.NewServer(reader, m)
	proto.RegisterMetricsServiceServer(grpcServer, apiServer)

	lis, err := net.Listen("tcp", ":"+grpcConfig.Port)
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		log.Info("gRPC server listening", "port", grpcConfig.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		log.Info("metrics server listening", "port", serverConfig.Port)
		if err := http.ListenAndServe(":"+serverConfig.Port, nil); err != nil {
			log.Error("failed to start metrics server", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("starting persister-service")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("service gracefully stopped")
}
