# SPEAR

A service to provide an access to API specifications from your Github repository
and render them via Swagger.

# Quickstart

You can use the pre build image. This is an example of `docker-compose` file:

```
version: '3'
services:
  spear:
    image: zerc/spear
    environment:
     - GITHUB_TOKEN=123
     - SPECS_DIR=owner;repo;path
    ports:
     - "8000:8000"
```

Where:

* `GITHUB_TOKEN` - your github token. Generate it [here](https://github.com/settings/tokens)
* `SPECS_DIR` - a string with information about the target repo. `owner` - repo's owner, `repo` - repo's name, `path` - a path to the folder with API specs.

Save this as `docker-compose.yml`.

Now run:

```
docker-compose up spear
```

Visit http://localhost:8000

## TODO

- [x] Build a docker image
- [x] Branch selector
- [ ] Build responses from examples whenever it is possible
  - [x] Get
  - [ ] Post
  - [ ] Not only 200 responses
- [ ] Cache requests (to avoid rate limits from Github)
  - [ ] Ability to reset the cache manually
- [ ] Clear the code
