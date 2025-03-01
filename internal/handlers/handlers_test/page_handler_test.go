package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"manga-reader/internal/handlers"
	"manga-reader/internal/handlers/handlers_test/helper"
	"manga-reader/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type MockPageRepository struct {
	mu     *sync.Mutex
	pages  map[int64]*models.Page
	nextID int64
}

func NewMockPageRepository() *MockPageRepository {
	return &MockPageRepository{
		mu:     &sync.Mutex{},
		pages:  make(map[int64]*models.Page),
		nextID: 1,
	}
}

func (r *MockPageRepository) Create(p *models.Page) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.ID = r.nextID
	r.nextID++
	r.pages[p.ID] = p
	return p.ID, nil
}

func (r *MockPageRepository) GetByID(id int64) (*models.Page, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.pages[id]
	if !ok {
		return nil, errors.New("page not found")
	}
	return p, nil
}

func (r *MockPageRepository) ListByChapter(chapterID int64) ([]*models.Page, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var pages []*models.Page
	for _, page := range r.pages {
		if page.ChapterID == chapterID {
			pages = append(pages, page)
		}
	}
	return pages, nil
}

func (r *MockPageRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.pages[id]; !ok {
		return errors.New("page not found")
	}
	delete(r.pages, id)
	return nil
}

func (r *MockPageRepository) Update(p *models.Page) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.pages[p.ID]; !ok {
		return errors.New("page not found")
	}
	r.pages[p.ID] = p
	return nil
}

func createTestImage(t *testing.T) string {
	tempDir := t.TempDir()

	imagePath := filepath.Join(tempDir, "test-image.jpg")

	jpegBytes := []byte{
		0xFF, 0xD8,
		0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
		0xFF, 0xDB, 0x00, 0x43, 0x00,
		0xFF, 0xC0, 0x00, 0x11, 0x08, 0x00, 0x01, 0x00, 0x01, 0x03, 0x01, 0x22, 0x00, 0x02, 0x11, 0x01, 0x03, 0x11, 0x01,
		0xFF, 0xC4, 0x00, 0x1F, 0x00,
		0xFF, 0xDA, 0x00, 0x0C, 0x03, 0x01, 0x00, 0x02, 0x11, 0x03, 0x11, 0x00, 0x3F, 0x00,
		0xFF, 0xD9,
	}

	if err := os.WriteFile(imagePath, jpegBytes, 0644); err != nil {
		t.Fatalf("Не удалось создать тестовое изображение: %v", err)
	}

	return imagePath
}

func createMultipartRequest(t *testing.T, imagePath, url string, fields map[string]string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("Не удалось добавить поле %s: %v", key, err)
		}
	}

	file, err := os.Open(imagePath)
	if err != nil {
		t.Fatalf("Не удалось открыть тестовое изображение: %v", err)
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		"image", filepath.Base(imagePath)))
	h.Set("Content-Type", "image/jpeg")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("Не удалось создать часть формы для файла: %v", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		t.Fatalf("Не удалось скопировать содержимое файла: %v", err)
	}

	if err = writer.Close(); err != nil {
		t.Fatalf("Не удалось завершить форму: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestPageHandler_UploadImage(t *testing.T) {
	os.RemoveAll("uploads")

	imagePath := createTestImage(t)

	mockRepo := NewMockPageRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pageHandler := &handlers.PageHandler{
		Repo:   mockRepo,
		Logger: testLogger,
	}

	fields := map[string]string{
		"chapter_id": "1",
		"number":     "1",
	}

	absPath, _ := filepath.Abs(imagePath)
	t.Logf("Путь к тестовому изображению: %s", absPath)

	req := createMultipartRequest(t, imagePath, "/page/upload", fields)

	t.Logf("Content-Type: %s", req.Header.Get("Content-Type"))

	resp := httptest.NewRecorder()

	pageHandler.UploadImage(resp, req)

	t.Logf("Тело ответа: %s", resp.Body.String())

	if resp.Code != http.StatusCreated {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusCreated, resp.Code)
	}

	if !json.Valid(resp.Body.Bytes()) {
		t.Fatalf("Получен невалидный JSON: %s", resp.Body.String())
	}

	page := models.Page{}
	if err := json.Unmarshal(resp.Body.Bytes(), &page); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	if page.ID == 0 {
		t.Error("Ожидался валидный ID, получен 0")
	}
	if page.ChapterID != 1 {
		t.Errorf("Ожидался ChapterID 1, получен %d", page.ChapterID)
	}
	if page.Number != 1 {
		t.Errorf("Ожидался Number 1, получен %d", page.Number)
	}
	if page.ImagePath == "" {
		t.Error("Ожидался непустой ImagePath")
	}

	if _, err := os.Stat(page.ImagePath); os.IsNotExist(err) {
		t.Errorf("Файл изображения не был создан по пути %s", page.ImagePath)
	} else {
		os.Remove(page.ImagePath)
		os.Remove(filepath.Dir(page.ImagePath))
	}
}

