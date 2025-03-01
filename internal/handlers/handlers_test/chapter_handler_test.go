package handlers_test

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"manga-reader/internal/handlers"
	"manga-reader/internal/handlers/handlers_test/helper"
	"manga-reader/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type MockChapterRepository struct {
	mu       sync.Mutex
	chapters map[int64]*models.Chapter
	nextID   int64
}

func NewMockChapterRepository() *MockChapterRepository {
	return &MockChapterRepository{
		chapters: make(map[int64]*models.Chapter),
		nextID:   1,
	}
}

func (m *MockChapterRepository) Create(ch *models.Chapter) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch.ID = m.nextID
	m.nextID++
	m.chapters[ch.ID] = ch
	return ch.ID, nil
}

func (m *MockChapterRepository) GetByID(id int64) (*models.Chapter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.chapters[id]
	if !ok {
		return nil, errors.New("chapter not found")
	}
	return ch, nil
}

func (m *MockChapterRepository) Delete(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.chapters[id]; !ok {
		return errors.New("chapter not found")
	}
	delete(m.chapters, id)
	return nil
}

func (m *MockChapterRepository) Update(ch *models.Chapter) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.chapters[ch.ID]; !ok {
		return errors.New("chapter not found")
	}
	m.chapters[ch.ID] = ch
	return nil
}

func (m *MockChapterRepository) ListByManga(mangaID int64) ([]*models.Chapter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var chapters []*models.Chapter
	for _, ch := range m.chapters {
		if ch.MangaID == mangaID {
			chapters = append(chapters, ch)
		}
	}
	return chapters, nil
}

func TestChapterHandler_CreateAndGet(t *testing.T) {
	mockRepo := NewMockChapterRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	chapterHandler := &handlers.ChapterHandler{
		Repo:   mockRepo,
		Logger: testLogger,
	}

	// Тест создания главы (POST /chapter)
	createBody := `{"manga_id": 1, "number": 1, "title": "Глава 1: Начало"}`
	createReq := httptest.NewRequest(http.MethodPost, "/chapter", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()

	err := chapterHandler.Create(createResp, createReq)
	if err != nil {
		t.Fatalf("Неожиданная ошибка при создании главы: %v", err)
	}

	if createResp.Code != http.StatusCreated {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusCreated, createResp.Code)
	}

	var chapter models.Chapter
	if err := helper.ExtractData(createResp.Body, &chapter); err != nil {
		t.Fatalf("Ошибка парсинга ответа создания главы: %v", err)
	}
	if chapter.ID == 0 {
		t.Error("Ожидался валидный ID, получен 0")
	}

	// Тест получения главы по ID (GET /chapter/{id})
	getURL := fmt.Sprintf("/chapter/%d", chapter.ID)
	getReq := httptest.NewRequest(http.MethodGet, getURL, nil)
	getResp := httptest.NewRecorder()

	err = chapterHandler.GetById(getResp, getReq)
	if err != nil {
		t.Fatalf("Неожиданная ошибка при получении главы: %v", err)
	}

	if getResp.Code != http.StatusOK {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusOK, getResp.Code)
	}

	var fetched models.Chapter
	if err := helper.ExtractData(getResp.Body, &fetched); err != nil {
		t.Fatalf("Ошибка парсинга ответа получения главы: %v", err)
	}
	if fetched.Title != chapter.Title {
		t.Errorf("Ожидалось название %q, получено %q", chapter.Title, fetched.Title)
	}
}
