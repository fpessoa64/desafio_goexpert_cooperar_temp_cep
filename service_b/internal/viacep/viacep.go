package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fernandopessoa/coopera_desafio/service_b/internal/domain"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type viaCEPResponse struct {
	Localidade string `json:"localidade"`
	Erro       any    `json:"erro"`
}

// Service implements usecase.LocationService using the ViaCEP public API.
type Service struct {
	tracer trace.Tracer
}

func New(tracer trace.Tracer) *Service {
	return &Service{tracer: tracer}
}

func (s *Service) GetCity(ctx context.Context, cep string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "viacep.GetCity")
	defer span.End()

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	slog.Debug("querying viacep", "cep", cep, "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("failed to build viacep request", "cep", cep, "error", err)
		return "", fmt.Errorf("building viacep request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("viacep request failed", "cep", cep, "error", err)
		return "", fmt.Errorf("calling viacep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		slog.Warn("cep not found in viacep", "cep", cep, "status", resp.StatusCode)
		return "", domain.ErrNotFound
	}

	var data viaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("failed to decode viacep response", "cep", cep, "error", err)
		return "", fmt.Errorf("decoding viacep response: %w", err)
	}

	// ViaCEP returns {"erro": true} for unfound CEPs
	if data.Erro != nil && data.Erro != false {
		slog.Warn("cep not found in viacep", "cep", cep)
		return "", domain.ErrNotFound
	}

	if data.Localidade == "" {
		slog.Warn("viacep returned empty city", "cep", cep)
		return "", domain.ErrNotFound
	}

	slog.Debug("viacep resolved city", "cep", cep, "city", data.Localidade)
	return data.Localidade, nil
}
