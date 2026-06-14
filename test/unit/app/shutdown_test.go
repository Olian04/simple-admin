package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/Olian04/simple-admin/internal/app"
	"github.com/Olian04/simple-admin/internal/config"
)

// A signal-driven (canceled-context) shutdown must return nil, not an error,
// so the process exits 0 on Ctrl+C.
func TestRunReturnsNilOnContextCancel(t *testing.T) {
	cfg := config.Config{}.WithDefaults()
	cfg.HTTP.ListenAddr = "127.0.0.1:0"
	cfg.Metrics.ListenAddr = "127.0.0.1:0"

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled → graceful shutdown path

	done := make(chan error, 1)
	go func() { done <- app.Run(ctx, cfg, ".") }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected nil on graceful shutdown, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}
