package routes

import (
	"context"
	"fmt"
	"net/http"
	"spear/utils"

	"github.com/google/go-github/github"
)

func IndexPage(rw http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	opts := github.ListOptions{PerPage: 100}
	branches, _, err := utils.GetGithubClient().Repositories.ListBranches(
		ctx, utils.OWNER, utils.REPO, &opts,
	)
	if utils.HTTPError(rw, err) {
		return
	}

	context := struct {
		Repo     string
		Branches []*utils.Link
	}{}

	context.Repo = fmt.Sprintf("%s/%s", utils.OWNER, utils.REPO)

	for _, b := range branches {
		context.Branches = append(context.Branches, &utils.Link{
			Name: *b.Name,
			URL:  fmt.Sprintf("/%s/", *b.Name),
		})
	}

	utils.RenderTemplate(rw, "index.html", &context)
}
