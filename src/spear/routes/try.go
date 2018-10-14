package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"spear/utils"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
)

func TryPage(rw http.ResponseWriter, r *http.Request) {
	branch := r.URL.Query().Get(":branch")
	fname := r.URL.Query().Get(":filename")
	expectedStatusCode := r.Header.Get("X-Status")
	specDoc, err := utils.LoadSpec(branch, fname)

	if utils.HTTPError(rw, err) {
		return
	}

	// TODO: for some reason r.URL.Scheme and r.URL.Host are empty
	base := fmt.Sprintf("%s://%s/%s/files/%s/json/", utils.SCHEME, utils.HOST, branch, fname)

	spec.ExpandSpec(specDoc.Spec(), &spec.ExpandOptions{
		RelativeBase: base,
		SkipSchemas:  false,
	})

	re := regexp.MustCompile(`\{[\d\w\-]+\}`) // regexp to replace things like {someVariable}

	for path, pathItem := range specDoc.Spec().SwaggerProps.Paths.Paths {
		pattern := re.ReplaceAllString(path, `[\d\w-]+`)
		matched, _ := regexp.MatchString(fmt.Sprintf("%s$", pattern), fmt.Sprintf("/%s", r.URL.Query().Get(":path")))

		methodName := strings.Title(strings.ToLower(strings.Replace(r.Method, "Method", "", -1)))
		operation := utils.GetOperation(methodName, &pathItem).(*spec.Operation)

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

			if utils.HTTPError(rw, err) {
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
		utils.BuildExample(schema, "", &resp)
		b, _ := json.MarshalIndent(resp, "", " ")
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(responseStatus)
		rw.Write(b)
		return
	}

	rw.WriteHeader(http.StatusNotFound)
	return
}
