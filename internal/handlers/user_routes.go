package handlers

import (
	"manga-reader/internal/apperror"
	"manga-reader/internal/middleware"
	"net/http"
)

func RegisterUserRoutes(mux *http.ServeMux, uh *UserHandler) {
	mux.HandleFunc("/user/register", middleware.ErrorHandler(uh.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
		return uh.Register(w, r)
	}))
	mux.HandleFunc("/user/login", middleware.ErrorHandler(uh.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
		return uh.Login(w, r)
	}))
}
