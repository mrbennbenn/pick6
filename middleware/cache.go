package middleware

import "net/http"

// CacheControl adds Cache-Control headers to static files for better performance
// and enables Fly.io edge caching
func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cache all static files for 1 hour
		w.Header().Set("Cache-Control", "public, max-age=3600")
		next.ServeHTTP(w, r)
	})
}
