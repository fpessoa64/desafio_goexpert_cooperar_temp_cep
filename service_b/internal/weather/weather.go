package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type weatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// Service implements usecase.WeatherService using the WeatherAPI.com REST API.
type Service struct {
	tracer trace.Tracer
	apiKey string
}

func New(tracer trace.Tracer, apiKey string) *Service {
	return &Service{tracer: tracer, apiKey: apiKey}
}

func (s *Service) GetTemperature(ctx context.Context, city string) (float64, error) {
	ctx, span := s.tracer.Start(ctx, "weatherapi.GetTemperature")
	defer span.End()

	apiURL := fmt.Sprintf(
		"http://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=pt",
		s.apiKey,
		url.QueryEscape(city),
	)
	slog.Debug("querying weatherapi", "city", city)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("failed to build weatherapi request", "city", city, "error", err)
		return 0, fmt.Errorf("building weatherapi request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("weatherapi request failed", "city", city, "error", err)
		return 0, fmt.Errorf("calling weatherapi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("weatherapi returned status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("weatherapi unexpected status", "city", city, "status", resp.StatusCode)
		return 0, err
	}

	var data weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("failed to decode weatherapi response", "city", city, "error", err)
		return 0, fmt.Errorf("decoding weatherapi response: %w", err)
	}

	slog.Debug("temperature fetched", "city", city, "temp_c", data.Current.TempC)
	return data.Current.TempC, nil
}
