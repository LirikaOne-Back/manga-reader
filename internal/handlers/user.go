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
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return apperror.NewInternalServerError("Внутренняя ошибка сервера", err)
	}
	user := &models.User{Username: req.Username, Password: string(hashed)}
	id, err := h.UserRepo.Create(user)
	if err != nil {
		return apperror.NewInternalServerError("Не удалось создать пользователя", err)
	}
	user.ID = id
	user.Password = ""
	if err = json.NewEncoder(w).Encode(user); err != nil {
		return apperror.NewInternalServerError("Ошибка кодирования нового  пользователя", err)
	}
	response.Success(w, http.StatusCreated, nil)
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
	user, err := h.UserRepo.GetByUsername(req.Username)
	if err != nil {
		return apperror.NewUnauthorizedError("Неверное имя пользователя или пароль", err)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return apperror.NewUnauthorizedError("Неверное имя пользователя или пароль", err)
	}
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return apperror.NewInternalServerError("Не удалось сгенерировать токен", err)
	}
	resp := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		return apperror.NewInternalServerError("Ошибка декодирования токена", err)
	}
	response.Success(w, http.StatusOK, resp)
	return nil
}
