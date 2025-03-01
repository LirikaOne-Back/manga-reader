package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"manga-reader/internal/apperror"
	"manga-reader/internal/db"
	"manga-reader/internal/response"
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

func (h *PageHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/page/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", nil)
	}

	page, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewDatabaseError("Ошибка получения страницы", err)
	}

	if err = h.Repo.Delete(id); err != nil {
		return apperror.NewDatabaseError("Ошибка удаления страницы", err)
	}

	if err = os.Remove(page.ImagePath); err != nil {
		return apperror.NewInternalServerError("Ошибка удаления файла страницы", err)
	}

	response.Success(w, http.StatusNoContent, nil)
	return nil
}

func (h *PageHandler) UploadImage(w http.ResponseWriter, r *http.Request) error {
	// Устанавливаем максимальный размер файла (10 МБ)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return apperror.NewBadRequestError("Ошибка допустимого размера файла", err)
	}

	chapterIDStr := r.FormValue("chapter_id")
	if chapterIDStr == "" {
		return apperror.NewBadRequestError("Ошибка получения id главы", nil)
	}

	chapterID, err := strconv.ParseInt(chapterIDStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный id главы", err)
	}

	numberStr := r.FormValue("number")
	if numberStr == "" {
		return apperror.NewBadRequestError("Ошибка получения номера страницы", nil)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный номер страницы", err)
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		return apperror.NewBadRequestError("Не удалось загрузить файл", err)
	}
	defer file.Close()

	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return apperror.NewBadRequestError("Файл должен быть изображением", nil)
	}

	uploadDir := fmt.Sprintf("uploads/chapters/%d", chapterID)
	if err = os.MkdirAll(uploadDir, 0755); err != nil {
		return apperror.NewInternalServerError("Ошибка создания директории", err)
	}

	filename := fmt.Sprintf("%d_%d%s", chapterID, number, filepath.Ext(handler.Filename))
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return apperror.NewInternalServerError("Ошибка создания файла", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		return apperror.NewInternalServerError("Ошибка копирования файла", err)
	}

	page := &models.Page{
		ChapterID: chapterID,
		Number:    number,
		ImagePath: filePath,
	}

	id, err := h.Repo.Create(page)
	if err != nil {
		os.Remove(filePath)
		return apperror.NewValidationError("Ошибка сохранения страницы в БД", err)
	}

	page.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jsonData, err := json.Marshal(page)
	if err != nil {
		return apperror.NewInternalServerError("Ошибка кодирования ответа", err)
	}

	if _, err = w.Write(jsonData); err != nil {
		return apperror.NewInternalServerError("Ошибка записи ответа", err)
	}
	response.Success(w, http.StatusCreated, nil)
	return nil
}

func (h *PageHandler) ListByChapter(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path

	pathParts := strings.Split(strings.TrimPrefix(path, "/pages/chapter/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		return apperror.NewBadRequestError("Неверный формат URL", nil)
	}

	chapterIDStr := pathParts[0]

	chapterId, err := strconv.ParseInt(chapterIDStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Ошибка получения ID главы", err)
	}

	pages, err := h.Repo.ListByChapter(chapterId)
	if err != nil {
		return apperror.NewInternalServerError("Ошибка получения списка страниц", err)
	}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(pages); err != nil {
		return apperror.NewInternalServerError("Ошибка отправки страниц", err)
	}
	response.Success(w, http.StatusOK, nil)
	return nil
}

func (h *PageHandler) ServeImage(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/page/image/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Неверный формат id страницы", err)
	}

	page, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Ошибка получения страницы", err)
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
	return nil
}
