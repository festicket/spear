package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
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
	r.Get("/{branch}/files/{filename}/json/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get(":filename")
		branchName := r.URL.Query().Get(":branch")

		parts := strings.Split(SPECS_DIR, ";")
		owner, repo, p := parts[0], parts[1], parts[2]
		p = path.Join(p, filename)

		ctx := context.Background()
		opts := github.RepositoryContentGetOptions{Ref: branchName}
		fileContent, _, _, err := GetGithubClient().Repositories.GetContents(ctx, owner, repo, p, &opts)
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
	r.Get("/{branch}/doc/{filename}", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get(":filename")
		branchName := r.URL.Query().Get(":branch")

		opts := RedocOpts{}
		opts.EnsureDefaults()
		// TODO: use .Name() and then GetRoute
		// http://www.gorillatoolkit.org/pkg/pat
		opts.SpecURL = fmt.Sprintf("/%s/files/%s/json/", branchName, filename)

		tmpl := template.Must(template.New("redoc").Parse(redocTemplate))
		buf := bytes.NewBuffer(nil)
		_ = tmpl.Execute(buf, opts)
		b := buf.Bytes()

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}))
	r.Get("/{branch}/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		branchName := r.URL.Query().Get(":branch")

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		parts := strings.Split(SPECS_DIR, ";")
		owner, repo, p := parts[0], parts[1], parts[2]

		ctx := context.Background()
		opts := github.RepositoryContentGetOptions{Ref: branchName}
		_, files, _, err := GetGithubClient().Repositories.GetContents(ctx, owner, repo, p, &opts)
		if err != nil {
			log.Fatal(err)
		}

		type File struct {
			URL  string
			Name string
		}

		templateContext := struct {
			BranchName string
			Files      []*File
		}{
			BranchName: branchName,
			Files:      []*File{},
		}

		for _, f := range files {
			if *f.Name != "README.md" {
				templateContext.Files = append(templateContext.Files, &File{
					Name: *f.Name,
					URL:  fmt.Sprintf("/%s/doc/%s", branchName, *f.Name),
				})
			}
		}

		t, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Fatal(err)
		}

		t.Execute(rw, &templateContext)
	}))
	r.Get("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, "/master/", 302)
	}))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
