package handlers

import (
	"manga-reader/internal/apperror"
	"manga-reader/internal/middleware"
	"net/http"
	"strings"
)

func RegisterMangaRoutes(mux *http.ServeMux, mh *MangaHandler, ch *ChapterHandler) {
	mux.HandleFunc("/manga", middleware.ErrorHandler(mh.Logger, func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return mh.List(w, r)
		case http.MethodPost:
			return mh.Create(w, r)
		default:
			return apperror.NewBadRequestError("Метод не поддерживается", nil)
		}
	}))

	mux.HandleFunc("/manga/", middleware.ErrorHandler(mh.Logger, func(w http.ResponseWriter, r *http.Request) error {
		if strings.HasSuffix(r.URL.Path, "/chapters") {
			return ch.ListByManga(w, r)
		}
		return mh.Detail(w, r)
	}))
}
