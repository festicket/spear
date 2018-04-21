package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

var SPECS_DIR = os.Getenv("SPECS_DIR")

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	parts := strings.Split(SPECS_DIR, ";")
	owner, repo, path := parts[0], parts[1], parts[2]

	_, files, _, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)

	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if *f.Name != "README.md" {
			log.Println(*f.Name)
		}
	}
}
