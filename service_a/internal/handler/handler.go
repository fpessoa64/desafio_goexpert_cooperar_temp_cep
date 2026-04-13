package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
)

var validCEP = regexp.MustCompile(`^\d{8}$`)

// Forwarder is the port for sending a validated CEP to the downstream service.
// Defined here (consumer side) following the Go interface convention.
type Forwarder interface {
	Forward(ctx context.Context, cep string) (statusCode int, body []byte, err error)
}

// CEPHandler handles POST /cep requests.
// It owns only HTTP concerns: decoding, validation and response writing.
type CEPHandler struct {
	forwarder Forwarder
}

func New(f Forwarder) *CEPHandler {
	return &CEPHandler{forwarder: f}
}

func (h *CEPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("method not allowed", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CEP string `json:"cep"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CEP == "" {
		slog.Warn("invalid request body", "remote_addr", r.RemoteAddr)
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	if !validCEP.MatchString(req.CEP) {
		slog.Warn("invalid cep format", "cep", req.CEP, "remote_addr", r.RemoteAddr)
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	slog.Info("forwarding cep to service_b", "cep", req.CEP)

	statusCode, body, err := h.forwarder.Forward(r.Context(), req.CEP)
	if err != nil {
		slog.Error("forward failed", "cep", req.CEP, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	slog.Info("request completed", "cep", req.CEP, "status", statusCode)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}
