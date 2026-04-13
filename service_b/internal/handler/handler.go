package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/fernandopessoa/coopera_desafio/service_b/internal/domain"
	"github.com/fernandopessoa/coopera_desafio/service_b/internal/usecase"
)

var validCEP = regexp.MustCompile(`^\d{8}$`)

// UseCase is the port for the weather use case.
// Defined here (consumer side) following the Go interface convention.
type UseCase interface {
	Execute(ctx context.Context, cep string) (*usecase.WeatherResult, error)
}

// WeatherHandler handles POST /weather requests.
// It owns only HTTP concerns: decoding, validation and response writing.
type WeatherHandler struct {
	useCase UseCase
}

func New(uc UseCase) *WeatherHandler {
	return &WeatherHandler{useCase: uc}
}

func (h *WeatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	slog.Info("processing weather request", "cep", req.CEP)

	result, err := h.useCase.Execute(r.Context(), req.CEP)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			slog.Warn("cep not found", "cep", req.CEP)
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}
		slog.Error("use case error", "cep", req.CEP, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	slog.Info("weather request completed", "cep", req.CEP, "city", result.City, "temp_c", result.TempC)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
