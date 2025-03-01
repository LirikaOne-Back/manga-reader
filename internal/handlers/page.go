package handlers

import (
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
		return apperror.NewBadRequestError("Некорректный ID страницы", err)
	}

	page, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Страница не найдена", err)
	}

	if err = h.Repo.Delete(id); err != nil {
		return apperror.NewDatabaseError("Ошибка удаления страницы из БД", err)
	}

	if err = os.Remove(page.ImagePath); err != nil {
		h.Logger.Error("Ошибка удаления файла изображения", "err", err)
		// Не возвращаем ошибку, так как запись из БД уже удалена
	}

	response.Success(w, http.StatusNoContent, nil)
	return nil
}

func (h *PageHandler) UploadImage(w http.ResponseWriter, r *http.Request) error {
	// Устанавливаем максимальный размер файла (10 МБ)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return apperror.NewBadRequestError("Ошибка при парсинге multipart формы", err)
	}

	// Получаем chapterID и number из формы
	chapterIDStr := r.FormValue("chapter_id")
	if chapterIDStr == "" {
		return apperror.NewValidationError("Поле chapter_id не может быть пустым",
			map[string]string{"chapter_id": "Это поле обязательно"})
	}

	chapterID, err := strconv.ParseInt(chapterIDStr, 10, 64)
	if err != nil {
		return apperror.NewValidationError("Некорректный chapter_id",
			map[string]string{"chapter_id": "Должно быть целое число"})
	}

	numberStr := r.FormValue("number")
	if numberStr == "" {
		return apperror.NewValidationError("Поле number не может быть пустым",
			map[string]string{"number": "Это поле обязательно"})
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return apperror.NewValidationError("Некорректный номер страницы",
			map[string]string{"number": "Должно быть целое число"})
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
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
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
		// Если не удалось создать запись в БД, удаляем загруженный файл
		os.Remove(filePath)
		return apperror.NewDatabaseError("Ошибка сохранения страницы в БД", err)
	}

	page.ID = id
	response.Success(w, http.StatusCreated, page)
	return nil
}

func (h *PageHandler) ListByChapter(w http.ResponseWriter, r *http.Request) error {
	// Использование улучшенного метода извлечения ID главы
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/pages/chapter/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		return apperror.NewBadRequestError("Некорректный формат URL", nil)
	}

	chapterIDStr := pathParts[0]
	chapterID, err := strconv.ParseInt(chapterIDStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID главы", err)
	}

	pages, err := h.Repo.ListByChapter(chapterID)
	if err != nil {
		return apperror.NewDatabaseError("Ошибка получения списка страниц", err)
	}

	response.Success(w, http.StatusOK, pages)
	return nil
}

func (h *PageHandler) ServeImage(w http.ResponseWriter, r *http.Request) error {
	idStr := strings.TrimPrefix(r.URL.Path, "/page/image/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apperror.NewBadRequestError("Некорректный ID страницы", err)
	}

	page, err := h.Repo.GetByID(id)
	if err != nil {
		return apperror.NewNotFoundError("Страница не найдена", err)
	}

	contentType := "image/jpeg"
	if strings.HasSuffix(page.ImagePath, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(page.ImagePath, ".jpg") || strings.HasSuffix(page.ImagePath, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(page.ImagePath, ".webp") {
		contentType = "image/webp"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, page.ImagePath)
	return nil
}
