package middleware

import (
	"log"
	"net/http"
)

type APIKey struct {
	APIKeys []string
	Log     *log.Logger
}

func (a *APIKey) ServeHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from X-API-Key header
		apiKey := r.Header.Get("X-API-Key")

		// Check if API key exists
		if apiKey == "" {
			if a.Log != nil {
				a.Log.Printf("API auth failed: missing API key - path=%s remote=%s", r.URL.Path, r.RemoteAddr)
			}
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}

		// Check against known list of keys
		validKey := false
		for _, key := range a.APIKeys {
			if key == apiKey {
				validKey = true
				break
			}
		}

		if !validKey {
			if a.Log != nil {
				a.Log.Printf("API auth failed: invalid API key - path=%s remote=%s", r.URL.Path, r.RemoteAddr)
			}
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		if a.Log != nil {
			a.Log.Printf("API auth success - path=%s remote=%s", r.URL.Path, r.RemoteAddr)
		}

		// API key is valid, proceed to next handler
		next.ServeHTTP(w, r)
	})
}
