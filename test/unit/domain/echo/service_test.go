package echo_test

import (
	"encoding/json"
	"testing"

	"github.com/Olian04/go-app-template/internal/domain/echo"
)

func TestEcho(t *testing.T) {
	svc := echo.NewService()
	got := svc.Echo(echo.Request{Message: " hello "})
	if got.Message != "hello" {
		t.Fatalf("got %q", got.Message)
	}
}

func TestEchoJSONViaHandlerShape(t *testing.T) {
	// Domain-only: ensure serialized round-trip sanity for handler contract.
	raw := []byte(`{"message":" hello "}`)
	var req echo.Request
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	out := echo.NewService().Echo(req)
	if out.Message != "hello" {
		t.Fatalf("got %#v", out)
	}
}
