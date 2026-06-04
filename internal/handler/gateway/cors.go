package gateway

import (
	"net/http"
	"strings"
)

// CORSConfig holds the allowed origins for CORS.
type CORSConfig struct {
	// AllowedOrigins is a comma-separated list of allowed origins.
	// Example: "http://localhost:4200,https://admin.example.com"
	AllowedOrigins string
}

// CORSMiddleware returns an HTTP middleware that sets CORS headers for
// requests from allowed origins. It handles preflight OPTIONS requests.
// withCredentials requires explicit origin (not *) and Allow-Credentials: true.
func CORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, origin := range strings.Split(config.AllowedOrigins, ",") {
		allowed[strings.TrimSpace(origin)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !allowed[origin] {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
