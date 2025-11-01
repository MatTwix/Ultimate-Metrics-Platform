package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/internal/processor"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/consumer"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/kafka"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/pkg/notifier"
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
	emailConfig := cfg.Email
	brokerConfig := cfg.Broker
	cfgMutex.RUnlock()

	var msgCons consumer.MessageConsumer

	switch brokerConfig.Type {
	case "kafka":
		msgCons = kafka.NewKafkaConsumer(brokerConfig.Kafka.Brokers, brokerConfig.Kafka.Topic, brokerConfig.Kafka.GroupID)
	default:
		log.Error("unsupported consumer type", "type", brokerConfig.Type)
		os.Exit(1)
	}
	defer msgCons.Close()

	emailNotifier := notifier.NewEmailNotifier(
		emailConfig.SMTPHost,
		emailConfig.SMTPPort,
		emailConfig.From,
		emailConfig.Username,
		emailConfig.Password,
		emailConfig.To,
	)

	proc := processor.New(msgCons, emailNotifier)

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
