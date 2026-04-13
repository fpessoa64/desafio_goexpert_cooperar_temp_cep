package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fernandopessoa/coopera_desafio/service_a/internal/config"
	"github.com/fernandopessoa/coopera_desafio/service_a/internal/forwarder"
	"github.com/fernandopessoa/coopera_desafio/service_a/internal/handler"
	"github.com/fernandopessoa/coopera_desafio/service_a/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	cfg := config.Load()
	slog.Info("config loaded",
		"port", cfg.Port,
		"service_b_url", cfg.ServiceBURL,
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

	fwd := forwarder.New(cfg.ServiceBURL)
	h := handler.New(fwd)

	mux := http.NewServeMux()
	mux.Handle("/cep", otelhttp.NewHandler(h, "POST /cep"))

	slog.Info("service_a listening", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
