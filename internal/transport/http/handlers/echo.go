package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Olian04/go-app-template/internal/domain/echo"
)

type EchoHandler struct{ service echo.Service }

func NewEchoHandler(service echo.Service) EchoHandler {
	return EchoHandler{service: service}
}

func (h EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req echo.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := h.service.Echo(req)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
