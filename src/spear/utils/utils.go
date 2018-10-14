package utils

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"reflect"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
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

func LoadFile(branchName string, name string) (string, error) {
	filename := path.Join(DIR, name)

	ctx := context.Background()
	opts := github.RepositoryContentGetOptions{Ref: branchName}
	fileContent, _, _, err := GetGithubClient().Repositories.GetContents(
		ctx, OWNER, REPO, filename, &opts,
	)

	if err != nil {
		return "", fmt.Errorf("Can't load the file due to the error: %s", err)
	}

	content, err := fileContent.GetContent()

	return content, err
}

// Loads a spec from the branch given.
func LoadSpec(branchName string, name string) (*loads.Document, error) {
	filename := path.Join(DIR, name)

	ctx := context.Background()
	opts := github.RepositoryContentGetOptions{Ref: branchName}
	fileContent, _, _, err := GetGithubClient().Repositories.GetContents(
		ctx, OWNER, REPO, filename, &opts,
	)

	if err != nil {
		return nil, fmt.Errorf("Can't load the file due to the error: %s", err)
	}

	specDoc, err := loads.Spec(*fileContent.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("Can't parse the spec due to the error: %s", err)
	}

	return specDoc, nil
}

// Populates the target map given by examples of the response.
func BuildExample(schema *spec.Schema, key string, target *map[string]interface{}) {
	for k, v := range schema.Properties {
		if v.Type.Contains("object") {
			(*target)[k] = make(map[string]interface{})
			BuildExample(&v, k, target)
		} else {
			example := v.SwaggerSchemaProps.Example

			if key != "" {
				(*target)[key].(map[string]interface{})[k] = example
			} else {
				(*target)[k] = example
			}
		}
	}
}

// Returns an operation for HTTP method specified
func GetOperation(method string, pathItem *spec.PathItem) interface{} {
	t := reflect.ValueOf(pathItem).Elem()
	field := t.FieldByName(method)

	if field.IsValid() {
		return field.Interface()
	} else {
		return nil
	}
}
