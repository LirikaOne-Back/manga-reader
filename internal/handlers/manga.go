package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
	Repo   db.MangaRepository
	Logger *slog.Logger
	Cache  cache.Cache
}

func (h *MangaHandler) List(w http.ResponseWriter, r *http.Request) error {
	mangas, err := h.Repo.List()
	if err != nil {
		return apperror.NewDatabaseError("Ошибка получения списка манги", err)
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
	response.Success(w, http.StatusOK, m)
	return nil
}
