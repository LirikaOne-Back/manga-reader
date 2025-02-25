package handlers

import (
	"net/http"
	"strings"
)

func RegisterPageRoutes(mux *http.ServeMux, ph *PageHandler) {
	mux.HandleFunc("/page/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			ph.UploadImage(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/pages/chapter/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && strings.HasSuffix(r.URL.Path, "/pages") {
			ph.ListByChapter(w, r)
		}
	})

	mux.HandleFunc("/page/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			ph.Delete(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/page/image/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			ph.ServeImage(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
}
