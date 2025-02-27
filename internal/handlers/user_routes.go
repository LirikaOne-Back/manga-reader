package handlers

import "net/http"

func RegisterUserRoutes(mux *http.ServeMux, uh *UserHandler) {
	mux.HandleFunc("/user/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		uh.Register(w, r)
	})
	mux.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		uh.Login(w, r)
	})
}
