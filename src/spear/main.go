package main

import (
	"log"
	"net/http"

	"spear/routes"

	"github.com/gorilla/pat"
)

func main() {
	r := pat.New()
	r.NewRoute().PathPrefix("/{branch}/try/{filename}/{path:.*}").Handler(http.HandlerFunc(routes.TryPage))
	r.Get("/{branch}/files/{filename}/json/{defs:.*}", http.HandlerFunc(routes.SpecPage))
	r.Get("/{branch}/doc/{filename}", http.HandlerFunc(routes.BranchDocument))
	r.Get("/{branch}/", http.HandlerFunc(routes.BranchPage))
	r.Get("/", http.HandlerFunc(routes.IndexPage))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
