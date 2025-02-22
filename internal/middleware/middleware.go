package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

func LoggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Info("Начало обработки запроса", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Info("Запрос обработан", "method", r.Method, "path", r.URL.Path, "duration", duration)
	})
}

func RecoveryMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("Паника в обработчике запроса", "error", rec, "stack", string(debug.Stack()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
