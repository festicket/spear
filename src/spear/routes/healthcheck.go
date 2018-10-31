package routes

import "net/http"

// HealthCheck responds 200 always
func HealthCheck(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("I am alive!"))
}
