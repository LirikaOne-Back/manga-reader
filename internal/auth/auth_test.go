package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateAndParseToken(t *testing.T) {
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
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	handler := AuthMiddleware(dummyHandler)
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
