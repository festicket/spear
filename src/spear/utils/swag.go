package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-openapi/swag"
)

// Copied from https://github.com/go-openapi/swag/blob/573d295/loading.go#L31
// But this implimentation uses own loadHTTPBytes implementation.
func LoadFromFileOrHTTP(path string) ([]byte, error) {
	return swag.LoadStrategy(path, ioutil.ReadFile, loadHTTPBytes(swag.LoadHTTPTimeout))(path)
}

// Copied from https://github.com/go-openapi/swag/blob/573d295/loading.go#L55
// Differences marked by (*)
func loadHTTPBytes(timeout time.Duration) func(path string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		client := &http.Client{Timeout: timeout}
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(USERNAME, PASSWORD) // (*)
		resp, err := client.Do(req)
		defer func() {
			if resp != nil {
				if e := resp.Body.Close(); e != nil {
					log.Println(e)
				}
			}
		}()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("could not access document at %q [%s] ", path, resp.Status)
		}

		return ioutil.ReadAll(resp.Body)
	}
}
