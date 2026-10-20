package ui

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	cookieName     = "sa_session"
	sessionTTL     = 24 * time.Hour
	pruneEvery     = time.Minute
	sessionIDBytes = 32
)

type sessionKey struct{}

// Sessions keeps the per-visitor upstream credentials in memory, keyed by an
// opaque cookie. webui does no authentication and no JavaScript runs, so the
// credential a visitor types on the Credentials page lives here, never in the
// browser's storage and never on disk.
type Sessions struct {
	mu       sync.Mutex
	byID     map[string]*session
	lastPrun time.Time
}

type session struct {
	creds map[string]string // backend id -> credential
	seen  time.Time
}

// NewSessions returns an empty store.
func NewSessions() *Sessions {
	return &Sessions{byID: map[string]*session{}}
}

// handle is what the middleware puts on the request context.
type handle struct {
	s  *Sessions
	id string
}

// Middleware gives every request a session, creating the cookie on first
// contact, and exposes it to the page functions through the request context.
func (s *Sessions) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ""
		if c, err := r.Cookie(cookieName); err == nil {
			id = c.Value
		}
		id, fresh := s.touch(id)
		if fresh {
			//nolint:gosec // G124: Secure follows the request's scheme, so plain-http development works.
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    id,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
			})
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, handle{s: s, id: id})))
	})
}

// touch returns the id of the live session, making a new one when id is
// unknown or expired. fresh reports whether the id is new.
func (s *Sessions) touch(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.Sub(s.lastPrun) > pruneEvery {
		for k, v := range s.byID {
			if now.Sub(v.seen) > sessionTTL {
				delete(s.byID, k)
			}
		}
		s.lastPrun = now
	}
	if sess, ok := s.byID[id]; ok && now.Sub(sess.seen) <= sessionTTL {
		sess.seen = now
		return id, false
	}
	raw := make([]byte, sessionIDBytes)
	_, _ = rand.Read(raw)
	id = hex.EncodeToString(raw)
	s.byID[id] = &session{creds: map[string]string{}, seen: now}
	return id, true
}

func handleOf(ctx context.Context) (handle, bool) {
	h, ok := ctx.Value(sessionKey{}).(handle)
	return h, ok
}

// Credential returns the visitor's credential for a backend, or "".
func (s *Sessions) Credential(ctx context.Context, backendID string) string {
	h, ok := handleOf(ctx)
	if !ok {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byID[h.id]; ok {
		return sess.creds[backendID]
	}
	return ""
}

// SetCredential stores (or, when value is empty, forgets) a credential.
func (s *Sessions) SetCredential(ctx context.Context, backendID, value string) {
	h, ok := handleOf(ctx)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byID[h.id]; ok {
		if value == "" {
			delete(sess.creds, backendID)
		} else {
			sess.creds[backendID] = value
		}
	}
}

// ClearCredentials forgets every credential of the visitor.
func (s *Sessions) ClearCredentials(ctx context.Context) {
	h, ok := handleOf(ctx)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byID[h.id]; ok {
		sess.creds = map[string]string{}
	}
}
