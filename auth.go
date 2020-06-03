package main

import "net/http"

// APIKey Authentication middleware
func APIKeyAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if len(key) == 0 {
				writeErrorResponse(w, http.StatusUnauthorized, "'X-API-Key' header is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
