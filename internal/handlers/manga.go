package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"manga-reader/internal/cache"
	"manga-reader/internal/db"
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

func (h *MangaHandler) List(w http.ResponseWriter, r *http.Request) {
	mangas, err := h.Repo.List()
	if err != nil {
		h.Logger.Error("Ошибка получения списка манги", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mangas); err != nil {
		h.Logger.Error("Ошибка кодирования ответа", "err", err)
	}
}

func (h *MangaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var m models.Manga
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.Logger.Error("Ошибка декодирования запроса", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	id, err := h.Repo.Create(&m)
	if err != nil {
		h.Logger.Error("Ошибка создания манги", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	m.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(m); err != nil {
		h.Logger.Error("Ошибка кодирования ответа", "err", err)
	}
}

func (h *MangaHandler) Detail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/manga/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		key := fmt.Sprintf("manga:%d", id)
		if h.Cache != nil {
			cached, err := h.Cache.Get(r.Context(), key)
			if err == nil && cached != "" {
				h.Logger.Info("Cache hit", "id", id)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(cached))
				if err != nil {
					h.Logger.Error("Ошибка отправки кэшированного ответа", "err", err)
				}
				return
			}
			h.Logger.Info("Cache miss", "id", id)
		}

		m, err := h.Repo.GetByID(id)
		if err != nil {
			h.Logger.Error("Ошибка получения манги", "err", err)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		jsonData, err := json.Marshal(m)
		if err != nil {
			h.Logger.Error("Ошибка маршалинга JSON", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if h.Cache != nil {
			if err := h.Cache.Set(r.Context(), key, string(jsonData), 5*time.Minute); err != nil {
				h.Logger.Error("Ошибка записи в Redis", "err", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			h.Logger.Error("Ошибка отправки ответа", "err", err)
		}
	// Остальные методы...
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
