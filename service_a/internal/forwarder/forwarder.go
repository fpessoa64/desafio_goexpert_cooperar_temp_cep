package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPForwarder implements handler.Forwarder by sending HTTP requests to Service B.
// The otelhttp transport injects W3C trace context headers automatically.
type HTTPForwarder struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string) *HTTPForwarder {
	return &HTTPForwarder{
		baseURL: baseURL,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (f *HTTPForwarder) Forward(ctx context.Context, cep string) (int, []byte, error) {
	payload, _ := json.Marshal(struct {
		CEP string `json:"cep"`
	}{CEP: cep})

	target := fmt.Sprintf("%s/weather", f.baseURL)
	slog.Debug("calling service_b", "url", target, "cep", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		slog.Error("failed to build request", "url", target, "error", err)
		return 0, nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		slog.Error("service_b call failed", "url", target, "error", err)
		return 0, nil, fmt.Errorf("forwarding request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read service_b response", "error", err)
		return 0, nil, fmt.Errorf("reading response body: %w", err)
	}

	slog.Debug("service_b responded", "status", resp.StatusCode, "cep", cep)

	return resp.StatusCode, body, nil
}
