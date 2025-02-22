package handlers

import (
	"encoding/json"
	"log/slog"
	"manga-reader/internal/cache"
	"manga-reader/internal/db"
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

func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}
	if err = h.Repo.Delete(id); err != nil {
		h.Logger.Error("Ошибка удаления манги", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChapterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ch models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		h.Logger.Error("Ошибка декодирования запроса", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	id, err := h.Repo.Create(&ch)
	if err != nil {
		h.Logger.Error("Ошибка создания главы", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ch.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(ch); err != nil {
		h.Logger.Error("Ошибка кодирования ответа", "err", err)
	}
}

func (h *ChapterHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}
	var ch models.Chapter
	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		h.Logger.Error("Ошибка декодирования запроса", "err", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	ch.ID = id
	if err = h.Repo.Update(&ch); err != nil {
		h.Logger.Error("Ошибка обновления главы", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChapterHandler) GetById(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/chapter/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}
	ch, err := h.Repo.GetByID(id)
	if err != nil {
		h.Logger.Error("Ошибка получения главы", "err", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	jsonData, err := json.Marshal(ch)
	if err != nil {
		h.Logger.Error("Ошибка маршелинга JSON", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(jsonData); err != nil {
		h.Logger.Error("Ошибка отпривки ответа", "err", err)
	}
}

func (h *ChapterHandler) ListByManga(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	mangaID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid manga ID", http.StatusBadRequest)
		return
	}
	chapters, err := h.Repo.ListByManga(mangaID)
	if err != nil {
		h.Logger.Error("Ошибка получения списка глав", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(chapters); err != nil {
		h.Logger.Error("Ошибка отправки глав", "err", err)
	}
}
