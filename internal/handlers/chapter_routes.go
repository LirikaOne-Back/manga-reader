package handlers

import "net/http"

func RegisterChapterRoutes(mux *http.ServeMux, ch *ChapterHandler) {
	mux.HandleFunc("/chapter", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ch.Create(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/chapter/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ch.GetById(w, r)
		case http.MethodPut:
			ch.Update(w, r)
		case http.MethodDelete:
			ch.Delete(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
}
