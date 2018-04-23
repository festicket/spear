package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var GithubClient *github.Client

// RenderTemplate renders template given as a filename with the context provided.
func RenderTemplate(rw http.ResponseWriter, name string, context interface{}) {
	p := path.Join("templates", name)

	t, err := template.ParseFiles(p)
	Fatal(err)

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	t.Execute(rw, context)
}

// Fatal handles errors.
func Fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// A shortcut to return HTTP error.
func HTTPError(rw http.ResponseWriter, err error) bool {
	if err != nil {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusInternalServerError)
		io.WriteString(rw, fmt.Sprintf("%s", err))
		return true
	}
	return false
}

func GetGithubClient() *github.Client {
	if GithubClient != nil {
		return GithubClient
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)
	GithubClient = github.NewClient(tc)

	return GithubClient
}
