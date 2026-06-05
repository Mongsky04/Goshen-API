package middleware

import (
	"net/http"
	"slices"
	"strings"
)

func CORS(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestOrigin := r.Header.Get("Origin")
			if slices.Contains(origins, requestOrigin) {
				w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			} else if len(origins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", origins[0])
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ParseOrigins splits a comma-separated origins string into a slice.
func ParseOrigins(raw string) []string {
	var out []string
	for _, s := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}
