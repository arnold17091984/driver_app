package handler

import (
	_ "embed"
	"net/http"
)

//go:embed swagger_ui.html
var swaggerHTML []byte

//go:embed openapi.yaml
var openapiSpec []byte

// SwaggerUI serves the Swagger UI page.
func SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(swaggerHTML)
}

// OpenAPISpec serves the OpenAPI YAML spec.
func OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(openapiSpec)
}
