package http

import (
	"io/fs"
	"log/slog"
	stdhttp "net/http"
	neturl "net/url"
	"time"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/observability/metrics"
	"github.com/Olian04/simple-admin/internal/transport/http/handlers"
	"github.com/Olian04/simple-admin/internal/backend"
	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/views/form"
	"github.com/Olian04/simple-admin/web"
)

func Router(registry *metrics.Registry, resources *resource.Registry, widgets *form.WidgetRegistry, ui config.UISection) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	// Root redirect + favicon. {$} anchors the exact root path so it does not
	// swallow other unmatched routes.
	mux.Handle("GET /{$}", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		stdhttp.Redirect(w, r, "/ui", stdhttp.StatusFound)
	}))
	mux.Handle("GET /favicon.ico", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	// Static assets (vendored Bulma/HTMX + app.css + islands).
	staticFS, _ := fs.Sub(web.Static, "static")
	mux.Handle("GET /static/", stdhttp.StripPrefix("/static/", stdhttp.FileServerFS(staticFS)))

	// Admin UI.
	uiHandler := handlers.NewUIHandler(resources, widgets, ui)
	mux.Handle("GET /ui", stdhttp.HandlerFunc(uiHandler.Index))
	mux.Handle("GET /ui/credentials", stdhttp.HandlerFunc(uiHandler.Credentials))
	mux.Handle("GET /ui/{backend}/{name}", stdhttp.HandlerFunc(uiHandler.List))
	mux.Handle("POST /ui/{backend}/{name}", stdhttp.HandlerFunc(uiHandler.Create))
	mux.Handle("GET /ui/{backend}/{name}/new", stdhttp.HandlerFunc(uiHandler.NewForm))
	mux.Handle("GET /ui/{backend}/{name}/options", stdhttp.HandlerFunc(uiHandler.Options))
	mux.Handle("GET /ui/{backend}/{name}/{id}", stdhttp.HandlerFunc(uiHandler.Detail))
	mux.Handle("GET /ui/{backend}/{name}/{id}/edit", stdhttp.HandlerFunc(uiHandler.EditForm))
	mux.Handle("PUT /ui/{backend}/{name}/{id}", stdhttp.HandlerFunc(uiHandler.Update))
	mux.Handle("GET /ui/{backend}/{name}/{id}/delete", stdhttp.HandlerFunc(uiHandler.DeleteConfirm))
	mux.Handle("DELETE /ui/{backend}/{name}/{id}", stdhttp.HandlerFunc(uiHandler.Delete))

	return middleware(securityHeaders(withCredentials(mux), ui), registry)
}

// securityHeaders sets a CSP. Self-hosted assets only need 'self'; the rich
// editor widgets (Monaco/EasyMDE) load from the configured CDNs, so those
// origins are allowlisted and style 'unsafe-inline' + blob/data workers are
// permitted (required by Monaco). Stricter when no CDN is configured.
func securityHeaders(next stdhttp.Handler, ui config.UISection) stdhttp.Handler {
	cdns := originList(ui.MonacoCDN, ui.EasyMDECDN)
	csp := "default-src 'self'; " +
		"script-src 'self'" + cdns + "; " +
		"style-src 'self' 'unsafe-inline'" + cdns + "; " +
		"img-src 'self' data:" + cdns + "; " +
		"font-src 'self'" + cdns + "; " +
		"connect-src 'self'" + cdns + "; " +
		"worker-src 'self' blob: data:; " +
		"base-uri 'self'; form-action 'self'"
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

// originList returns a space-prefixed, de-duplicated list of scheme://host
// origins for the given URLs (empty string if none parse).
func originList(urls ...string) string {
	seen := map[string]bool{}
	var out string
	for _, raw := range urls {
		u, err := neturl.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			continue
		}
		origin := u.Scheme + "://" + u.Host
		if seen[origin] {
			continue
		}
		seen[origin] = true
		out += " " + origin
	}
	return out
}

// withCredentials lifts the browser-supplied passthrough credential
// (X-Sa-Credential header) into the request context for the upstream client.
func withCredentials(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if cred := r.Header.Get("X-Sa-Credential"); cred != "" {
			r = r.WithContext(backend.WithCredential(r.Context(), cred))
		}
		next.ServeHTTP(w, r)
	})
}

func middleware(next stdhttp.Handler, registry *metrics.Registry) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		start := time.Now()
		registry.IncRequests()
		next.ServeHTTP(w, r)
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
