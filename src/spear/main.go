package main

import (
	"log"
	"net/http"

	"spear/middlewares"
	"spear/routes"

	"github.com/gorilla/pat"
)

func main() {
	r := pat.New()

	r.NewRoute().PathPrefix("/{branch}/try/{filename}/{path:.*}").Handler(
		http.HandlerFunc(middlewares.BasicAuth(routes.TryPage)))

	// NOTE: this URL will be used to fetch additional files i.e. definitions
	// so it needs to be open for everyone :(
	// Need to change the approach i.e. add a service to work with Github as
	// filesystem instead of web thing.
	r.Get("/{branch}/files/{filename}/json/{defs:.*}", http.HandlerFunc(
		routes.SpecPage))

	r.Get("/{branch}/doc/{filename}", http.HandlerFunc(
		middlewares.BasicAuth(routes.BranchDocument)))

	r.Get("/{branch}/", http.HandlerFunc(
		middlewares.BasicAuth(routes.BranchPage)))

	r.Get("/", http.HandlerFunc(
		middlewares.BasicAuth(routes.IndexPage)))

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