func TestPageHandler_ListByChapter(t *testing.T) {
	mockRepo := NewMockPageRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pageHandler := &handlers.PageHandler{
		Repo:   mockRepo,
		Logger: testLogger,
	}

	chapterID := int64(1)
	for i := 1; i <= 3; i++ {
		page := &models.Page{
			ChapterID: chapterID,
			Number:    i,
			ImagePath: fmt.Sprintf("/path/to/image_%d.jpg", i),
		}
		_, err := mockRepo.Create(page)
		if err != nil {
			t.Fatalf("Ошибка при создании тестовой страницы: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/pages/chapter/%d", chapterID), nil)
	resp := httptest.NewRecorder()

	err := pageHandler.ListByChapter(resp, req)
	if err != nil {
		t.Fatalf("Неожиданная ошибка при получении списка страниц: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, resp.Code)
	}

	var pages []*models.Page
	if err := helper.ExtractData(resp.Body, &pages); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	if len(pages) != 3 {
		t.Errorf("Ожидалось 3 страницы, получено %d", len(pages))
	}

	for _, page := range pages {
		if page.ChapterID != chapterID {
			t.Errorf("Страница %d имеет неверный ChapterID %d (ожидался %d)", page.ID, page.ChapterID, chapterID)
		}
	}
}

func TestPageHandler_UploadImage_InvalidImagePath(t *testing.T) {
	imagePath := createTestImage(t)

	mockRepo := NewMockPageRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	pageHandler := &handlers.PageHandler{
		Repo:   mockRepo,
		Logger: testLogger,
	}

	page := &models.Page{
		ChapterID: 1,
		Number:    1,
		ImagePath: imagePath,
	}
	id, err := mockRepo.Create(page)
	if err != nil {
		t.Fatalf("Ошибка при создании тестовой страницы: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/page/%d", id), nil)
	req.URL.Path = fmt.Sprintf("/page/%d", id)
	resp := httptest.NewRecorder()

	pageHandler.Delete(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusNoContent, resp.Code)
	}

	_, err = mockRepo.GetByID(id)
	if err == nil {
		t.Error("Страница не была удалена из репозитория")
	}
}

func TestPageHandler_ServeImage(t *testing.T) {
	imagePath := createTestImage(t)

	mockRepo := NewMockPageRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	pageHandler := &handlers.PageHandler{
		Repo:   mockRepo,
		Logger: testLogger,
	}

	page := &models.Page{
		ChapterID: 1,
		Number:    1,
		ImagePath: imagePath,
	}
	id, err := mockRepo.Create(page)
	if err != nil {
		t.Fatalf("Ошибка при создании тестовой страницы: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/page/image/%d", id), nil)
	req.URL.Path = fmt.Sprintf("/page/image/%d", id)
	resp := httptest.NewRecorder()

	pageHandler.ServeImage(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, resp.Code)
	}

	contentType := resp.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		t.Errorf("Ожидался Content-Type image/*, получен %s", contentType)
	}

	if resp.Body.Len() == 0 {
		t.Error("Ответ не содержит данных изображения")
	}

	responseBytes := resp.Body.Bytes()
	if len(responseBytes) < 2 || responseBytes[0] != 0xFF || responseBytes[1] != 0xD8 {
		t.Error("Ответ не содержит маркер начала JPEG")
	}
}
