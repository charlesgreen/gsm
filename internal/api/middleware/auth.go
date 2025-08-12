// Package middleware provides HTTP middleware functions for authentication, CORS, and logging.
package middleware

import (
	"net/http"
	"strings"
)

// MockAuth is a middleware that validates Bearer token authentication headers.
func MockAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": {"code": 401, "message": "Request is missing required authentication credential", "status": "UNAUTHENTICATED"}}`))
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": {"code": 401, "message": "Invalid authentication credentials", "status": "UNAUTHENTICATED"}}`))
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": {"code": 401, "message": "Invalid authentication token", "status": "UNAUTHENTICATED"}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// NoAuth is a middleware that bypasses authentication and passes all requests through.
func NoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}