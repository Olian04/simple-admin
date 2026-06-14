// Package backend is the outbound CRUD client: it maps resource operations to
// HTTP calls against the backend, injecting passthrough credentials per the
// resource's auth scheme. It implements resource.Store.
package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// Client performs CRUD for a single resource against an upstream backend.
type Client struct {
	base       string
	paths      resource.CRUDPaths
	auth       resource.AuthScheme
	updateVerb string
	hc         *http.Client
}

// New builds a Client from a resource descriptor.
func New(d resource.ResourceDescriptor) *Client {
	verb := d.Capabilities.UpdateVerb
	if verb == "" {
		verb = http.MethodPut
	}
	return &Client{
		base:       strings.TrimRight(d.Backend.BaseURL, "/"),
		paths:      d.CRUDPaths,
		auth:       d.AuthScheme,
		updateVerb: verb,
		hc:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) collectionURL() string { return c.base + c.paths.Collection }

func (c *Client) itemURL(id string) string {
	return c.base + strings.ReplaceAll(c.paths.Item, "{id}", url.PathEscape(id))
}

// List fetches a page of records, accepting either a bare array or a
// pagination envelope ({items:[...], totalItems, ...}).
func (c *Client) List(ctx context.Context, params resource.ListParams) (resource.ListResult, error) {
	u := c.collectionURL()
	if params.Page > 0 || params.PerPage > 0 {
		q := url.Values{}
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			q.Set("perPage", strconv.Itoa(params.PerPage))
		}
		u += "?" + q.Encode()
	}

	body, err := c.do(ctx, http.MethodGet, u, nil)
	if err != nil {
		return resource.ListResult{}, err
	}

	var env struct {
		Items      []resource.Record `json:"items"`
		TotalItems int               `json:"totalItems"`
		Page       int               `json:"page"`
		PerPage    int               `json:"perPage"`
	}
	if json.Unmarshal(body, &env) == nil && env.Items != nil {
		return resource.ListResult{
			Items:      env.Items,
			TotalItems: env.TotalItems,
			Page:       env.Page,
			PerPage:    env.PerPage,
		}, nil
	}

	var arr []resource.Record
	if err := json.Unmarshal(body, &arr); err != nil {
		return resource.ListResult{}, fmt.Errorf("decode list response: %w", err)
	}
	return resource.ListResult{Items: arr, TotalItems: len(arr)}, nil
}

func (c *Client) Get(ctx context.Context, id string) (resource.Record, error) {
	body, err := c.do(ctx, http.MethodGet, c.itemURL(id), nil)
	if err != nil {
		return nil, err
	}
	return decodeRecord(body)
}

func (c *Client) Create(ctx context.Context, rec resource.Record) (resource.Record, error) {
	body, err := c.do(ctx, http.MethodPost, c.collectionURL(), rec)
	if err != nil {
		return nil, err
	}
	return decodeRecord(body)
}

func (c *Client) Update(ctx context.Context, id string, rec resource.Record) (resource.Record, error) {
	body, err := c.do(ctx, c.updateVerb, c.itemURL(id), rec)
	if err != nil {
		return nil, err
	}
	return decodeRecord(body)
}

func (c *Client) Delete(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, c.itemURL(id), nil)
	return err
}

// do performs the HTTP request, injects auth, and maps non-2xx to errors.
func (c *Client) do(ctx context.Context, method, u string, payload any) ([]byte, error) {
	var reqBody io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	applyAuth(ctx, req, c.auth)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, ErrUnauthorized
	default:
		return nil, &Error{Status: resp.StatusCode, Body: string(body)}
	}
}

func decodeRecord(body []byte) (resource.Record, error) {
	if len(body) == 0 {
		return resource.Record{}, nil
	}
	var rec resource.Record
	if err := json.Unmarshal(body, &rec); err != nil {
		return nil, fmt.Errorf("decode record: %w", err)
	}
	return rec, nil
}
