package server

import (
	"context"
	"net/http"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/config"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	httpServer *http.Server
	log        logger.Logger
}

func New(cfg config.ServerConfig, log logger.Logger, metricsRepo repository.MetricRepository) *Server {
	mux := http.NewServeMux()

	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)
	reg.MustRegister(collectors.NewGoCollector())

	apiRouter := http.NewServeMux()
	apiRouter.Handle("POST /v1/metrics", newMetricsHandler(metricsRepo, log, m))

	apiHandler := recoverMiddleware(apiRouter, log)
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	mux.HandleFunc("/healthz", healthCheckHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	finalHandler := prometheusMiddleware(mux, m)

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      finalHandler,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		log: log,
	}
}

func (s *Server) Start() error {
	s.log.Info("server is ready to handle requests", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
