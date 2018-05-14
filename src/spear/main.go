package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/google/go-github/github"
	"github.com/gorilla/pat"
)

func main() {
	r := pat.New()
	r.NewRoute().PathPrefix("/{branch}/try/{filename}/{path:.*}").Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		branch := r.URL.Query().Get(":branch")
		fname := r.URL.Query().Get(":filename")
		expectedStatusCode := r.Header.Get("X-Status")
		specDoc, err := LoadSpec(branch, fname)

		if HTTPError(rw, err) {
			return
		}

		// TODO: for some reason r.URL.Scheme and r.URL.Host are empty
		base := fmt.Sprintf("%s://%s/%s/files/%s/json/", SCHEME, HOST, branch, fname)

		spec.ExpandSpec(specDoc.Spec(), &spec.ExpandOptions{
			RelativeBase: base,
			SkipSchemas:  false,
		})

		re := regexp.MustCompile(`\{[\d\w\-]+\}`) // regexp to replace things like {someVariable}

		for path, pathItem := range specDoc.Spec().SwaggerProps.Paths.Paths {
			pattern := re.ReplaceAllString(path, `[\d\w-]+`)
			matched, _ := regexp.MatchString(fmt.Sprintf("%s$", pattern), fmt.Sprintf("/%s", r.URL.Query().Get(":path")))

			methodName := strings.Title(strings.ToLower(strings.Replace(r.Method, "Method", "", -1)))
			operation := GetOperation(methodName, &pathItem).(*spec.Operation)

			fmt.Println("=====")
			fmt.Println(methodName)
			fmt.Println(matched)
			fmt.Println(pattern)
			fmt.Println(operation)
			fmt.Println(expectedStatusCode)
			fmt.Println("=====")

			if matched == false || operation == nil {
				continue
			}

			var schema *spec.Schema
			responseStatus := http.StatusOK

			if expectedStatusCode != "" {
				status, err := strconv.ParseInt(expectedStatusCode, 10, 64)

				if HTTPError(rw, err) {
					return
				}

				responseStatus = int(status)
				schema = operation.OperationProps.Responses.ResponsesProps.StatusCodeResponses[responseStatus].Schema
			} else {
				for _, successStatusCode := range []int{200, 302, 301} {
					schema = operation.OperationProps.Responses.ResponsesProps.StatusCodeResponses[successStatusCode].Schema
					if schema != nil {
						responseStatus = successStatusCode
						break
					}
				}
			}

			if schema == nil {
				rw.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			resp := make(map[string]interface{})
			BuildExample(schema, "", &resp)
			b, _ := json.MarshalIndent(resp, "", " ")
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.WriteHeader(responseStatus)
			rw.Write(b)
			return
		}

		rw.WriteHeader(http.StatusNotFound)
		return
	}))
	r.Get("/{branch}/files/{filename}/json/{defs:.*}", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fname := r.URL.Query().Get(":defs") // TODO: sanitaize the path!
		branch := r.URL.Query().Get(":branch")

		if fname != "" {
			content, err := LoadFile(branch, fname)
			if HTTPError(rw, err) {
				return
			}

			j, err := yaml.YAMLToJSON([]byte(content))
			if HTTPError(rw, err) {
				return
			}

			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusOK)
			rw.Write(j)

			return
		}

		fname = r.URL.Query().Get(":filename")

		specDoc, err := LoadSpec(branch, fname)
		if HTTPError(rw, err) {
			return
		}

		// Patch properties to have proper "try it now" links
		specDoc.Spec().SwaggerProps.Host = HOST
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
