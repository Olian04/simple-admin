package metricshttp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func Serve(ctx context.Context, addr string, handler http.Handler) error {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusFound)
	}))
	mux.Handle("GET /metrics", handler)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("metrics shutdown", "error", err)
		}
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("metrics server: %w", err)
	}
	return nil
}
