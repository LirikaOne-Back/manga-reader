package handlers

import (
	"manga-reader/internal/apperror"
	"manga-reader/internal/middleware"
	"net/http"
)

func RegisterPageRoutes(mux *http.ServeMux, ph *PageHandler) {
	mux.HandleFunc("/page/upload", middleware.ErrorHandler(ph.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return ph.UploadImage(w, r)
		} else {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))

	mux.HandleFunc("/pages/chapter/", middleware.ErrorHandler(ph.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodGet {
			return ph.ListByChapter(w, r)
		} else {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))

	mux.HandleFunc("/page/", middleware.ErrorHandler(ph.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodDelete {
			return ph.Delete(w, r)
		} else {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))

	mux.HandleFunc("/page/image/", middleware.ErrorHandler(ph.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodGet {
			return ph.ServeImage(w, r)
		} else {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))
}
