package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"manga-reader/internal/db"
	"manga-reader/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PageHandler struct {
	Repo   db.PageRepository
	Logger *slog.Logger
}

func (h *PageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/page/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	page, err := h.Repo.GetByID(id)
	if err != nil {
		h.Logger.Error("Ошибка получения страницы", "err", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if err = h.Repo.Delete(id); err != nil {
		h.Logger.Error("Ошибка удаления страницы из БД", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = os.Remove(page.ImagePath); err != nil {
		h.Logger.Error("Ошибка удаления файла изображения", "err", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем максимальный размер файла (10 МБ)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.Logger.Error("Ошибка при парсинге multipart формы", "err", err)
		http.Error(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	// Получаем chapterID и number из формы
	chapterIDStr := r.FormValue("chapter_id")
	if chapterIDStr == "" {
		http.Error(w, "Не указан chapter_id", http.StatusBadRequest)
		return
	}

	chapterID, err := strconv.ParseInt(chapterIDStr, 10, 64)
	if err != nil {
		h.Logger.Error("Некорректный chapter_id", "err", err)
		http.Error(w, "Некорректный chapter_id", http.StatusBadRequest)
		return
	}

	numberStr := r.FormValue("number")
	if numberStr == "" {
		http.Error(w, "Не указан номер страницы", http.StatusBadRequest)
		return
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		h.Logger.Error("Некорректный номер страницы", "err", err)
		http.Error(w, "Некорректный номер страницы", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		h.Logger.Error("Ошибка получения файла", "err", err)
		http.Error(w, "Не удалось загрузить файл", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "Файл должен быть изображением", http.StatusBadRequest)
		return
	}

	uploadDir := fmt.Sprintf("uploads/chapters/%d", chapterID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.Logger.Error("Ошибка создания директории", "err", err, "dir", uploadDir)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%d_%d%s", chapterID, number, filepath.Ext(handler.Filename))
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		h.Logger.Error("Ошибка создания файла", "err", err, "path", filePath)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		h.Logger.Error("Ошибка копирования файла", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	page := &models.Page{
		ChapterID: chapterID,
		Number:    number,
		ImagePath: filePath,
	}

	id, err := h.Repo.Create(page)
	if err != nil {
		h.Logger.Error("Ошибка сохранения страницы в БД", "err", err)
		os.Remove(filePath)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	page.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jsonData, err := json.Marshal(page)
	if err != nil {
		h.Logger.Error("Ошибка кодирования ответа", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(jsonData); err != nil {
		h.Logger.Error("Ошибка записи ответа", "err", err)
	}
}

func (h *PageHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	pathParts := strings.Split(strings.TrimPrefix(path, "/pages/chapter/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		h.Logger.Error("Неверный формат URL", "path", path)
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	chapterIDStr := pathParts[0]
	h.Logger.Info("Извлечен ID главы", "chapterIDStr", chapterIDStr)

	chapterId, err := strconv.ParseInt(chapterIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
		return
	}

	pages, err := h.Repo.ListByChapter(chapterId)
	if err != nil {
		h.Logger.Error("Ошибка получения списка страниц", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(pages); err != nil {
		h.Logger.Error("Ошибка отправки страниц", "err", err)
		return
	}
}

func (h *PageHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/page/image/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	page, err := h.Repo.GetByID(id)
	if err != nil {
		h.Logger.Error("Ошибка получения страницы", "err", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	contentType := "image/jpeg"
	if strings.HasSuffix(page.ImagePath, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(page.ImagePath, ".jpg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(page.ImagePath, ".webp") {
		contentType = "image/webp"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, page.ImagePath)
}
