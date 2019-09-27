# SPEAR

A service to render OpenAPI specification files stored on GitHub. It also generates API endpoints which return dummy data based on the specification.

# Quickstart

Change `$GOPATH` to the current directory using:

```bash
source activate
```

Install the dependencies required:

```bash
make install
```

We use [dep](https://github.com/golang/dep) for that.


To build the binary for Linux, run the command:


```bash
make build
```

After that you can just run the binary created:

```bash
./bin/spear
```

These are environment variables you need to set to make it work:

* `GITHUB_TOKEN` - your github token. Generate it [here](https://github.com/settings/tokens).
* `SPECS_DIR` - a string with information about the target repository in format `owner;repo;path`: `owner` - owner of the repository, `repo` - name of the repository, `path` - a path within the repository to the root folder with spec files.
* `USERNAME` and `PASSWORD` to enable [Basic Auth](https://en.wikipedia.org/wiki/Basic_access_authentication).

## Docker

You can use Docker for local development as well.

First, you need to create the compose file, use the example provided:

```bash
 cp docker-compose.yml.example docker-compose.yml
```
 
Don't forget to update values for environment variables in `docker-compose.yml`. 

You can automatically build the binary and image then run it via next command:


```bash
make bup
```

Visit http://localhost:8000

## TODO

- [x] Build a docker image
- [x] Branch selector
- [x] Support definitions in separate files
- [x] Build responses from examples whenever it is possible
  - [x] Get
  - [x] Post
  - [x] Not only 200 responses but provide a way to request an expected response code
- [ ] Cache requests (to avoid rate limits from Github)
  - [ ] Ability to reset the cache manually
- [ ] Clear the code
