package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"manga-reader/internal/handlers"
	"manga-reader/models"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type MockMangaRepository struct {
	mu     sync.Mutex
	mangas map[int64]*models.Manga
	nextID int64
}

func NewMockMangaRepository() *MockMangaRepository {
	return &MockMangaRepository{
		mangas: make(map[int64]*models.Manga),
		nextID: 1,
	}
}

func (m *MockMangaRepository) Create(manga *models.Manga) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	manga.ID = m.nextID
	m.nextID++
	m.mangas[manga.ID] = manga
	return manga.ID, nil
}

func (m *MockMangaRepository) GetByID(id int64) (*models.Manga, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	manga, ok := m.mangas[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return manga, nil
}

func (m *MockMangaRepository) List() ([]*models.Manga, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var mangas []*models.Manga
	for _, manga := range m.mangas {
		mangas = append(mangas, manga)
	}
	return mangas, nil
}

func (m *MockMangaRepository) Update(manga *models.Manga) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.mangas[manga.ID]; !ok {
		return errors.New("not found")
	}
	m.mangas[manga.ID] = manga
	return nil
}

func (m *MockMangaRepository) Delete(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.mangas[id]; !ok {
		return errors.New("not found")
	}
	delete(m.mangas, id)
	return nil
}

type DummyRedisCache struct{}

func (d *DummyRedisCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (d *DummyRedisCache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return nil
}

func TestMangaHandler_CreateAndGet(t *testing.T) {
	mockRepo := NewMockMangaRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	dummyCache := &DummyRedisCache{}

	mangaHandler := &handlers.MangaHandler{
		Repo:   mockRepo,
		Logger: testLogger,
		Cache:  dummyCache,
	}
	// Тест создания манги (POST /manga)
	createBody := `{"title": "Naruto", "description": "Ниндзя приключения"}`
	createReq := httptest.NewRequest(http.MethodPost, "/manga", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()

	mangaHandler.Create(createResp, createReq)

	if createResp.Code != http.StatusCreated {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusCreated, createResp.Code)
	}

	var created models.Manga
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("Ошибка парсинга ответа: %v", err)
	}
	if created.ID == 0 {
		t.Error("Ожидался валидный ID, получен 0")
	}

	// Тест получения деталей манги (GET /manga/{id})
	getURL := fmt.Sprintf("/manga/%d", created.ID)
	getReq := httptest.NewRequest(http.MethodGet, getURL, nil)
	getResp := httptest.NewRecorder()

	mangaHandler.Detail(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusOK, getResp.Code)
	}

	var fetched models.Manga
	if err := json.Unmarshal(getResp.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("Ошибка парсинга get-ответа: %v", err)
	}
	if fetched.ID == 0 {
		t.Errorf("Ожидался заголовок %q, получен %q", created.Title, fetched.Title)
	}
}

func TestMangaHandler_List(t *testing.T) {
	mockRepo := NewMockMangaRepository()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dummyCache := &DummyRedisCache{}
	mangaHandler := &handlers.MangaHandler{
		Repo:   mockRepo,
		Logger: testLogger,
		Cache:  dummyCache,
	}

	for i := 0; i < 3; i++ {
		title := fmt.Sprintf("Manga %d", i)
		_, _ = mockRepo.Create(&models.Manga{Title: title, Description: "Desc"})
	}

	listReq := httptest.NewRequest(http.MethodGet, "/manga", nil)
	listResp := httptest.NewRecorder()
	mangaHandler.List(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusOK, listResp.Code)
	}
	var mangas []*models.Manga
	if err := json.Unmarshal(listResp.Body.Bytes(), &mangas); err != nil {
		t.Fatalf("Ошибка парсинга списка: %v", err)
	}
	if len(mangas) != 3 {
		t.Errorf("Ожидалось 3 манги, получено %d", len(mangas))
	}
}
