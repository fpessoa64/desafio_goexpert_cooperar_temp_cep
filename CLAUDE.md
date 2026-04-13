# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Distributed weather-by-CEP system in Go with OpenTelemetry distributed tracing. Two cooperating microservices expose the full request flow (Service A → Service B) in Zipkin.

---

## Running

```bash
# Start everything (Services A, B, OTEL Collector, Zipkin)
docker compose up --build

# Service A accepts POST requests
curl -X POST http://localhost:8080/cep -H "Content-Type: application/json" -d '{"cep": "29902555"}'

# Zipkin UI
open http://localhost:9411
```

---

## Architecture

### Service A — Input (`service_a/`)
- HTTP POST endpoint (default port **8080**)
- Accepts `{ "cep": "29902555" }` — CEP must be a string of exactly 8 digits
- Validates the CEP and forwards to Service B via HTTP
- Returns `422 invalid zipcode` on validation failure
- Propagates OTEL trace context to Service B

### Service B — Orchestration (`service_b/`)
- Receives a validated CEP from Service A (default port **8081**)
- Queries **ViaCEP** (`viacep.com.br`) to resolve city name — manual span
- Queries **WeatherAPI** (`weatherapi.com`) for current temperature — manual span
- Returns `{ "city": "São Paulo", "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.65 }`
- Returns `404 can not find zipcode` if ViaCEP finds nothing

### Temperature conversion formulas
- Fahrenheit: `F = C * 1.8 + 32`
- Kelvin: `K = C + 273`

### Observability stack
- Both services export OTLP traces to the **OTEL Collector**
- OTEL Collector forwards to **Zipkin** (`http://zipkin:9411`)
- Every incoming HTTP request is automatically traced; CEP lookup and temperature lookup each get a **manual child span**

---

## Docker requirements

- Multistage Dockerfiles: `FROM golang:1.21` build stage → `FROM scratch` final image
- `docker-compose.yaml` must start: `service_a`, `service_b`, `otel-collector`, `zipkin`

---

## API Contracts

**Service A**
```
POST /cep
Body: { "cep": "29902555" }
200 → proxies Service B response
422 → { "message": "invalid zipcode" }
```

**Service B**
```
POST /weather  (or similar, called internally by Service A)
200 → { "city": "...", "temp_C": 0.0, "temp_F": 0.0, "temp_K": 0.0 }
404 → { "message": "can not find zipcode" }
422 → { "message": "invalid zipcode" }
```

---

## Environment Variables

| Variable | Service | Description |
|---|---|---|
| `WEATHER_API_KEY` | B | WeatherAPI.com API key |
| `SERVICE_B_URL` | A | Base URL of Service B (e.g. `http://service_b:8081`) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | A, B | OTEL Collector endpoint (e.g. `http://otel-collector:4317`) |
| `OTEL_SERVICE_NAME` | A, B | Service name shown in Zipkin |
