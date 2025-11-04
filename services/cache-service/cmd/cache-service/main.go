package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/cache"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/config"
	grpcInternal "github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/grpc"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/processor"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/kafka"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/proto"
	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
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
	brokerConfig := cfg.Broker
	redisConfig := cfg.Redis
	grpcConfig := cfg.GRPC
	serverConfig := cfg.Server
	cfgMutex.RUnlock()

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	defer redisClient.Close()

	cacheImpl := cache.NewredisCache(redisClient)

	var msgCons consumer.MessageConsumer

	switch brokerConfig.Type {
	case "kafka":
		msgCons = kafka.NewKafkaConsumer(brokerConfig.Kafka.Brokers, brokerConfig.Kafka.Topic, brokerConfig.Kafka.GroupID)
	default:
		log.Error("unsupported consumer type", "type", brokerConfig.Type)
		os.Exit(1)
	}
	defer msgCons.Close()

	proc := processor.NewConsumer(msgCons, cacheImpl, log, m)

	grpcServer := grpc.NewServer()
	cacheServer := grpcInternal.NewServer(cacheImpl, m)
	proto.RegisterCacheServiceServer(grpcServer, cacheServer)

	lis, err := net.Listen("tcp", ":"+grpcConfig.Port)
	if err != nil {
		log.Error("failed to listen gRPC server", "error", err)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("starting cache-service")
	proc.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()

	log.Info("service gracefully stopped")
}
