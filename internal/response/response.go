package response

import (
	"encoding/json"
	"log/slog"
	"manga-reader/internal/apperror"
	"net/http"
	"os"
)

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   interface{} `json:"error"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("Ошибка кодирования ответа", "err", err)
		}
	}
}

func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	resp := SuccessResponse{
		Success: true,
		Data:    data,
	}
	JSON(w, statusCode, resp)
}

func Error(w http.ResponseWriter, logger *slog.Logger, err error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	var statusCode int
	var errorResp interface{}

	if appErr, ok := err.(*apperror.AppError); ok {
		statusCode = appErr.StatusCode
		errorResp = appErr

		if appErr.Err != nil {
			logger.Error("Ошибка обработки запроса",
				"statusCode", statusCode,
				"errorCode", appErr.Code,
				"message", appErr.Message,
				"err", appErr.Err)
		} else {
			logger.Error("Ошибка обработки запроса",
				"statusCode", statusCode,
				"errorCode", appErr.Code,
				"message", appErr.Message)
		}
	} else {
		statusCode = http.StatusInternalServerError
		errorResp = map[string]string{
			"code":    apperror.ErrInternalServerError,
			"message": "Внутренняя ошибка сервера",
		}
		logger.Error("Неизвестная ошибка", "err", err)
	}

	resp := ErrorResponse{
		Success: false,
		Error:   errorResp,
	}
	JSON(w, statusCode, resp)
}
