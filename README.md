# Coopera Desafio — CEP Weather + Distributed Tracing

Sistema distribuído em Go com dois microsserviços que consultam o clima de uma cidade a partir de um CEP, instrumentados com OpenTelemetry e Zipkin.

## Pré-requisitos

- Docker e Docker Compose instalados
- Chave de API gratuita do [WeatherAPI.com](https://www.weatherapi.com/)

## Como executar

```bash
WEATHER_API_KEY=sua_chave docker compose up --build
```

Aguarde todos os serviços iniciarem. A ordem de dependência é:
`zipkin` → `otel-collector` → `service_b` → `service_a`

---

## Realizando requisição POST no Serviço A

O Serviço A escuta na porta **8080** e aceita `POST /cep` com o CEP no corpo JSON.

### Requisição válida

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}'
```

**Resposta esperada (200):**
```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

### CEP inválido (não tem 8 dígitos)

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

**Resposta esperada (422):**
```
invalid zipcode
```

### CEP não encontrado

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}'
```

**Resposta esperada (404):**
```
can not find zipcode
```

---

## Visualizando os traços no Zipkin

O Zipkin fica disponível em:

```
http://localhost:9411
```

### Passos para visualizar uma requisição

1. Acesse `http://localhost:9411` no navegador
2. Faça uma requisição POST ao Serviço A (exemplos acima)
3. Na interface do Zipkin, clique em **"Run Query"**
4. Localize o trace com o nome de serviço `service_a` e clique nele

### O que você verá

Cada requisição gera um trace com **4 spans** encadeados:

```
service_a: POST /cep
└── service_b: POST /weather
    ├── viacep.GetCity          (chamada à API ViaCEP)
    └── weatherapi.GetTemperature  (chamada à API WeatherAPI)
```

O `TraceID` é compartilhado entre os dois serviços, permitindo visualizar o fluxo completo da requisição de ponta a ponta.

---

## Fórmulas de conversão de temperatura

| Escala | Fórmula |
|---|---|
| Fahrenheit | `F = C × 1.8 + 32` |
| Kelvin | `K = C + 273` |

---

## Portas

| Serviço | Porta |
|---|---|
| Service A (entrada) | `8080` |
| Service B (interno) | `8081` |
| Zipkin UI | `9411` |
| OTEL Collector (gRPC) | `4317` |
