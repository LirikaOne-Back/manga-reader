package handlers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"manga-reader/internal/apperror"
	"manga-reader/internal/auth"
	"manga-reader/internal/db/sqlite"
	"manga-reader/internal/response"
	"manga-reader/models"
	"net/http"
)

type UserHandler struct {
	UserRepo *sqlite.SQLiteUserRepository
	Logger   *slog.Logger
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apperror.NewBadRequestError("Не удалось прочитать тело запроса", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return apperror.NewBadRequestError("Неверный формат запроса", err)
	}

	// Валидация
	if req.Username == "" {
		return apperror.NewValidationError("Поле username не может быть пустым",
			map[string]string{"username": "Это поле обязательно"})
	}
	if req.Password == "" {
		return apperror.NewValidationError("Поле password не может быть пустым",
			map[string]string{"password": "Это поле обязательно"})
	}
	if len(req.Password) < 6 {
		return apperror.NewValidationError("Пароль слишком короткий",
			map[string]string{"password": "Минимальная длина - 6 символов"})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apperror.NewInternalServerError("Ошибка хеширования пароля", err)
	}

	user := &models.User{Username: req.Username, Password: string(hashed)}
	id, err := h.UserRepo.Create(user)
	if err != nil {
		return apperror.NewDatabaseError("Не удалось создать пользователя", err)
	}

	user.ID = id
	user.Password = "" // Не возвращаем пароль в ответе

	response.Success(w, http.StatusCreated, user)
	return nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperror.NewBadRequestError("Неверный формат запроса", err)
	}

	// Валидация
	if req.Username == "" || req.Password == "" {
		return apperror.NewValidationError("Имя пользователя и пароль обязательны", nil)
	}

	user, err := h.UserRepo.GetByUsername(req.Username)
	if err != nil {
		return apperror.NewUnauthorizedError("Неверное имя пользователя или пароль", nil)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return apperror.NewUnauthorizedError("Неверное имя пользователя или пароль", nil)
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return apperror.NewInternalServerError("Не удалось сгенерировать токен", err)
	}

	resp := LoginResponse{Token: token}
	response.Success(w, http.StatusOK, resp)
	return nil
}
