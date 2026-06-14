// Package openapi builds resource descriptors by fetching and parsing an
// OpenAPI 3 spec (discovery mode). It uses a focused encoding/json parser
// sufficient for JSON specs with local component $refs (see .plan decisions);
// libopenapi can replace it later if richer parsing is needed.
package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/domain/resource"
)

var itemPathRE = regexp.MustCompile(`/\{[^/]+\}$`)

// Source is a resource.DescriptorSource over the discovery backends in config.
type Source struct {
	backends []config.BackendSection
	client   *http.Client
}

func New(backends []config.BackendSection) *Source {
	return &Source{backends: backends, client: &http.Client{Timeout: 10 * time.Second}}
}

// Descriptors implements resource.DescriptorSource by fetching each discovery
// backend's spec and parsing it.
func (s *Source) Descriptors(ctx context.Context) ([]resource.ResourceDescriptor, error) {
	var out []resource.ResourceDescriptor
	for _, b := range s.backends {
		if b.Discovery == nil {
			continue
		}
		raw, err := s.fetch(ctx, b.Discovery)
		if err != nil {
			return nil, fmt.Errorf("backend %q discovery: %w", b.ID, err)
		}
		backend := resource.Backend{ID: b.ID, Label: b.Label, BaseURL: b.BaseURL}
		descs, err := Parse(raw, backend)
		if err != nil {
			return nil, fmt.Errorf("backend %q parse spec: %w", b.ID, err)
		}
		out = append(out, descs...)
	}
	return out, nil
}

func (s *Source) fetch(ctx context.Context, d *config.DiscoverySection) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
	if err != nil {
		return nil, err
	}
	if d.Auth != nil && d.Auth.Type == "apikey" && d.Auth.Value != "" {
		switch d.Auth.In {
		case "header":
			req.Header.Set(d.Auth.Name, d.Auth.Value)
		case "query":
			q := req.URL.Query()
			q.Set(d.Auth.Name, d.Auth.Value)
			req.URL.RawQuery = q.Encode()
		}
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", d.URL, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

type specDoc struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas         map[string]json.RawMessage `json:"schemas"`
		SecuritySchemes map[string]securityScheme  `json:"securitySchemes"`
	} `json:"components"`
}

type securityScheme struct {
	Type   string `json:"type"`
	In     string `json:"in"`
	Name   string `json:"name"`
	Scheme string `json:"scheme"`
}

// Parse derives descriptors from an OpenAPI 3 JSON spec for one backend. It is
// the pure, testable core of the source.
func Parse(raw []byte, backend resource.Backend) ([]resource.ResourceDescriptor, error) {
	var doc specDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi json: %w", err)
	}

	auth := deriveAuthScheme(doc.Components.SecuritySchemes)

	pathSet := map[string]bool{}
	for p := range doc.Paths {
		pathSet[p] = true
	}

	// Visit collection paths in a stable order.
	collections := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		if !itemPathRE.MatchString(p) {
			collections = append(collections, p)
		}
	}
	sort.Strings(collections)

	var out []resource.ResourceDescriptor
	for _, coll := range collections {
		collOps := doc.Paths[coll]
		itemPath := coll + "/{id}"
		// Locate the actual item path (param may be named anything).
		actualItem := ""
		for p := range doc.Paths {
			if itemPathRE.MatchString(p) && itemPathRE.ReplaceAllString(p, "") == coll {
				actualItem = p
				break
			}
		}
		itemOps := map[string]json.RawMessage{}
		if actualItem != "" {
			itemOps = doc.Paths[actualItem]
		}

		schemaName := resolveSchemaName(collOps, itemOps)
		if schemaName == "" {
			continue // not a CRUD resource we can render
		}
		schemaRaw, ok := doc.Components.Schemas[schemaName]
		if !ok {
			continue
		}
		fields, err := resource.FieldsFromJSONSchema(schemaRaw)
		if err != nil {
			return nil, fmt.Errorf("schema %q: %w", schemaName, err)
		}

		caps := resource.Capabilities{
			List:   hasOp(collOps, "get"),
			Create: hasOp(collOps, "post"),
			Read:   hasOp(itemOps, "get"),
			Delete: hasOp(itemOps, "delete"),
		}
		switch {
		case hasOp(itemOps, "put"):
			caps.Update, caps.UpdateVerb = true, "PUT"
		case hasOp(itemOps, "patch"):
			caps.Update, caps.UpdateVerb = true, "PATCH"
		}

		out = append(out, resource.ResourceDescriptor{
			Backend:      backend,
			Name:         strings.ToLower(schemaName),
			Label:        schemaName,
			Schema:       resource.Schema{Raw: schemaRaw, Fields: fields},
			Capabilities: caps,
			AuthScheme:   auth,
			IDField:      "id",
			CRUDPaths:    resource.CRUDPaths{Collection: coll, Item: itemPath},
		})
	}
	return out, nil
}

