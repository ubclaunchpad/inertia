package util

import (
	"net/http"
)

// allowedRequest checks if given request uses one of the allowed methods.
// Always returns true if no methods are provided.
func allowedRequest(r *http.Request, methods ...string) bool {
	if len(methods) == 0 {
		return true
	}
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	return false
}

// WithMethods uses handler with only the declared methods. If no methods are
// provided, all methods are allowed.
func WithMethods(handler http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowedRequest(r, methods...) {
			http.Error(w, "request method not allowed", http.StatusBadRequest)
			return
		}
		handler(w, r)
	}
}
