package handlers

import (
	"net/http"
	"strings"
)

func RegisterMangaRoutes(mux *http.ServeMux, mh *MangaHandler, ch *ChapterHandler) {
	mux.HandleFunc("/manga", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mh.List(w, r)
		case http.MethodPost:
			mh.Create(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/manga/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/chapters") {
			ch.ListByManga(w, r)
			return
		}
		mh.Detail(w, r)
	})
}
