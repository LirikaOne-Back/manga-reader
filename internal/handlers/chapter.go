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

type ChapterHandler struct {
	Repo      db.ChapterRepository
	Logger    *slog.Logger
	Cache     *cache.RedisCache
	Analytics *analytics.AnalyticsService
}

func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", err)
	}

	chapter, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Глава не найдена", err)
	}

	if err = h.Repo.Delete(id); err != nil {
		return apperror.NewDatabaseError("Ошибка удаления главы", err)
	}

	cacheKey := fmt.Sprintf("manga:%d:chapters", chapter.MangaID)
	if h.Cache != nil {
		if err = h.Cache.Delete(r.Context(), cacheKey); err != nil {
			h.Logger.Error("Ошибка инвалидации кеша", "key", cacheKey, "err", err)
		}
	}

	response.Success(w, http.StatusNoContent, nil)
	return nil
}

func (h *ChapterHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var ch models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		return apperror.NewBadRequestError("Ошибка декодирования запроса", err)
	}

	if ch.Title == "" {
		return apperror.NewValidationError("Поле title не может быть пустым",
			map[string]string{"title": "Это поле обязательно"})
	}
	if ch.MangaID <= 0 {
		return apperror.NewValidationError("Некорректный ID манги",
			map[string]string{"manga_id": "Должен быть положительным числом"})
	}

	id, err := h.Repo.Create(&ch)
	if err != nil {
		return apperror.NewDatabaseError("Ошибка создания главы", err)
	}

	ch.ID = id

	cacheKey := fmt.Sprintf("manga:%d:chapters", ch.MangaID)
	if h.Cache != nil {
		if err = h.Cache.Delete(r.Context(), cacheKey); err != nil {
			h.Logger.Error("Ошибка инвалидации кеша", "key", cacheKey, "err", err)
		}
	}

	response.Success(w, http.StatusCreated, ch)
	return nil
}

func (h *ChapterHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", err)
	}

	oldChapter, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Глава не найдена", err)
	}

	var ch models.Chapter
	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		return apperror.NewBadRequestError("Ошибка декодирования запроса", err)
	}

	ch.ID = id
	if err = h.Repo.Update(&ch); err != nil {
		return apperror.NewDatabaseError("Ошибка обновления главы", err)
	}

	cacheKey := fmt.Sprintf("manga:%d:chapters", oldChapter.MangaID)
	if h.Cache != nil {
		if err = h.Cache.Delete(r.Context(), cacheKey); err != nil {
			h.Logger.Error("Ошибка инвалидации кеша", "key", cacheKey, "err", err)
		}

		chapterKey := fmt.Sprintf("manga:%d:chapters", id)
		if err = h.Cache.Delete(r.Context(), chapterKey); err != nil {
			h.Logger.Error("Ошибка инвалидации кеша", "key", cacheKey, "err", err)
		}
	}

	response.Success(w, http.StatusNoContent, nil)
	return nil
}

func (h *ChapterHandler) GetById(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", err)
	}

	var ch *models.Chapter
	var mangaID int64

	cacheKey := fmt.Sprintf("chapter:%d", id)
	if h.Cache != nil {
		cachedData, err := h.Cache.Get(r.Context(), cacheKey)
		if err == nil && cachedData != "" {
			h.Logger.Info("Cache hit for chapter", "id", id)

			if err = json.Unmarshal([]byte(cachedData), &ch); err == nil {
				mangaID = ch.MangaID
			} else {
				h.Logger.Error("Ошибка десериализации главы из кеша", "err", err)
			}
		}
	}

	if ch == nil {
		ch, err = h.Repo.GetByID(id)
		if err != nil {
			return apperror.NewNotFoundError("Глава не найдена", err)
		}
		mangaID = ch.MangaID

		if h.Cache != nil {
			jsonData, err := json.Marshal(ch)
			if err == nil {
				if err = h.Cache.Set(r.Context(), cacheKey, string(jsonData), 30*time.Minute); err != nil {
					h.Logger.Error("Ошибка кеширования главы", "err", err)
				}
			}
		}
	}

	var views int64 = 0
	if h.Analytics != nil {
		if err = h.Analytics.RecordChapterView(r.Context(), id, mangaID); err != nil {
			h.Logger.Error("Ошибка записи просмотра главы", "err", err, "chapter_id", id)
		}

		chapterViewsKey := fmt.Sprintf("views:chapter:%d", id)
		viewsStr, err := h.Cache.Get(r.Context(), chapterViewsKey)
		if err == nil {
			views, _ = strconv.ParseInt(viewsStr, 10, 64)
		}
	}

	chapterWithViews := analytics.ChapterWithViews{
		ID:      ch.ID,
		MangaID: ch.MangaID,
		Number:  ch.Number,
		Title:   ch.Title,
		Views:   views,
	}

	response.Success(w, http.StatusOK, chapterWithViews)
	return nil
}

func (h *ChapterHandler) ListByManga(w http.ResponseWriter, r *http.Request) error {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		return apperror.NewBadRequestError("Некорректный URL", nil)
	}

	mangaID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID манги", err)
	}

	cacheKey := fmt.Sprintf("manga:%d:chapters", mangaID)
	if h.Cache != nil {
		cachedData, err := h.Cache.Get(r.Context(), cacheKey)
		if err == nil && cachedData != "" {
			h.Logger.Info("Cache hit for chapters list", "manga_id", mangaID)

			var chapters []models.Chapter
			if err = json.Unmarshal([]byte(cachedData), &chapters); err != nil {
				h.Logger.Error("")
			} else {
				response.Success(w, http.StatusOK, chapters)
				return nil
			}
		}
	}

	chapters, err := h.Repo.ListByManga(mangaID)
	if err != nil {
		return apperror.NewDatabaseError("Ошибка получения списка глав", err)
	}

	if h.Cache != nil {
		jsonData, err := json.Marshal(chapters)
		if err == nil {
			if err = h.Cache.Set(r.Context(), cacheKey, string(jsonData), 15*time.Minute); err != nil {
				h.Logger.Error("Ошибка кеширования списка глав", "err", err)
			}
		}
	}

	response.Success(w, http.StatusOK, chapters)
	return nil
}
