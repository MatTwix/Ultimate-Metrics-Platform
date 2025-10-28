package server

import (
	"net/http"
	"runtime/debug"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/pkg/logger"
)

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
