// Package routes handles HTTP routing and middleware configuration for the GSM emulator.
package routes

import (
	"net/http"
	"os"
	"strings"

	"github.com/charlesgreen/gsm/internal/api/handlers"
	"github.com/charlesgreen/gsm/internal/api/middleware"
	"github.com/charlesgreen/gsm/internal/storage"
)

// SetupRoutes configures and returns an HTTP router with all API endpoints and middleware.
func SetupRoutes(storage storage.Storage) *http.ServeMux {
	mux := http.NewServeMux()

	secretsHandler := handlers.NewSecretsHandler(storage)
	versionsHandler := handlers.NewVersionsHandler(storage)
	healthHandler := handlers.NewHealthHandler()

	enableAuth := os.Getenv("GSM_ENABLE_AUTH") == "true"
	enableCORS := os.Getenv("GSM_ENABLE_CORS") != "false"
	
	var authMiddleware func(http.Handler) http.Handler
	if enableAuth {
		authMiddleware = middleware.MockAuth
	} else {
		authMiddleware = middleware.NoAuth
	}

	applyMiddleware := func(handler http.Handler) http.Handler {
		handler = middleware.Logging(handler)
		if enableCORS {
			handler = middleware.CORS(handler)
		}
		return handler
	}

	applyAuthMiddleware := func(handler http.Handler) http.Handler {
		return applyMiddleware(authMiddleware(handler))
	}

	mux.Handle("/health", applyMiddleware(http.HandlerFunc(healthHandler.Health)))
	mux.Handle("/ready", applyMiddleware(http.HandlerFunc(healthHandler.Ready)))

	mux.Handle("/v1/projects/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && matchesPattern(r.URL.Path, "/v1/projects/*/secrets"):
			applyAuthMiddleware(http.HandlerFunc(secretsHandler.CreateSecret)).ServeHTTP(w, r)
		
		case r.Method == http.MethodGet && matchesPattern(r.URL.Path, "/v1/projects/*/secrets"):
			applyAuthMiddleware(http.HandlerFunc(secretsHandler.ListSecrets)).ServeHTTP(w, r)
		
		case r.Method == http.MethodGet && matchesPattern(r.URL.Path, "/v1/projects/*/secrets/*") && !containsVersions(r.URL.Path):
			applyAuthMiddleware(http.HandlerFunc(secretsHandler.GetSecret)).ServeHTTP(w, r)
		
		case r.Method == http.MethodDelete && matchesPattern(r.URL.Path, "/v1/projects/*/secrets/*") && !containsVersions(r.URL.Path):
			applyAuthMiddleware(http.HandlerFunc(secretsHandler.DeleteSecret)).ServeHTTP(w, r)
		
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, ":addVersion") && matchesPattern(strings.TrimSuffix(r.URL.Path, ":addVersion"), "/v1/projects/*/secrets/*"):
			applyAuthMiddleware(http.HandlerFunc(versionsHandler.AddSecretVersion)).ServeHTTP(w, r)
		
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, ":access") && matchesPattern(strings.TrimSuffix(r.URL.Path, ":access"), "/v1/projects/*/secrets/*/versions/*"):
			applyAuthMiddleware(http.HandlerFunc(versionsHandler.AccessSecretVersion)).ServeHTTP(w, r)
		
		case r.Method == http.MethodGet && matchesPattern(r.URL.Path, "/v1/projects/*/secrets/*/versions"):
			applyAuthMiddleware(http.HandlerFunc(versionsHandler.ListSecretVersions)).ServeHTTP(w, r)
		
		case r.Method == http.MethodDelete && matchesPattern(r.URL.Path, "/v1/projects/*/secrets/*/versions/*") && !containsAccess(r.URL.Path):
			applyAuthMiddleware(http.HandlerFunc(versionsHandler.DeleteSecretVersion)).ServeHTTP(w, r)
		
		default:
			applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error": {"code": 404, "message": "Not found", "status": "NOT_FOUND"}}`))
			})).ServeHTTP(w, r)
		}
	}))

	return mux
}

func matchesPattern(path, pattern string) bool {
	return pathMatches(path, pattern)
}

func pathMatches(path, pattern string) bool {
	pathParts := splitPath(path)
	patternParts := splitPath(pattern)
	
	if len(pathParts) != len(patternParts) {
		return false
	}
	
	for i, patternPart := range patternParts {
		if patternPart == "*" {
			continue
		}
		if pathParts[i] != patternPart {
			return false
		}
	}
	
	return true
}

func splitPath(path string) []string {
	parts := []string{}
	current := ""
	
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

func containsVersions(path string) bool {
	return contains(path, "/versions")
}

func containsAccess(path string) bool {
	return contains(path, ":access")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}