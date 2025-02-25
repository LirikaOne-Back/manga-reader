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

func (h *PageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	chapterID, err := strconv.ParseInt(r.FormValue("chapter_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid chapter_id", http.StatusBadRequest)
		return
	}

	number, err := strconv.Atoi(r.FormValue("number"))
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
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

	uploadDir := fmt.Sprintf("upload/page/%d", chapterID)
	if err = os.MkdirAll(uploadDir, 0755); err != nil {
		h.Logger.Error("Ошибка создания директории", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%d_%d%s", chapterID, number, filepath.Ext(handler.Filename))
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		h.Logger.Error("Ошибка создания файла", "err", err)
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	page.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(page); err != nil {
		h.Logger.Error("Ошибка кодирования ответа", "err", err)
	}
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

func (h *PageHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	chapterId, err := strconv.ParseInt(parts[2], 10, 64)
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
