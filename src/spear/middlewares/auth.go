package middlewares

import (
	"net/http"
	"spear/utils"
)

type handler func(w http.ResponseWriter, r *http.Request)

func BasicAuth(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if ok && password == utils.PASSWORD && username == utils.USERNAME {
			h(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Login Required\"")
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		}
	}
}
