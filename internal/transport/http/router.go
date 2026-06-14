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
	"github.com/Olian04/simple-admin/web"
)

func Router(svc echo.Service, registry *metrics.Registry, resources *resource.Registry) stdhttp.Handler {
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
	ui := handlers.NewUIHandler(resources)
	mux.Handle("GET /ui", stdhttp.HandlerFunc(ui.Index))

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
