package middlewares

import (
	"net/http"
	"spear/utils"
)

type handler func(w http.ResponseWriter, r *http.Request)

func BasicAuth(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		// Basic auth is disabled.
		if utils.USERNAME == "" || utils.PASSWORD == "" {
			h(w, r)
			return
		}

		if ok && username == utils.USERNAME && password == utils.PASSWORD {
			h(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=\"Login Required\"")
		http.Error(w, "authorization failed", http.StatusUnauthorized)
	}
}
