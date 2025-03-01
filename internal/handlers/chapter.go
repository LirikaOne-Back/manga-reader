package handlers

import (
	"encoding/json"
	"log/slog"
	"manga-reader/internal/apperror"
	"manga-reader/internal/cache"
	"manga-reader/internal/db"
	"manga-reader/internal/response"
	"manga-reader/models"
	"net/http"
	"strconv"
	"strings"
)

type ChapterHandler struct {
	Repo   db.ChapterRepository
	Logger *slog.Logger
	Cache  *cache.RedisCache
}

func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", err)
	}

	if err = h.Repo.Delete(id); err != nil {
		return apperror.NewDatabaseError("Ошибка удаления главы", err)
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
	response.Success(w, http.StatusCreated, ch)
	return nil
}

func (h *ChapterHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", err)
	}

	var ch models.Chapter
	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		return apperror.NewBadRequestError("Ошибка декодирования запроса", err)
	}

	ch.ID = id
	if err = h.Repo.Update(&ch); err != nil {
		return apperror.NewDatabaseError("Ошибка обновления главы", err)
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

	ch, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Глава не найдена", err)
	}

	response.Success(w, http.StatusOK, ch)
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

	chapters, err := h.Repo.ListByManga(mangaID)
	if err != nil {
		return apperror.NewDatabaseError("Ошибка получения списка глав", err)
	}

	response.Success(w, http.StatusOK, chapters)
	return nil
}
