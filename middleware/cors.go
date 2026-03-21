package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowedOrigins ...string) func(http.Handler) http.Handler {
	allowOrigin := "*"
	if len(allowedOrigins) > 0 {
		allowOrigin = strings.Join(allowedOrigins, ", ")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
