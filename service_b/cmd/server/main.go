package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fernandopessoa/coopera_desafio/service_b/internal/config"
	"github.com/fernandopessoa/coopera_desafio/service_b/internal/handler"
	"github.com/fernandopessoa/coopera_desafio/service_b/internal/usecase"
	"github.com/fernandopessoa/coopera_desafio/service_b/internal/viacep"
	"github.com/fernandopessoa/coopera_desafio/service_b/internal/weather"
	"github.com/fernandopessoa/coopera_desafio/service_b/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("config loaded",
		"port", cfg.Port,
		"otel_endpoint", cfg.OTELEndpoint,
		"service_name", cfg.ServiceName,
	)

	ctx := context.Background()

	shutdown, err := telemetry.InitTracer(ctx, cfg.OTELEndpoint, cfg.ServiceName)
	if err != nil {
		slog.Warn("tracer initialization failed, continuing without tracing", "error", err)
	} else {
		slog.Info("tracer initialized")
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				slog.Error("tracer shutdown error", "error", err)
			}
		}()
	}

	tracer := otel.Tracer(cfg.ServiceName)

	locationSvc := viacep.New(tracer)
	weatherSvc := weather.New(tracer, cfg.WeatherAPIKey)
	uc := usecase.New(locationSvc, weatherSvc)
	h := handler.New(uc)

	mux := http.NewServeMux()
	mux.Handle("/weather", otelhttp.NewHandler(h, "POST /weather"))

	slog.Info("service_b listening", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
