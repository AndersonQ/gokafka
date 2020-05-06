package handlers

import (
	"fmt"
	"net/http"
)

// Constants for http headers
const (
	ContentType     = "Content-Type"                    // ContentType header key
	ContentTypeJSON = "application/json; charset=utf-8" // ContentTypeJSON content type for json requests

	jsonResp = `{"status":"%s","version":"%s","build_time":"%s"}`
)

var version = "development"
var buildTime = "build tome not set"

// NewLivenessHandler a handler for kubernetes liveness probe
func NewLivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, ContentTypeJSON)
		resp := fmt.Sprintf(
			jsonResp,
			"Kubernetes I'm ok', no need to restart me",
			version,
			buildTime)

		_, _ = w.Write([]byte(resp))
	}
}

// NewReadinessHandler a handler for kubernetes readness probe
func NewReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, ContentTypeJSON)
		resp := fmt.Sprintf(
			jsonResp,
			"Kubernetes I'm ok', you can send requests to me",
			version,
			buildTime)

		_, _ = w.Write([]byte(resp))
	}
}
