// Package http wires the admin UI and the operational endpoints into one mux.
package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/Olian04/simple-admin/internal/observability/metrics"
	"github.com/Olian04/simple-admin/internal/ui"
)

// Router serves the admin UI under ui.Prefix beside /healthz, and counts and
// logs every request.
func Router(registry *metrics.Registry, uiHandler stdhttp.Handler) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))
	// {$} anchors the exact root path so it does not swallow other routes.
	mux.Handle("GET /{$}", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		stdhttp.Redirect(w, r, ui.Prefix+"/", stdhttp.StatusFound)
	}))
	mux.Handle(ui.Prefix+"/", uiHandler)
	mux.Handle("GET "+ui.Prefix, stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		stdhttp.Redirect(w, r, ui.Prefix+"/", stdhttp.StatusFound)
	}))
	return middleware(mux, registry)
}

func middleware(next stdhttp.Handler, registry *metrics.Registry) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		start := time.Now()
		registry.IncRequests()
		next.ServeHTTP(w, r)
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
