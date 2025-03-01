package middleware

import (
	"log/slog"
	"manga-reader/internal/apperror"
	"manga-reader/internal/response"
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
				stackTrace := string(debug.Stack())
				log.Error("Паника в обработчике запроса",
					"error", rec,
					"stack", stackTrace,
					"method", r.Method,
					"path", r.URL.Path)
				appErr := apperror.NewInternalServerError("Внутренняя ошибка сервера", nil)

				response.Error(w, log, appErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

func ErrorHandler(log *slog.Logger, f ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			response.Error(w, log, err)
		}
	}
}
