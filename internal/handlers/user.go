package handlers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"manga-reader/internal/auth"
	"manga-reader/internal/db/sqlite"
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

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.Error("Ошибка хеширования пароля", "err", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	user := &models.User{Username: req.Username, Password: string(hashed)}
	id, err := h.UserRepo.Create(user)
	if err != nil {
		http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
		return
	}
	user.ID = id
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(user); err != nil {
		h.Logger.Error("Ошибка кодирования нового  пользователя")
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	user, err := h.UserRepo.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Не удалось сгенерировать токен", http.StatusInternalServerError)
		return
	}
	resp := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error("Ошибка декодирования токена", "err", err)
	}
}
