package handlers

import (
	"manga-reader/internal/response"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{"status": "OK"})
}
