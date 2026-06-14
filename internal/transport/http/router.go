package http

import (
	"io/fs"
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/Olian04/simple-admin/internal/domain/echo"
	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/observability/metrics"
	"github.com/Olian04/simple-admin/internal/transport/http/handlers"
	"github.com/Olian04/simple-admin/internal/ui/render"
	"github.com/Olian04/simple-admin/internal/upstream"
	"github.com/Olian04/simple-admin/web"
)

func Router(svc echo.Service, registry *metrics.Registry, resources *resource.Registry, widgets *render.Registry) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	echoHandler := handlers.NewEchoHandler(svc)
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))
	mux.Handle("POST /echo", echoHandler)

	// Static assets (vendored Bulma/HTMX + app.css + islands).
	staticFS, _ := fs.Sub(web.Static, "static")
	mux.Handle("GET /static/", stdhttp.StripPrefix("/static/", stdhttp.FileServerFS(staticFS)))

	// Admin UI.
	ui := handlers.NewUIHandler(resources, widgets)
	mux.Handle("GET /ui", stdhttp.HandlerFunc(ui.Index))
	mux.Handle("GET /ui/credentials", stdhttp.HandlerFunc(ui.Credentials))
	mux.Handle("GET /ui/{backend}/{name}", stdhttp.HandlerFunc(ui.List))
	mux.Handle("POST /ui/{backend}/{name}", stdhttp.HandlerFunc(ui.Create))
	mux.Handle("GET /ui/{backend}/{name}/new", stdhttp.HandlerFunc(ui.NewForm))
	mux.Handle("GET /ui/{backend}/{name}/options", stdhttp.HandlerFunc(ui.Options))
	mux.Handle("GET /ui/{backend}/{name}/{id}", stdhttp.HandlerFunc(ui.Detail))
	mux.Handle("GET /ui/{backend}/{name}/{id}/edit", stdhttp.HandlerFunc(ui.EditForm))
	mux.Handle("PUT /ui/{backend}/{name}/{id}", stdhttp.HandlerFunc(ui.Update))
	mux.Handle("GET /ui/{backend}/{name}/{id}/delete", stdhttp.HandlerFunc(ui.DeleteConfirm))
	mux.Handle("DELETE /ui/{backend}/{name}/{id}", stdhttp.HandlerFunc(ui.Delete))

	return middleware(securityHeaders(withCredentials(mux)), registry)
}

// securityHeaders sets a strict CSP. All scripts/styles are served from /static
// (no inline JS; inline styles moved to app.css), so 'self' suffices.
func securityHeaders(next stdhttp.Handler) stdhttp.Handler {
	const csp = "default-src 'self'; script-src 'self'; style-src 'self'; " +
		"img-src 'self' data:; connect-src 'self'; base-uri 'self'; form-action 'self'"
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

// withCredentials lifts the browser-supplied passthrough credential
// (X-Sa-Credential header) into the request context for the upstream client.
func withCredentials(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if cred := r.Header.Get("X-Sa-Credential"); cred != "" {
			r = r.WithContext(upstream.WithCredential(r.Context(), cred))
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
