package handlers

import (
	"manga-reader/internal/apperror"
	"manga-reader/internal/auth"
	"manga-reader/internal/middleware"
	"net/http"
)

func RegisterAnalyticsRoutes(mux *http.ServeMux, ah *AnalyticsHandler) {
	mux.HandleFunc("/analytics/popular", middleware.ErrorHandler(ah.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodGet {
			return ah.GetPopularManga(w, r)
		}
		return apperror.NewBadRequestError("Метод не поддерживается", nil)
	}))

	mux.Handle("/analytics/reset/daily", auth.AuthMiddleware(middleware.ErrorHandler(ah.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return ah.ResetDailyStats(w, r)
		}
		return apperror.NewBadRequestError("Метод не поддерживается", nil)
	})))

	mux.Handle("/analytics/reset/weekly", auth.AuthMiddleware(middleware.ErrorHandler(ah.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return ah.ResetWeeklyStats(w, r)
		}
		return apperror.NewBadRequestError("Метод не поддерживается", nil)
	})))

	mux.Handle("/analytics/reset/monthly", auth.AuthMiddleware(middleware.ErrorHandler(ah.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return ah.ResetMonthlyStats(w, r)
		}
		return apperror.NewBadRequestError("Метод не поддерживается", nil)
	})))
}
