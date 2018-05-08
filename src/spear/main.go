package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/go-openapi/spec"
	"github.com/google/go-github/github"
	"github.com/gorilla/pat"
)

func main() {
	r := pat.New()
	r.Get("/{branch}/try/{filename}/{path:.*}", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		specDoc, err := LoadSpec(r.URL.Query().Get(":branch"),
			r.URL.Query().Get(":filename"))
		if HTTPError(rw, err) {
			return
		}

		spec.ExpandSpec(specDoc.Spec(), &spec.ExpandOptions{
			SkipSchemas: false,
		})

		re := regexp.MustCompile(`\{[\d\w]+\}`) // regexp to replace things like {someVariable}

		for path, pathItem := range specDoc.Spec().SwaggerProps.Paths.Paths {
			pattern := re.ReplaceAllString(path, `[\d\w-]+`)
			matched, _ := regexp.MatchString(pattern, fmt.Sprintf("/%s", r.URL.Query().Get(":path")))

			if matched == false {
				continue
			}

			// TODO: Support other operations
			if pathItem.Get == nil {
				continue
			}

			resp := make(map[string]interface{})
			BuildExample(pathItem.Get.OperationProps.Responses.ResponsesProps.StatusCodeResponses[200].Schema, "", &resp)

			b, _ := json.MarshalIndent(resp, "", " ")
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusOK)
			rw.Write(b)
			return
		}

		rw.WriteHeader(http.StatusNotFound)
		return
	}))
	r.Get("/{branch}/files/{filename}/json/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		specDoc, err := LoadSpec(r.URL.Query().Get(":branch"),
			r.URL.Query().Get(":filename"))
		if HTTPError(rw, err) {
			return
		}

		// Patch properties to have proper "try it now" links
		specDoc.Spec().SwaggerProps.Host = HOSTNAME
		specDoc.Spec().SwaggerProps.BasePath = fmt.Sprintf("/%s/try/%s", r.URL.Query().Get(":branch"), r.URL.Query().Get(":filename"))

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
			Branches   []*Link
		}{}
		templateContext.BranchName = branchName

		for _, f := range files {
			if *f.Name != "README.md" {
				templateContext.Files = append(templateContext.Files, &Link{
					Name: *f.Name,
					URL:  fmt.Sprintf("/%s/doc/%s", branchName, *f.Name),
				})
			}
		}

		RenderTemplate(rw, "branch.html", &templateContext)
	}))
	r.Get("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		opts := github.ListOptions{PerPage: 100}
		branches, _, err := GetGithubClient().Repositories.ListBranches(
			ctx, OWNER, REPO, &opts,
		)
		if HTTPError(rw, err) {
			return
		}

		context := struct {
			Repo     string
			Branches []*Link
		}{}

		context.Repo = fmt.Sprintf("%s/%s", OWNER, REPO)

		for _, b := range branches {
			context.Branches = append(context.Branches, &Link{
				Name: *b.Name,
				URL:  fmt.Sprintf("/%s/", *b.Name),
			})
		}

		RenderTemplate(rw, "index.html", &context)
	}))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
