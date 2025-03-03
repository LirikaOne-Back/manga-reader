package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"manga-reader/internal/analytics"
	"manga-reader/internal/apperror"
	"manga-reader/internal/cache"
	"manga-reader/internal/db"
	"manga-reader/internal/response"
	"manga-reader/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MangaHandler struct {
	Repo      db.MangaRepository
	Logger    *slog.Logger
	Cache     cache.Cache
	Analytics *analytics.AnalyticsService
}

func (h *MangaHandler) List(w http.ResponseWriter, r *http.Request) error {
	cacheKey := "manga:list"
	cachedDate, err := h.Cache.Get(r.Context(), cacheKey)
	if err == nil && cachedDate != "" {
		h.Logger.Info("Cache hit", "key", cacheKey)

		var mangas []*models.Manga
		if err := json.Unmarshal([]byte(cachedDate), &mangas); err != nil {
			h.Logger.Error("Ошибка десериализации из кеша", "err", err)
		} else {
			response.Success(w, http.StatusOK, mangas)
			return nil
		}
	}
	h.Logger.Info("Cache miss", "key", cacheKey)
	mangas, err := h.Repo.List()
	if err != nil {
		return apperror.NewDatabaseError("Ошибка получения списка манги", err)
	}

	jsonDate, err := json.Marshal(mangas)
	if err != nil {
		if err = h.Cache.Set(r.Context(), cacheKey, string(jsonDate), 5*time.Minute); err != nil {
			h.Logger.Error("Ошибка записи в кеш", "err", err)
		}
	}

	response.Success(w, http.StatusOK, mangas)
	return nil
}

func (h *MangaHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var m models.Manga
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		return apperror.NewBadRequestError("Ошибка декодирования запроса", err)
	}

	if m.Title == "" {
		return apperror.NewValidationError("Поле title не может быть пустым", map[string]string{"title": "Это поле обязательно"})
	}

	id, err := h.Repo.Create(&m)
	if err != nil {
		return apperror.NewDatabaseError("Ошибка создания манги", err)
	}
	m.ID = id

	_ = h.Cache.Delete(r.Context(), "manga:list")

	response.Success(w, http.StatusCreated, m)
	return nil
}

func (h *MangaHandler) Detail(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/manga/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID", err)
	}

	key := fmt.Sprintf("manga:%d", id)
	if h.Cache != nil {
		cached, err := h.Cache.Get(r.Context(), key)
		if err == nil && cached != "" {
			h.Logger.Info("Cache hit", "id", id)

			var manga models.Manga
			if err = json.Unmarshal([]byte(cached), &manga); err != nil {
				h.Logger.Error("Ошибка десериализации из кэша", "err", err)
			} else {
				if h.Analytics != nil {
					views, err := h.Analytics.GetMangaView(r.Context(), id)
					if err != nil {
						mangaWithViews := analytics.MangaWithViews{
							ID:          manga.ID,
							Title:       manga.Title,
							Description: manga.Description,
							Views:       views,
						}
						response.Success(w, http.StatusOK, mangaWithViews)
						return nil
					}
				}
				response.Success(w, http.StatusOK, manga)
				return nil
			}
		}
		h.Logger.Info("Cache miss", "id", id)
	}

	m, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Манга не найдена", err)
	}

	if h.Cache != nil {
		jsonData, err := json.Marshal(m)
		if err == nil {
			if err = h.Cache.Set(r.Context(), key, string(jsonData), 5*time.Minute); err != nil {
				h.Logger.Error("Ошибка записи в кэш", "err", err)
			}
		}
	}

	var views int64 = 0
	if h.Analytics != nil {
		if err = h.Analytics.RecordMangaView(r.Context(), id); err != nil {
			h.Logger.Error("Ошибка записи просмотра манги", "err", err, "manga_id", id)
		}

		views, err = h.Analytics.GetMangaView(r.Context(), id)
		if err != nil {
			h.Logger.Error("Ошибка получения счетчика просмотров", "err", err, "manga_id", id)
		}
	}

	mangaWithViews := analytics.MangaWithViews{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Views:       views,
	}

	response.Success(w, http.StatusOK, mangaWithViews)
	return nil
}

func (h *MangaHandler) GetPopular(w http.ResponseWriter, r *http.Request) error {
	period := r.URL.Query().Get("period")
	limitStr := r.URL.Query().Get("limit")

	if period == "" {
		period = "all"
	}
	limit := int64(10)
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	cacheKey := fmt.Sprintf("manga:popular:%s:%d", period, limit)
	cachedDate, err := h.Cache.Get(r.Context(), cacheKey)
	if err == nil && cachedDate != "" {
		h.Logger.Info("Cache hit for popular manga", "period", period, "limit", limit)

		var result []analytics.MangaWithViews
		if err = json.Unmarshal([]byte(cachedDate), &result); err != nil {
			h.Logger.Error("Ошибка десериализации популярной манги из кеша", "err", err)
		} else {
			response.Success(w, http.StatusOK, result)
			return nil
		}
	}

	if h.Analytics != nil {
		return apperror.NewInternalServerError("Сервис аналитики недоступен", nil)
	}

	topEntries, err := h.Analytics.GetTopManga(r.Context(), period, limit)
	if err != nil {
		return apperror.NewInternalServerError("Ошибка получения рейтинга манги", err)
	}

	var result []analytics.MangaWithViews
	for _, entry := range topEntries {
		manga, err := h.Repo.GetByID(entry.MangaID)
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

	jsonData, err := json.Marshal(result)
	if err == nil {
		var ttl time.Duration
		switch period {
		case "daily":
			ttl = 1 * time.Hour
		case "weekly":
			ttl = 4 * time.Hour
		case "monthly":
			ttl = 12 * time.Hour
		default:
			ttl = 24 * time.Hour
		}

		if err = h.Cache.Set(r.Context(), cacheKey, string(jsonData), ttl); err != nil {
			h.Logger.Error("Ошибка кеширования популярной манги", "err", err)
		}
	}

	response.Success(w, http.StatusOK, result)
	return nil
}
