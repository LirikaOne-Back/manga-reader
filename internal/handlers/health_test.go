package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("Ошибка создания запроса: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var successResp struct {
		Success bool `json:"success"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &successResp); err != nil {
		t.Fatalf("Ошибка парсинга ответа: %v", err)
	}

	if !successResp.Success {
		t.Error("Ожидалось поле success: true")
	}

	if successResp.Data.Status != "OK" {
		t.Errorf("Ожидался статус %q, получен %q", "OK", successResp.Data.Status)
	}
}
