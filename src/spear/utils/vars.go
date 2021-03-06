package utils

import (
	"os"
	"strings"
)

var GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")

// Expected format: "owner;repo;path"
var SPECS_DIR = os.Getenv("SPECS_DIR")

var chunks = strings.Split(SPECS_DIR, ";")

var OWNER, REPO, DIR = chunks[0], chunks[1], chunks[2]

// Link represents data required to render <a> tag
type Link struct {
	URL  string
	Name string
}

var HOST = os.Getenv("HOST")

var SCHEME = os.Getenv("SCHEME")

// For basic auth
var PASSWORD = os.Getenv("PASSWORD")
var USERNAME = os.Getenv("USERNAME")
