package server

import (
	"encoding/json"
	"net/http"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/repository"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/models"
)

type metricsHandler struct {
	repo repository.MetricRepository
	log  logger.Logger
}

func newMetricsHandler(repo repository.MetricRepository, log logger.Logger) *metricsHandler {
	return &metricsHandler{
		repo: repo,
		log:  log,
	}
}

func (h metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metric
	if err := json.NewDecoder(r.Body); err != nil {
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}

	if err := h.repo.StoreBranch(r.Context(), metrics); err != nil {
		if batchErr, ok := err.(*repository.BatchInsertError); ok {
			h.log.Warn("some metrics failed to be stored", "error", batchErr)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMultiStatus)
			json.NewEncoder(w).Encode(map[string]any{
				"message":           "Request processed with some failures",
				"successfull_count": batchErr.SuccessfullCount,
				"failure_count":     batchErr.FailedCount,
			})

			return
		}

		h.log.Error("failed to store metrics batch", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Metrics accepted"))
}
