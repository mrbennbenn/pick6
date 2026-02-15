package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrbennbenn/pick6/database"
	cache "github.com/patrickmn/go-cache"
	"github.com/segmentio/ksuid"
)

type Session struct {
	SecureCookie bool
	Log          *log.Logger
	Queries      *database.Queries
	Cache        *cache.Cache // In-memory cache for session validation
}

func (s *Session) ServeHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("vote_session")
		if err != nil {
			// No cookie found - create new session
			sessionID, err := s.createSession(w, r)
			if err != nil {
				if s.Log != nil {
					s.Log.Printf("Session creation failed: %v - path=%s remote=%s", err, r.URL.Path, r.RemoteAddr)
				}
				http.Error(w, "could not generate user session", http.StatusInternalServerError)
				return
			}

			if s.Log != nil {
				s.Log.Printf("New session created: %s - path=%s remote=%s", sessionID, r.URL.Path, r.RemoteAddr)
			}

			next.ServeHTTP(w, r.WithContext(ctxWithSession(r.Context(), sessionID)))
			return
		}

		// Cookie exists - validate it's a valid session
		if err := cookie.Valid(); err != nil {
			if s.Log != nil {
				s.Log.Printf("Session auth failed: invalid cookie - path=%s remote=%s", r.URL.Path, r.RemoteAddr)
			}
			http.Error(w, "invalid vote_session cookie", http.StatusUnauthorized)
			return
		}

		// Validate session exists - check cache first to reduce DB load
		sessionID := cookie.Value

		// Check cache first
		if s.Cache != nil {
			if _, found := s.Cache.Get(sessionID); found {
				// Session found in cache, skip DB query
				if s.Log != nil {
					s.Log.Printf("Session auth success (cached): %s - path=%s remote=%s", sessionID, r.URL.Path, r.RemoteAddr)
				}
				next.ServeHTTP(w, r.WithContext(ctxWithSession(r.Context(), sessionID)))
				return
			}
		}

		// Not in cache, validate against database (with retry for transient failures)
		err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
			_, queryErr := s.Queries.GetSession(r.Context(), sessionID)
			return queryErr
		})
		if err != nil {
			if s.Log != nil {
				s.Log.Printf("Session auth failed: session not found in database - sessionID=%s path=%s remote=%s error=%v", sessionID, r.URL.Path, r.RemoteAddr, err)
			}
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		// Cache the valid session for 5 minutes
		if s.Cache != nil {
			s.Cache.Set(sessionID, true, 5*time.Minute)
		}

		if s.Log != nil {
			s.Log.Printf("Session auth success: %s - path=%s remote=%s", sessionID, r.URL.Path, r.RemoteAddr)
		}

		next.ServeHTTP(w, r.WithContext(ctxWithSession(r.Context(), sessionID)))
	})
}

func (s *Session) createSession(w http.ResponseWriter, r *http.Request) (string, error) {
	sessionID := fmt.Sprintf("voter_%s", ksuid.New().String())

	// Insert session into database with NULL fields (will be updated later) with retry
	err := database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		_, queryErr := s.Queries.UpsertSession(r.Context(), database.UpsertSessionParams{
			SessionID: sessionID,
			Name:      sql.NullString{Valid: false},
			Email:     sql.NullString{Valid: false},
			Mobile:    sql.NullString{Valid: false},
		})
		return queryErr
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert session into database: %w", err)
	}

	// Set simple cookie with just the session ID
	http.SetCookie(w, &http.Cookie{
		Name:     "vote_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.SecureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours in seconds
	})

	// Cache the new session immediately
	if s.Cache != nil {
		s.Cache.Set(sessionID, true, 5*time.Minute)
	}

	return sessionID, nil
}

type sessionCtxKeyType string

var sessionCtxKey = "vote_session"

func ctxWithSession(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionCtxKey, sessionID)
}

func SessionFromCtx(ctx context.Context) (sessionID string, err error) {
	val := ctx.Value(sessionCtxKey)
	if val == nil {
		return "", errors.New("no session set on ctx")
	}

	return val.(string), nil
}
