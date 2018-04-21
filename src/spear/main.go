package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/google/go-github/github"
	"github.com/gorilla/pat"
	"golang.org/x/oauth2"
)

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

var SPECS_DIR = os.Getenv("SPECS_DIR")

var GithubClient *github.Client

const (
	redocLatest   = "https://rebilly.github.io/ReDoc/releases/latest/redoc.min.js"
	redocTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Title }}</title>
    <!-- needed for adaptive design -->
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!--
    ReDoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='{{ .SpecURL }}'></redoc>
    <script src="{{ .RedocURL }}"> </script>
  </body>
</html>
`
)

// RedocOpts configures the Redoc middlewares
type RedocOpts struct {
	// BasePath for the UI path, defaults to: /
	BasePath string
	// Path combines with BasePath for the full UI path, defaults to: docs
	Path string
	// SpecURL the url to find the spec for
	SpecURL string
	// RedocURL for the js that generates the redoc site, defaults to: https://rebilly.github.io/ReDoc/releases/latest/redoc.min.js
	RedocURL string
	// Title for the documentation site, default to: API documentation
	Title string
}

// EnsureDefaults in case some options are missing
func (r *RedocOpts) EnsureDefaults() {
	if r.BasePath == "" {
		r.BasePath = "/"
	}
	if r.Path == "" {
		r.Path = "docs"
	}
	if r.SpecURL == "" {
		r.SpecURL = "/swagger.json"
	}
	if r.RedocURL == "" {
		r.RedocURL = redocLatest
	}
	if r.Title == "" {
		r.Title = "API documentation"
	}
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

func main() {
	r := pat.New()
	r.Get("/files/{filename}", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get(":filename")

		parts := strings.Split(SPECS_DIR, ";")
		owner, repo, p := parts[0], parts[1], parts[2]
		p = path.Join(p, filename)

		ctx := context.Background()
		fileContent, _, _, err := GetGithubClient().Repositories.GetContents(ctx, owner, repo, p, nil)
		if err != nil {
			log.Fatal(err)
		}

		specDoc, err := loads.Spec(*fileContent.DownloadURL)
		if err != nil {
			log.Fatal(err)
		}

		b, err := json.MarshalIndent(specDoc.Spec(), "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}))
	r.Get("/doc/{filename}", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get(":filename")

		opts := RedocOpts{}
		opts.EnsureDefaults()
		opts.SpecURL = fmt.Sprintf("/files/%s", filename)

		tmpl := template.Must(template.New("redoc").Parse(redocTemplate))
		buf := bytes.NewBuffer(nil)
		_ = tmpl.Execute(buf, opts)
		b := buf.Bytes()

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}))
	r.Get("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		parts := strings.Split(SPECS_DIR, ";")
		owner, repo, p := parts[0], parts[1], parts[2]

		ctx := context.Background()
		_, files, _, err := GetGithubClient().Repositories.GetContents(ctx, owner, repo, p, nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			if *f.Name != "README.md" {
				io.WriteString(rw, fmt.Sprintf(`<p><a href="/doc/%s" target="_blank" rel="nofollow">%s</a></p>`, *f.Name, *f.Name))
			}
		}
	}))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
