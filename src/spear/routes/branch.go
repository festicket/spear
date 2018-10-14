package routes

import (
	"context"
	"fmt"
	"net/http"
	"spear/utils"

	"github.com/google/go-github/github"
)

func BranchPage(rw http.ResponseWriter, r *http.Request) {
	branchName := r.URL.Query().Get(":branch")

	ctx := context.Background()
	opts := github.RepositoryContentGetOptions{Ref: branchName}
	_, files, _, err := utils.GetGithubClient().Repositories.GetContents(
		ctx, utils.OWNER, utils.REPO, utils.DIR, &opts,
	)
	if utils.HTTPError(rw, err) {
		return
	}

	templateContext := struct {
		BranchName string
		Files      []*utils.Link
		Branches   []*utils.Link
	}{}
	templateContext.BranchName = branchName

	for _, f := range files {
		if *f.Name != "README.md" {
			templateContext.Files = append(templateContext.Files, &utils.Link{
				Name: *f.Name,
				URL:  fmt.Sprintf("/%s/doc/%s", branchName, *f.Name),
			})
		}
	}

	utils.RenderTemplate(rw, "branch.html", &templateContext)
}

func BranchDocument(rw http.ResponseWriter, r *http.Request) {
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

	utils.RenderTemplate(rw, "redoc.html", &context)
}
