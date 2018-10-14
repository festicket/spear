package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"spear/utils"

	"github.com/ghodss/yaml"
)

func SpecPage(rw http.ResponseWriter, r *http.Request) {
	fname := r.URL.Query().Get(":defs") // TODO: sanitaize the path!
	branch := r.URL.Query().Get(":branch")

	if fname != "" {
		content, err := utils.LoadFile(branch, fname)
		if utils.HTTPError(rw, err) {
			return
		}

		j, err := yaml.YAMLToJSON([]byte(content))
		if utils.HTTPError(rw, err) {
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(j)

		return
	}

	fname = r.URL.Query().Get(":filename")

	specDoc, err := utils.LoadSpec(branch, fname)
	if utils.HTTPError(rw, err) {
		return
	}

	// Patch properties to have proper "try it now" links
	specDoc.Spec().SwaggerProps.Host = utils.HOST
	specDoc.Spec().SwaggerProps.BasePath = fmt.Sprintf("/%s/try/%s", r.URL.Query().Get(":branch"), r.URL.Query().Get(":filename"))

	b, err := json.MarshalIndent(specDoc.Spec(), "", "  ")
	if utils.HTTPError(rw, err) {
		return
	}

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b)
}
