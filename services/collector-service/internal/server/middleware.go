package server

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/metrics"
	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func recoverMiddleware(next http.Handler, log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered", "error", err, "stack", debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func prometheusMiddleware(next http.Handler, m *metrics.Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriterInterceptor{
			ResponseWriter: w,
			statusCode:     http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(ww.statusCode)

		m.HttpDuration.WithLabelValues(r.URL.Path).Observe(duration)
		m.TotalRequests.Inc()
		m.ResponseStatus.WithLabelValues(statusCode).Inc()
	})
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
