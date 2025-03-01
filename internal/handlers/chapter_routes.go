package handlers

import (
	"manga-reader/internal/apperror"
	"manga-reader/internal/middleware"
	"net/http"
)

func RegisterChapterRoutes(mux *http.ServeMux, ch *ChapterHandler) {
	mux.HandleFunc("/chapter", middleware.ErrorHandler(ch.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return ch.Create(w, r)
		} else {
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))
	mux.HandleFunc("/chapter/", middleware.ErrorHandler(ch.Logger, func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return ch.GetById(w, r)
		case http.MethodPut:
			return ch.Update(w, r)
		case http.MethodDelete:
			return ch.Delete(w, r)
		default:
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))
}
