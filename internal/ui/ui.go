// Package ui builds the admin UI as a webui declaration over a
// resource.Registry: one list, detail, create, edit and delete page per
// resource, plus a credentials page. webui renders and serves it; nothing here
// writes HTML, CSS or JavaScript.
package ui

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Olian04/webui/pkg/webui"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Olian04/simple-admin/internal/backend"
	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// Prefix is where the UI is mounted.
const Prefix = "/ui"

// Options configure a build.
type Options struct {
	Sessions *Sessions
	PageSize int
	Logger   *slog.Logger
	// Metrics is where webui registers its metrics. Rebuilds count into the same ones.
	Metrics prometheus.Registerer
	// Secret keys webui's per-process addresses; it must not change between
	// builds. NewLive sets a random one when it is empty.
	Secret []byte
}

type builder struct {
	reg      *resource.Registry
	sessions *Sessions
	pageSize int
	log      *slog.Logger
	cache    *optionCache
}

// Build compiles the UI for the registry's current resources.
func Build(reg *resource.Registry, o Options) (http.Handler, error) {
	if o.Sessions == nil {
		o.Sessions = NewSessions()
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.PageSize < 1 {
		o.PageSize = 25
	}
	b := &builder{reg: reg, sessions: o.Sessions, pageSize: o.PageSize, log: o.Logger, cache: newOptionCache()}

	app := webui.App{
		Brand:  webui.Brand{Name: "Simple Admin"},
		Pages:  b.pages(),
		Logger: o.Logger,
		Secret: o.Secret,
	}
	app.Metrics = o.Metrics
	h, err := app.Compile(Prefix)
	if err != nil {
		return nil, fmt.Errorf("compile ui: %w", err)
	}
	return o.Sessions.Middleware(h), nil
}

// Live serves the most recently built UI and swaps it when the registry
// changes, so discovery can add and remove resources without a restart.
type Live struct {
	reg  *resource.Registry
	opts Options
	cur  atomic.Pointer[http.Handler]
	mu   sync.Mutex
}

// NewLive builds the first UI. It fails when that build does not compile.
func NewLive(reg *resource.Registry, o Options) (*Live, error) {
	if len(o.Secret) == 0 {
		o.Secret = make([]byte, 32)
		_, _ = rand.Read(o.Secret)
	}
	if o.Sessions == nil {
		o.Sessions = NewSessions()
	}
	l := &Live{reg: reg, opts: o}
	if err := l.Rebuild(); err != nil {
		return nil, err
	}
	return l, nil
}

// Rebuild recompiles the UI from the registry. On failure the previous UI
// keeps serving and the error is returned.
func (l *Live) Rebuild() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	h, err := Build(l.reg, l.opts)
	if err != nil {
		return err
	}
	l.cur.Store(&h)
	return nil
}

func (l *Live) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := l.cur.Load(); h != nil {
		(*h).ServeHTTP(w, r)
		return
	}
	http.Error(w, "ui not ready", http.StatusServiceUnavailable)
}

// scope puts the visitor's credential for a backend on the context, where the
// upstream client reads it.
func (b *builder) scope(ctx context.Context, backendID string) context.Context {
	if cred := b.sessions.Credential(ctx, backendID); cred != "" {
		return backend.WithCredential(ctx, cred)
	}
	return ctx
}

// readErr turns a failed read into what the visitor should be told: a missing
// record or a missing credential, with a way on. Any other error is logged by
// webui and shown as "Could not load".
func (b *builder) readErr(ctx context.Context, d resource.ResourceDescriptor, err error) error {
	switch {
	case errors.Is(err, backend.ErrNotFound):
		return webui.NotFound("There is no such " + strings.ToLower(singular(d.Label)) + ".")
	case errors.Is(err, backend.ErrUnauthorized):
		msg := d.Backend.Label + " needs a credential."
		if b.sessions.Credential(ctx, d.Backend.ID) != "" {
			msg = d.Backend.Label + " refused the credential."
		}
		r := webui.Unauthorized(msg + " Set one on the Credentials page.")
		return r.WithLink("Credentials", webui.Open(ctx, credentialsID, webui.NoArgs{}))
	default:
		return err
	}
}

type optionCache struct {
	mu sync.Mutex
	m  map[string]cachedOptions
}

type cachedOptions struct {
	opts    []webui.Option
	expires time.Time
}

func newOptionCache() *optionCache { return &optionCache{m: map[string]cachedOptions{}} }

func (c *optionCache) get(key string) ([]webui.Option, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok || time.Now().After(e.expires) {
		delete(c.m, key)
		return nil, false
	}
	return e.opts, true
}

func (c *optionCache) put(key string, opts []webui.Option, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.m) > 256 {
		now := time.Now()
		for k, e := range c.m {
			if now.After(e.expires) {
				delete(c.m, k)
			}
		}
	}
	c.m[key] = cachedOptions{opts: opts, expires: time.Now().Add(ttl)}
}
