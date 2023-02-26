package openapi

import (
	"io/ioutil"
	"net/http"
)

// OpenAPIHandler returns an HTTP handler that serves the OpenAPI specification for the service.
func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {

	swaggerBytes, err := ioutil.ReadFile("docs/swagger.json")
	if err != nil {
		http.Error(w, "failed to read OpenAPI specification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(swaggerBytes)
}
