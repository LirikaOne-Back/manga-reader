package handlers_test

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log/slog"
	"manga-reader/internal/db/sqlite"
	"manga-reader/internal/handlers"
	"manga-reader/internal/handlers/handlers_test/helper"
	"manga-reader/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestUserRepo(t *testing.T) *sqlite.SQLiteUserRepository {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Ошибка открытия in-memory базы: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := sqlite.NewSQLiteUserRepository(db, logger)
	return repo
}

func TestUserHandler_RegisterAndLogin(t *testing.T) {
	userRepo := setupTestUserRepo(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
		Logger:   testLogger,
	}

	// Тест регистрации
	regBody := `{"username": "testuser", "password": "secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/user/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp := httptest.NewRecorder()

	err := userHandler.Register(regResp, regReq)
	if err != nil {
		t.Fatalf("Неожиданная ошибка при регистрации: %v", err)
	}

	if regResp.Code != http.StatusCreated {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusCreated, regResp.Code)
	}

	var registeredUser models.User
	if err := helper.ExtractData(regResp.Body, &registeredUser); err != nil {
		t.Fatalf("Ошибка парсинга ответа регистрации: %v", err)
	}
	if registeredUser.Username != "testuser" {
		t.Errorf("Ожидалось имя пользователя %q, получено %q", "testuser", registeredUser.Username)
	}
	if registeredUser.ID == 0 {
		t.Error("Ожидался валидный ID, получен 0")
	}

	// Тест логина
	loginBody := `{"username": "testuser", "password": "secret123"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()

	err = userHandler.Login(loginResp, loginReq)
	if err != nil {
		t.Fatalf("Неожиданная ошибка при логине: %v", err)
	}

	if loginResp.Code != http.StatusOK {
		t.Fatalf("Ожидался статус %d, получен %d", http.StatusOK, loginResp.Code)
	}

	var loginResult struct {
		Token string `json:"token"`
	}
	if err := helper.ExtractData(loginResp.Body, &loginResult); err != nil {
		t.Fatalf("Ошибка парсинга ответа логина: %v", err)
	}
	if loginResult.Token == "" {
		t.Error("Ожидался непустой токен")
	}
}
