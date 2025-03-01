package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGenerateAndParseToken(t *testing.T) {
	SetJWTSecret("test-secret")

	token, err := GenerateToken(42)
	if err != nil {
		t.Fatalf("Ошибка генерации токена: %v", err)
	}
	userID, err := ParseToken(token)
	if err != nil {
		t.Fatalf("Ошибка парсинга токена: %v", err)
	}
	if userID != 42 {
		t.Errorf("Ожидался user_id 42, получен %d", userID)
	}
}

func TestAuthMiddleware(t *testing.T) {
	SetJWTSecret("test-secret")

	testLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Store the logger in the request context
			ctx := context.WithValue(r.Context(), "logger", testLogger)
			AuthMiddleware(handler).ServeHTTP(w, r.WithContext(ctx))
		})
	}(originalHandler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusUnauthorized, rr.Code)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidToken")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusUnauthorized, rr.Code)
	}

	token, _ := GenerateToken(42)
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "OK") {
		t.Errorf("Ожидался ответ OK, получен %s", rr.Body.String())
	}
}
