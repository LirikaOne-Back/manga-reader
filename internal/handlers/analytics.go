package handlers

import (
	"log/slog"
	"manga-reader/internal/analytics"
	"manga-reader/internal/apperror"
	"manga-reader/internal/db"
	"manga-reader/internal/response"
	"net/http"
	"strconv"
)

type AnalyticsHandler struct {
	MangaRepo db.MangaRepository
	Analytics *analytics.AnalyticsService
	Logger    *slog.Logger
}

func (h *AnalyticsHandler) GetPopularManga(w http.ResponseWriter, r *http.Request) error {
	period := r.URL.Query().Get("period")
	limitStr := r.URL.Query().Get("limit")

	if period == "" {
		period = "all"
	}
	limit := int64(10)
	if limitStr == "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	topEntries, err := h.Analytics.GetTopManga(r.Context(), period, limit)
	if err != nil {
		return apperror.NewInternalServerError("Ошибка получения рейтинга манги", err)
	}

	var result []analytics.MangaWithViews
	for _, entry := range topEntries {
		manga, err := h.MangaRepo.GetByID(entry.MangaID)
		if err != nil {
			h.Logger.Error("Ошибка получения информации о манге", "manga_id", entry.MangaID, "err", err)
			continue
		}
		result = append(result, analytics.MangaWithViews{
			ID:          manga.ID,
			Title:       manga.Title,
			Description: manga.Description,
			Views:       entry.Views,
		})
	}

	response.Success(w, http.StatusOK, result)
	return nil
}

func (h *AnalyticsHandler) ResetDailyStats(w http.ResponseWriter, r *http.Request) error {
	if err := h.Analytics.InitializeDailyStats(r.Context()); err != nil {
		return apperror.NewInternalServerError("Ошибка сброса дневной статистики", err)
	}

	response.Success(w, http.StatusOK, map[string]string{"status": "success", "message": "Дневная статистика сброшена"})
	return nil
}

func (h *AnalyticsHandler) ResetWeeklyStats(w http.ResponseWriter, r *http.Request) error {
	if err := h.Analytics.InitializeWeeklyStats(r.Context()); err != nil {
		return apperror.NewInternalServerError("Ошибка сброса недельной статистики", err)
	}

	response.Success(w, http.StatusOK, map[string]string{"status": "success", "message": "Недельная статистика сброшена"})
	return nil
}

func (h *AnalyticsHandler) ResetMonthlyStats(w http.ResponseWriter, r *http.Request) error {
	if err := h.Analytics.InitializeMonthlyStats(r.Context()); err != nil {
		return apperror.NewInternalServerError("Ошибка сброса месячной статистики", err)
	}

	response.Success(w, http.StatusOK, map[string]string{"status": "success", "message": "Месячная статистика сброшена"})
	return nil
}
