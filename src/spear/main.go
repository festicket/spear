package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/go-openapi/loads"
	"github.com/google/go-github/github"
	"github.com/gorilla/pat"
)

func main() {
	r := pat.New()
	r.Get("/{branch}/files/{filename}/json/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		branchName := r.URL.Query().Get(":branch")
		filename := path.Join(DIR, r.URL.Query().Get(":filename"))

		ctx := context.Background()
		opts := github.RepositoryContentGetOptions{Ref: branchName}
		fileContent, _, _, err := GetGithubClient().Repositories.GetContents(
			ctx, OWNER, REPO, filename, &opts,
		)
		if HTTPError(rw, err) {
			return
		}

		specDoc, err := loads.Spec(*fileContent.DownloadURL)
		if HTTPError(rw, err) {
			return
		}

		b, err := json.MarshalIndent(specDoc.Spec(), "", "  ")
		if HTTPError(rw, err) {
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}))
	r.Get("/{branch}/doc/{filename}", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get(":filename")
		branchName := r.URL.Query().Get(":branch")
		context := struct {
			SpecURL string
			Title   string
		}{
			// TODO: use .Name() and then GetRoute
			// http://www.gorillatoolkit.org/pkg/pat
			SpecURL: fmt.Sprintf("/%s/files/%s/json/", branchName, filename),
			Title:   fmt.Sprintf("%s :: %s", branchName, filename),
		}

		RenderTemplate(rw, "redoc.html", &context)
	}))
	r.Get("/{branch}/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		branchName := r.URL.Query().Get(":branch")

		ctx := context.Background()
		opts := github.RepositoryContentGetOptions{Ref: branchName}
		_, files, _, err := GetGithubClient().Repositories.GetContents(
			ctx, OWNER, REPO, DIR, &opts,
		)
		if HTTPError(rw, err) {
			return
		}

		templateContext := struct {
			BranchName string
			Files      []*Link
		}{
			BranchName: branchName,
			Files:      []*Link{},
		}

		for _, f := range files {
			if *f.Name != "README.md" {
				templateContext.Files = append(templateContext.Files, &Link{
					Name: *f.Name,
					URL:  fmt.Sprintf("/%s/doc/%s", branchName, *f.Name),
				})
			}
		}

		RenderTemplate(rw, "index.html", &templateContext)
	}))
	r.Get("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, "/master/", 302)
	}))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
