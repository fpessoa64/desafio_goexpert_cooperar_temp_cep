package usecase

import (
	"context"
	"log/slog"

	"github.com/fernandopessoa/coopera_desafio/service_b/internal/temperature"
)

// LocationService is the port for resolving a CEP to a city name.
// Implementations must return domain.ErrNotFound when the CEP yields no result.
type LocationService interface {
	GetCity(ctx context.Context, cep string) (string, error)
}

// WeatherService is the port for fetching the current temperature of a city.
type WeatherService interface {
	GetTemperature(ctx context.Context, city string) (float64, error)
}

// WeatherResult is the use-case output returned to the delivery layer.
type WeatherResult struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// WeatherUseCase orchestrates CEP → city → temperature retrieval and conversion.
// It depends only on the two ports above, never on concrete infrastructure.
type WeatherUseCase struct {
	location LocationService
	weather  WeatherService
}

func New(l LocationService, w WeatherService) *WeatherUseCase {
	return &WeatherUseCase{location: l, weather: w}
}

func (uc *WeatherUseCase) Execute(ctx context.Context, cep string) (*WeatherResult, error) {
	slog.Debug("resolving city from cep", "cep", cep)

	city, err := uc.location.GetCity(ctx, cep)
	if err != nil {
		slog.Error("city lookup failed", "cep", cep, "error", err)
		return nil, err
	}

	slog.Debug("fetching temperature", "cep", cep, "city", city)

	tempC, err := uc.weather.GetTemperature(ctx, city)
	if err != nil {
		slog.Error("temperature fetch failed", "city", city, "error", err)
		return nil, err
	}

	slog.Info("use case completed", "cep", cep, "city", city, "temp_c", tempC)

	return &WeatherResult{
		City:  city,
		TempC: tempC,
		TempF: temperature.ToFahrenheit(tempC),
		TempK: temperature.ToKelvin(tempC),
	}, nil
}
