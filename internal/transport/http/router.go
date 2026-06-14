package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/Olian04/go-app-template/internal/domain/echo"
	"github.com/Olian04/go-app-template/internal/observability/metrics"
	"github.com/Olian04/go-app-template/internal/transport/http/handlers"
)

func Router(svc echo.Service, registry *metrics.Registry) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	echoHandler := handlers.NewEchoHandler(svc)
	mux.Handle("GET /healthz", stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))
	mux.Handle("POST /echo", echoHandler)
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
