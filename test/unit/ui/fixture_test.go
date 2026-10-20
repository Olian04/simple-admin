package ui_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Olian04/simple-admin/internal/backend"
	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/schemavalidation"
	"github.com/Olian04/simple-admin/internal/ui"
)

const postSchema = `{
  "type": "object",
  "required": ["title"],
  "properties": {
    "id": { "type": "string", "readOnly": true },
    "title": { "type": "string", "maxLength": 200 },
    "body": { "type": "string" },
    "status": { "type": "string", "enum": ["draft", "published", "archived"] },
    "rating": { "type": "number" },
    "featured": { "type": "boolean" },
    "author": { "type": "string" },
    "tags": { "type": "array", "items": { "type": "string" } },
    "meta": { "type": "object" },
    "created": { "type": "string", "format": "date-time", "readOnly": true }
  }
}`

const authorSchema = `{
  "type": "object",
  "required": ["name"],
  "properties": {
    "id": { "type": "string", "readOnly": true },
    "name": { "type": "string" }
  }
}`

// memStore is an in-memory resource.Store that records what it was asked.
type memStore struct {
	mu       sync.Mutex
	rows     []resource.Record
	next     int
	lastList resource.ListParams
	lastBody resource.Record
	cred     string // credential seen on the last call
	needCred string // when set, calls without it answer 401
}

func (s *memStore) check(ctx context.Context) error {
	cred, _ := backend.CredentialFromContext(ctx)
	s.cred = cred
	if s.needCred != "" && cred != s.needCred {
		return backend.ErrUnauthorized
	}
	return nil
}

func (s *memStore) List(ctx context.Context, p resource.ListParams) (resource.ListResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(ctx); err != nil {
		return resource.ListResult{}, err
	}
	s.lastList = p
	return resource.ListResult{Items: append([]resource.Record(nil), s.rows...), TotalItems: len(s.rows), Page: p.Page, PerPage: p.PerPage}, nil
}

func (s *memStore) Get(ctx context.Context, id string) (resource.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	for _, r := range s.rows {
		if r["id"] == id {
			cp := resource.Record{}
			for k, v := range r {
				cp[k] = v
			}
			return cp, nil
		}
	}
	return nil, backend.ErrNotFound
}

func (s *memStore) Create(ctx context.Context, body resource.Record) (resource.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	s.lastBody = body
	s.next++
	rec := resource.Record{"id": "n" + strconv.Itoa(s.next)}
	for k, v := range body {
		rec[k] = v
	}
	s.rows = append(s.rows, rec)
	return rec, nil
}

func (s *memStore) Update(ctx context.Context, id string, body resource.Record) (resource.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	s.lastBody = body
	for i, r := range s.rows {
		if r["id"] == id {
			for k, v := range body {
				s.rows[i][k] = v
			}
			return s.rows[i], nil
		}
	}
	return nil, backend.ErrNotFound
}

func (s *memStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(ctx); err != nil {
		return err
	}
	for i, r := range s.rows {
		if r["id"] == id {
			s.rows = append(s.rows[:i], s.rows[i+1:]...)
			return nil
		}
	}
	return backend.ErrNotFound
}

type env struct {
	h       http.Handler
	posts   *memStore
	authors *memStore
	reg     *resource.Registry
}

func runtime(t *testing.T, store *memStore, name, label, schema string, auth resource.AuthScheme, caps resource.Capabilities) *resource.Runtime {
	t.Helper()
	fields, err := resource.FieldsFromJSONSchema([]byte(schema))
	if err != nil {
		t.Fatalf("fields: %v", err)
	}
	if name == "posts" {
		for i := range fields {
			if fields[i].Name == "rating" {
				fields[i].Hint = &resource.FieldHint{Widget: resource.WidgetSlider, Min: 0, Max: 5, Step: 1}
			}
		}
	}
	v, err := schemavalidation.New([]byte(schema))
	if err != nil {
		t.Fatalf("validator: %v", err)
	}
	d := resource.ResourceDescriptor{
		Backend:      resource.Backend{ID: "ex", Label: "Example", BaseURL: "http://x"},
		Name:         name,
		Label:        label,
		Schema:       resource.Schema{Raw: []byte(schema), Fields: fields},
		Capabilities: caps,
		AuthScheme:   auth,
		IDField:      "id",
	}
	return &resource.Runtime{Descriptor: d, Service: &resource.Service{Descriptor: d, Store: store, Validator: v}}
}

var fullCaps = resource.Capabilities{List: true, Read: true, Create: true, Update: true, Delete: true, UpdateVerb: "PATCH"}

func newEnv(t *testing.T, auth resource.AuthScheme) *env {
	t.Helper()
	e := &env{posts: &memStore{}, authors: &memStore{}, reg: resource.NewRegistry()}
	e.authors.rows = []resource.Record{{"id": "a1", "name": "Ada"}, {"id": "a2", "name": "Grace"}}
	e.posts.rows = []resource.Record{
		{"id": "p1", "title": "Hello", "status": "draft", "featured": true, "author": "a1", "rating": 3.0, "tags": []any{"x", "y"}},
		{"id": "p2", "title": "World", "status": "published", "author": "a2"},
	}
	e.reg.Replace([]*resource.Runtime{
		runtime(t, e.posts, "posts", "Posts", postSchema, auth, fullCaps),
		runtime(t, e.authors, "authors", "Authors", authorSchema, auth, fullCaps),
	})
	return e
}

func (e *env) build(t *testing.T) {
	t.Helper()
	live, err := ui.NewLive(e.reg, ui.Options{PageSize: 25})
	if err != nil {
		t.Fatalf("build ui: %v", err)
	}
	e.h = live
}

func get(h http.Handler, target string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func post(h http.Handler, target string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func contains(t *testing.T, rec *httptest.ResponseRecorder, want ...string) {
	t.Helper()
	body := rec.Body.String()
	for _, w := range want {
		if !strings.Contains(body, w) {
			t.Errorf("response (%d) lacks %q", rec.Code, w)
		}
	}
}

var _ = fmt.Sprint

func resourceNone() resource.AuthScheme { return resource.AuthScheme{Type: "none"} }

func mustLive(t *testing.T, e *env) *ui.Live {
	t.Helper()
	live, err := ui.NewLive(e.reg, ui.Options{PageSize: 25})
	if err != nil {
		t.Fatalf("build ui: %v", err)
	}
	return live
}