func resolveSchemaName(collOps, itemOps map[string]json.RawMessage) string {
	// Prefer the single-object item GET response, then create body, then list.
	if name := refFromOp(itemOps["get"], "response"); name != "" {
		return name
	}
	if name := refFromOp(collOps["post"], "request"); name != "" {
		return name
	}
	return refFromOp(collOps["get"], "response")
}

func refFromOp(opRaw json.RawMessage, kind string) string {
	if len(opRaw) == 0 {
		return ""
	}
	var op struct {
		RequestBody json.RawMessage            `json:"requestBody"`
		Responses   map[string]json.RawMessage `json:"responses"`
	}
	if err := json.Unmarshal(opRaw, &op); err != nil {
		return ""
	}
	if kind == "request" {
		return findRef(contentSchema(op.RequestBody))
	}
	for _, code := range []string{"200", "201"} {
		if r, ok := op.Responses[code]; ok {
			if name := findRef(contentSchema(r)); name != "" {
				return name
			}
		}
	}
	return ""
}

// contentSchema digs `content -> application/json -> schema` out of a
// requestBody or response object.
func contentSchema(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	var node struct {
		Content map[string]struct {
			Schema json.RawMessage `json:"schema"`
		} `json:"content"`
	}
	if json.Unmarshal(raw, &node) != nil {
		return nil
	}
	if c, ok := node.Content["application/json"]; ok {
		return c.Schema
	}
	for _, c := range node.Content {
		return c.Schema
	}
	return nil
}

// findRef recursively searches a schema fragment for the first
// "#/components/schemas/X" reference, descending into items and properties.items.
func findRef(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var node map[string]json.RawMessage
	if json.Unmarshal(raw, &node) != nil {
		return ""
	}
	if r, ok := node["$ref"]; ok {
		var s string
		if json.Unmarshal(r, &s) == nil {
			if name := refLeaf(s); name != "" {
				return name
			}
		}
	}
	if items, ok := node["items"]; ok {
		if name := findRef(items); name != "" {
			return name
		}
	}
	if props, ok := node["properties"]; ok {
		var pm map[string]json.RawMessage
		if json.Unmarshal(props, &pm) == nil {
			if items, ok := pm["items"]; ok {
				if name := findRef(items); name != "" {
					return name
				}
			}
		}
	}
	return ""
}

func refLeaf(ref string) string {
	const prefix = "#/components/schemas/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ""
}

func deriveAuthScheme(schemes map[string]securityScheme) resource.AuthScheme {
	names := make([]string, 0, len(schemes))
	for n := range schemes {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		s := schemes[n]
		if s.Type == "apiKey" && (s.In == "header" || s.In == "query") {
			return resource.AuthScheme{Type: "apikey", In: s.In, Name: s.Name}
		}
	}
	// http bearer / others recorded but unsupported in the MVP.
	for _, n := range names {
		if schemes[n].Type == "http" && schemes[n].Scheme == "bearer" {
			return resource.AuthScheme{Type: "bearer"}
		}
	}
	return resource.AuthScheme{Type: "none"}
}

func hasOp(ops map[string]json.RawMessage, method string) bool {
	_, ok := ops[method]
	return ok
}
