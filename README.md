# go-template

Production-ready Go service template built on:

- **[Echo v5](https://echo.labstack.com/)** — HTTP framework
- **[GORM](https://gorm.io/) + PostgreSQL** — relational database
- **[go-redis v9](https://github.com/redis/go-redis)** — Redis client (single / cluster / sentinel)
- **[Paho MQTT](https://github.com/eclipse/paho.mqtt.golang)** — MQTT pub/sub via EMQX
- **[Viper](https://github.com/spf13/viper)** — config file + env var override
- **[OpenTelemetry](https://opentelemetry.io/)** — distributed tracing (OTLP/gRPC)
- **[slog](https://pkg.go.dev/log/slog)** — structured JSON logging

Architecture: Clean layered CQRS — Handler → App (Command/Query) → Repository/Adapter.  
See [`docs/GUIDELINE.md`](docs/GUIDELINE.md) for the full architectural reference.

---

## Requirements

| Tool | Version |
|---|---|
| Go | 1.26+ |
| Docker & Docker Compose | any recent version |
| golangci-lint | for `make lint` |

---

## Quick start

### 1. Start infrastructure

```sh
docker compose up -d postgres redis emqx
```

> `otel-collector` is optional — only needed when `tracing.enabled: true`.

### 2. Run the service locally

```sh
make run
# or
go run ./cmd/main.go
```

The server starts on `:5000` by default.

### 3. Run everything in Docker

```sh
docker compose up -d
```

The `app` service builds from the local `Dockerfile` and connects to the other containers via Docker's internal network.

---

## Configuration

Default config: `config/config.yml`

Override the config file path:
```sh
CONFIG_PATH=/path/to/config/config go run ./cmd/main.go
```

Override individual keys via environment variables (Viper maps `.` → `_`, uppercase):

| Env var | Config key | Example |
|---|---|---|
| `DATABASE_HOST` | `database.host` | `postgres` |
| `DATABASE_PASSWORD` | `database.password` | `secret` |
| `REDIS_ADDRESS` | `redis.address` | `redis:6379` |
| `MQTT_BROKER` | `mqtt.broker` | `tcp://emqx:1883` |
| `LOGGER_LEVEL` | `logger.level` | `debug` |
| `TRACING_ENABLED` | `tracing.enabled` | `true` |

Docker Compose already sets `DATABASE_HOST`, `REDIS_ADDRESS`, and `MQTT_BROKER` for the `app` service.

---

## Project structure

```
cmd/                    ← entry point
config/                 ← AppConfig struct + config.yml
internal/
  server.go             ← dependency wiring
  http.go               ← Echo setup (middleware, validator, error handler)
  router/routes.go      ← all HTTP routes
  domain/               ← pure types (models, DTOs, msgs, enums)
  app/<domain>/         ← CQRS commands and queries
  handler/<domain>/     ← HTTP handlers
  worker/event/         ← MQTT workers
  infras/               ← repository, adapter, mqttpub implementations
pkgs/
  httpclient/           ← Resty client factory + circuit breaker
  probe/                ← framework-agnostic health check (net/http)
  db/orm/               ← GORM connection factory
  db/rdclient/          ← Redis client factory
  mqtt/                 ← MQTT client wrapper
  tracing/              ← OpenTelemetry initialisation
  logger/               ← slog handler
  utilities/            ← response, svcerr, paging, mapper, validate, decorator
```

---

## HTTP endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/liveness` | Always `200 ok` — process is alive |
| `GET` | `/readiness` | `200 ok` / `503 degraded` — checks DB + Redis |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/api/v1/users` | Create user |
| `GET` | `/api/v1/users` | List users (paginated) |
| `GET` | `/api/v1/users/:id` | Get user by ID |
| `PUT` | `/api/v1/users/:id` | Update user |
| `DELETE` | `/api/v1/users/:id` | Delete user |

---

## Key features

### Health probe (`pkgs/probe`)

Framework-agnostic readiness check. Register any component that implements `probe.Checker`:

```go
readyProbe := probe.New(3 * time.Second).
    Register("postgres", probe.DBChecker(database)).
    Register("redis",    probe.RedisChecker(redisClient))
```

Returns per-component details on failure:
```json
{
  "status": "degraded",
  "details": { "postgres": "ok", "redis": "dial tcp: connection refused" }
}
```

### Circuit breaker (`pkgs/httpclient`)

Optional circuit breaker at the `http.RoundTripper` level — transparent to all resty calls:

```go
httpclient.New(httpclient.Config{
    BaseURL: "https://api.example.com",
    CircuitBreaker: &httpclient.CBConfig{
        MaxFailures:    5,
        HalfOpenProbes: 2,
        OpenTimeout:    10 * time.Second,
    },
})
```

States: `Closed` → `Open` → `HalfOpen` → `Closed`.  
Returns `httpclient.ErrCircuitOpen` immediately when the circuit is open.

### Error handler (`pkgs/utilities/svcerr`)

Unified Echo error handler. Resolution order:

1. `*svcerr.Error` — domain typed error
2. `validator.ValidationErrors` — 400 validation failure
3. `echo.HTTPStatusCoder` — generic HTTP error from framework
4. Default — 500 internal server error

---

## Make targets

```sh
make run          # run locally
make build        # build binary → build/app-bin
make test         # go test ./...
make lint         # go vet + golangci-lint
make fmt          # go fmt ./...
make tidy         # go mod tidy
make mocks        # regenerate mocks via mockery
make dockerize    # docker build -t go-service:local .
make clean        # remove build/
```

---

## Development guide

Full architectural reference, layer contracts, patterns for commands/queries/handlers/workers, and step-by-step checklist for adding a new domain:

**[docs/GUIDELINE.md](docs/GUIDELINE.md)**
