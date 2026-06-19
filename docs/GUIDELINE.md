# Development Guideline

This document is the single source of truth for implementing new features in this template.
It covers project structure, layer contracts, conventions, and the exact steps to replace the sample `user` domain with your own business domain.

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Layer Contracts](#2-layer-contracts)
3. [Domain Layer](#3-domain-layer)
4. [Application Layer — CQRS](#4-application-layer--cqrs)
5. [Handler Layer — HTTP](#5-handler-layer--http)
6. [Worker Layer — MQTT](#6-worker-layer--mqtt)
7. [Infrastructure Layer](#7-infrastructure-layer)
8. [Shared Packages](#8-shared-packages)
9. [Configuration](#9-configuration)
10. [Error Handling](#10-error-handling)
11. [Implementing a New Domain](#11-implementing-a-new-domain)
12. [Checklist — New Feature End-to-End](#12-checklist--new-feature-end-to-end)

---

## 1. Project Structure

```
.
├── cmd/
│   └── main.go                     ← Entry point. Starts HTTP + Worker components.
├── config/
│   ├── config.go                   ← AppConfig struct and all sub-configs.
│   └── config.yml                  ← Default values. Override via env vars.
├── internal/                       ← Everything that must NOT be imported by external packages.
│   ├── server.go                   ← Wires all dependencies. The only place that calls constructors.
│   ├── http.go                     ← Echo instance setup (middleware, validator, error handler).
│   ├── router/
│   │   └── routes.go               ← Registers all HTTP routes. One place, no magic.
│   ├── domain/                     ← Pure business types — no framework imports.
│   │   ├── models/                 ← GORM models (DB schema).
│   │   ├── dtos/                   ← Request/Response structs per domain.
│   │   ├── msgs/                   ← MQTT message payload structs.
│   │   └── enums/                  ← Shared constants/error codes.
│   ├── app/                        ← Application logic. One sub-package per domain.
│   │   └── <domain>/
│   │       ├── application.go      ← Aggregates commands + queries for the domain.
│   │       ├── command/            ← Write operations (create, update, delete).
│   │       └── query/              ← Read operations (get, list).
│   ├── handler/                    ← HTTP handlers. One sub-package per domain.
│   │   └── <domain>/
│   │       └── handler.go
│   ├── worker/                     ← MQTT subscribers. Transport layer, like handlers.
│   │   ├── worker_group.go         ← WorkerGroup (Register + StartGroup). Framework code — do not modify.
│   │   └── event/                  ← Child workers, one file per topic.
│   └── infras/                     ← Infrastructure implementations.
│       ├── repository/<domain>/    ← DB access. Implements domain repository interface.
│       ├── adapter/<service>/      ← Outbound REST adapters. Implements domain adapter interface.
│       └── mqttpub/                ← MQTT publisher. Wraps pkgs/mqtt.
├── pkgs/                           ← Shared, reusable packages. No business logic.
│   ├── caller/                     ← Resty HTTP client factory.
│   ├── db/orm/                     ← GORM connection factory.
│   ├── db/rdclient/                ← Redis client factory.
│   ├── decorator/                  ← Generic command/query handler interfaces.
│   ├── logger/                     ← slog JSON handler.
│   ├── mqtt/                       ← Paho MQTT client wrapper.
│   └── ultis/
│       ├── mapper/                 ← Struct-to-struct and struct-to-map helpers.
│       ├── paging/                 ← Page/size pagination (ListQuery, ListResponse).
│       ├── response/               ← HTTP response envelope (BaseRes).
│       ├── svcerr/                 ← Typed errors + Echo error handler.
│       └── validate/               ← go-playground/validator singleton.
└── docker/
    └── mosquitto/mosquitto.conf    ← Mosquitto broker config for local dev.
```

---

## 2. Layer Contracts

```
HTTP Request ──► Handler ──► App (Command/Query) ──► Repository / Adapter
MQTT Message ──► Worker  ──► App (Command/Query) ──► Repository / Adapter
```

| Layer | Responsibility | Allowed dependencies |
|---|---|---|
| Handler / Worker | Decode input, validate format, call app, encode response | `app`, `domain/dtos`, `domain/msgs`, `pkgs/ultis` |
| App (Command/Query) | Business logic and orchestration | `domain/models`, `domain/dtos`, `infras` interfaces, `pkgs/ultis/svcerr` |
| Repository / Adapter | Data access and external calls | `domain/models`, ORM / HTTP client |
| Domain | Pure types — no behaviour | Nothing |

**Rules:**
- A lower layer must **never** import a higher layer.
- `infras` implementations are injected via interfaces defined in the domain or app layer.
- Handlers and Workers are parallel entry points — they call the same app layer.

---

## 3. Domain Layer

### Model (`internal/domain/models/`)

GORM model. Maps directly to a DB table.

```go
// internal/domain/models/order.go
package models

import "time"

type Order struct {
    ID        uint   `gorm:"primaryKey"`
    UserID    uint   `gorm:"not null;index"`
    Status    string `gorm:"size:50;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### DTOs (`internal/domain/dtos/<domain>/`)

Request and response types. Validation tags live here — **only here**.

```go
// internal/domain/dtos/orderdtos/order_dto.go
package orderdtos

type CreateOrderReq struct {
    UserID uint   `json:"user_id" validate:"required"`
    Item   string `json:"item"    validate:"required,min=2"`
}

type OrderDTO struct {
    ID     uint   `json:"id"`
    UserID uint   `json:"user_id"`
    Status string `json:"status"`
}

func NewOrderDTO(m *models.Order) *OrderDTO { ... }
```

### Messages (`internal/domain/msgs/`)

MQTT payload structs. Mirrors DTOs but for message-based transport.

```go
// internal/domain/msgs/order_event_msg.go
package msgs

type OrderEventPayload struct {
    Action  string `json:"action"`   // "created" | "cancelled"
    OrderID uint   `json:"order_id"`
}
```

### Enums (`internal/domain/enums/`)

Numeric error codes shared across layers. Add new codes here if needed.
`svcerr.Error.Code` uses these values.

---

## 4. Application Layer — CQRS

Each domain has one `Application` struct that bundles all commands and queries.

### Command (write — returns value)

```go
// internal/app/order/command/create_order.go
package command

import (
    "github.com/tdatIT/go-template/.agents/skills/go-clean-cqrs-architecture/references/decorator"
    "github.com/tdatIT/go-template/internal/domain/dtos/orderdtos"
)

// Declare the interface using the decorator generic — keeps the app layer testable.
type ICreateOrderCommand decorator.CommandReturnHandler[*orderdtos.CreateOrderReq, *orderdtos.OrderDTO]

type createOrderCommand struct{ repo order.Repository }

func NewCreateOrderCommand(repo order.Repository) ICreateOrderCommand {
    return &createOrderCommand{repo: repo}
}

func (c *createOrderCommand) Handle(ctx context.Context, req *orderdtos.CreateOrderReq) (*orderdtos.OrderDTO, error) {
    // Business logic here. Do NOT re-validate nil/zero — the handler already did.
    model := &models.Order{UserID: req.UserID, Status: "pending"}
    if err := c.repo.Create(ctx, model); err != nil {
        return nil, svcerr.ErrInternalServer
    }
    return orderdtos.NewOrderDTO(model), nil
}
```

### Command (write — no return value)

Use `decorator.CommandHandler[T]` instead of `CommandReturnHandler[T, E]`.

### Query (read)

```go
type IGetOrderQuery decorator.QueryHandler[*orderdtos.GetOrderReq, *orderdtos.OrderDTO]
```

### Application aggregator

```go
// internal/app/order/application.go
type Application struct {
    Command commands
    Query   queries
}

type commands struct {
    CreateOrder command.ICreateOrderCommand
}

type queries struct {
    GetOrder query.IGetOrderQuery
}

func NewOrderApplication(repo order.Repository) *Application {
    return &Application{
        Command: commands{CreateOrder: command.NewCreateOrderCommand(repo)},
        Query:   queries{GetOrder:    query.NewGetOrderQuery(repo)},
    }
}
```

---

## 5. Handler Layer — HTTP

One handler file per domain. Handlers decode, validate, call app, encode response.

```go
// internal/handler/order/handler.go
package order

type Handler struct{ app *orderapp.Application }

func NewOrderHandler(app *orderapp.Application) *Handler { return &Handler{app: app} }

func (h *Handler) CreateOrder(c *echo.Context) error {
    var req orderdtos.CreateOrderReq
    if err := c.Bind(&req); err != nil {
        return svcerr.ErrBadRequest           // bind failure
    }
    if err := c.Validate(&req); err != nil {
        return err                             // validation failure — already a svcerr
    }

    data, err := h.app.Command.CreateOrder.Handle(c.Request().Context(), &req)
    if err != nil {
        return err
    }

    res := response.SuccessRes
    res.Data = data
    return res.JSON(c)
}
```

**Register in router:**

```go
// internal/router/routes.go
orderRoute := v1.Group("/orders")
orderRoute.POST("", orderHandler.CreateOrder)
orderRoute.GET("/:id", orderHandler.GetOrder)
```

**Parsing helpers (keep in handler file):**

```go
func parseIDParam(c *echo.Context) (uint, error) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil || id == 0 {
        return 0, svcerr.ErrInvalidIdParam
    }
    return uint(id), nil
}
```

---

## 6. Worker Layer — MQTT

Workers are the MQTT equivalent of HTTP handlers — same app layer, different transport.

### Child worker (`internal/worker/event/`)

```go
// internal/worker/event/order_event_worker.go
package event

type OrderEventWorker struct {
    client mqttclient.Client
    app    *orderapp.Application
    cfg    *config.WorkerConfig
}

func NewOrderEventWorker(client mqttclient.Client, app *orderapp.Application, cfg *config.WorkerConfig) *OrderEventWorker {
    return &OrderEventWorker{client: client, app: app, cfg: cfg}
}

func (w *OrderEventWorker) Start(ctx context.Context) error {
    err := w.client.Subscribe(w.cfg.Topic, w.cfg.QoS, func(topic string, payload []byte) {
        go w.handle(ctx, payload)   // non-blocking: never hold the MQTT message loop
    })
    if err != nil {
        return fmt.Errorf("order event worker: subscribe %s: %w", w.cfg.Topic, err)
    }
    <-ctx.Done()
    return nil
}

func (w *OrderEventWorker) handle(ctx context.Context, raw []byte) {
    var event msgs.OrderEventPayload
    if err := json.Unmarshal(raw, &event); err != nil {
        slog.Error("order event: malformed payload", slog.String("error", err.Error()))
        return
    }
    switch event.Action {
    case "cancel":
        // call w.app.Command.CancelOrder.Handle(ctx, &req)
    }
}
```

### Register in server.go

```go
workerGroup.Register(
    event.NewOrderEventWorker(mqttCli, orderApplication, &cfg.Workers.OrderEvent),
)
```

### Config

```yaml
# config/config.yml
workers:
  orderEvent:
    topic: "events/order"
    qos: 1
```

```go
// config/config.go
type Workers struct {
    OrderEvent WorkerConfig
}
```

---

## 7. Infrastructure Layer

### Repository (`internal/infras/repository/<domain>/`)

Two files: interface + implementation.

**Interface** (owned by the domain, not infras):

```go
// internal/infras/repository/order/repository.go
package order

type Repository interface {
    FindByID(ctx context.Context, id uint) (*models.Order, error)
    Create(ctx context.Context, order *models.Order) error
}
```

**Implementation** uses `orm.ORM`:

```go
// internal/infras/repository/order/repos_impl.go
type orderRepository struct{ orm orm.ORM }

func NewOrderRepository(o orm.ORM) Repository { return &orderRepository{orm: o} }

func (r *orderRepository) FindByID(ctx context.Context, id uint) (*models.Order, error) {
    var order models.Order
    if err := r.orm.GormDB().WithContext(ctx).First(&order, id).Error; err != nil {
        return nil, err
    }
    return &order, nil
}
```

**Generate mock** for unit tests:

```bash
go generate ./internal/infras/repository/order/...
```

Add `//go:generate mockery --name=Repository` to `repository.go`.

### Outbound Adapter (`internal/infras/adapter/<service>/`)

Interface defined in the package, implementation wires `pkgs/caller`:

```go
// adapter.go — interface
type ServiceAdapter interface {
    GetData(ctx context.Context, req *dto.Req) (*dto.Resp, error)
}

// adapter_impl.go — implementation
func NewAdapter(cfg *config.HTTPClient) ServiceAdapter {
    c := caller.New(caller.Config{BaseURL: cfg.BaseURL, ...})
    return &adapter{caller: c}
}
```

### MQTT Publisher (`internal/infras/mqttpub/`)

Wraps `pkgs/mqtt.Client`. Inject `Publisher` into app commands that need to emit events.

```go
pub.Publish(ctx, "events/order", payload)  // topic, qos=1, retained=false
```

---

## 8. Shared Packages

These packages contain zero business logic. Do not modify unless fixing a bug.

| Package | Use when |
|---|---|
| `pkgs/ultis/svcerr` | Returning or wrapping errors from handlers/app |
| `pkgs/ultis/response` | Building HTTP responses: `res := response.SuccessRes; res.Data = dto; res.JSON(c)` |
| `pkgs/ultis/paging` | Paginated list endpoints: `ListQuery` (input), `ListResponse` (output) |
| `pkgs/ultis/mapper` | Copying structs or converting struct → map |
| `pkgs/ultis/validate` | Already wired via `e.Validator = validate.GetValidator()` — no direct use needed |
| `pkgs/decorator` | Generic `CommandHandler[T]` / `QueryHandler[T,E]` interfaces |
| `pkgs/caller` | Building outbound HTTP adapters |
| `pkgs/mqtt` | Direct MQTT pub/sub — used by `mqttpub` and workers, not by app layer |
| `pkgs/logger` | Only used in `server.go` to set the default `slog` handler |

---

## 9. Configuration

### Adding a config key

1. Add the field to the relevant struct in `config/config.go`.
2. Add the default value to `config/config.yml`.
3. Document the env override as a comment: `# override: SECTION_FIELD`.

Viper maps YAML keys to env vars with the rule: `.` → `_`, all uppercase.
Example: `adapters.product.apiKey` → `ADAPTERS_PRODUCT_APIKEY`.

### Sensitive fields

Never commit real credentials. Leave the YAML value empty and document the env var:

```yaml
database:
  password: ""   # override: DATABASE_PASSWORD
```

---

## 10. Error Handling

### Predefined errors (`pkgs/ultis/svcerr/common_err.go`)

Use existing errors before creating new ones:

| Error var | HTTP | Use case |
|---|---|---|
| `ErrBadRequest` | 400 | Bind/parse failure |
| `ErrInvalidParameters` | 400 | Malformed query params |
| `ErrInvalidIdParam` | 400 | Non-numeric path `/:id` |
| `ErrUnauthenticated` | 401 | Missing/invalid token |
| `ErrPermissionDenied` | 403 | Insufficient rights |
| `ErrNotFound` | 404 | Generic not found |
| `ErrRecordNotFound` | 404 | DB record missing |
| `ErrAlreadyExists` | 409 | Unique constraint violation |
| `ErrInternalServer` | 500 | Unexpected error — always log before returning |

### Adding a new error

```go
// pkgs/ultis/svcerr/common_err.go
ErrQuotaExceeded = &Error{
    Status:  402,
    Code:    enums.ResourceExhausted,
    Message: "Quota exceeded",
}
```

### Logging before returning 500

```go
slog.Error("create order failed", slog.String("error", err.Error()))
return nil, svcerr.ErrInternalServer
```

Never expose raw error messages to the caller. Log them, return a typed error.

---

## 11. Implementing a New Domain

### Step 1 — Define the domain types

```
internal/domain/models/<domain>.go          ← GORM model
internal/domain/dtos/<domain>dtos/          ← request/response types with validate tags
internal/domain/msgs/<domain>_event_msg.go  ← MQTT payload (only if worker needed)
```

### Step 2 — Define the repository interface

```
internal/infras/repository/<domain>/repository.go
```

Add `//go:generate mockery --name=Repository` and run `go generate` for test mocks.

### Step 3 — Implement the repository

```
internal/infras/repository/<domain>/repos_impl.go
```

### Step 4 — Write application commands and queries

```
internal/app/<domain>/command/<action>.go   ← one file per command
internal/app/<domain>/query/<action>.go     ← one file per query
internal/app/<domain>/application.go        ← aggregates all commands/queries
```

### Step 5 — Write the HTTP handler

```
internal/handler/<domain>/handler.go
```

### Step 6 — Register routes

```go
// internal/router/routes.go
domainRoute := v1.Group("/<domain>s")
domainRoute.POST("", domainHandler.Create<Domain>)
```

### Step 7 — Wire in server.go

```go
// internal/server.go — inside NewServer()
<domain>Repo        := <domain>repos.New<Domain>Repository(database)
<domain>Application := <domain>app.New<Domain>Application(<domain>Repo)
<domain>Handle      := <domain>handler.New<Domain>Handler(<domain>Application)
router.RegisterRoutes(echoApp, ..., <domain>Handle)
```

### Step 8 — (Optional) Add a worker

```
internal/worker/event/<domain>_event_worker.go
```

Add config to `config/config.go` (`Workers.<Domain>Event WorkerConfig`) and `config/config.yml`.

Register in server.go:

```go
workerGroup.Register(
    event.New<Domain>EventWorker(mqttCli, <domain>Application, &cfg.Workers.<Domain>Event),
)
```

### Step 9 — (Optional) Add an outbound adapter

```
internal/infras/adapter/<service>/adapter.go
internal/infras/adapter/<service>/adapter_impl.go
internal/infras/adapter/<service>/dto/
```

Add HTTP client config to `config/config.go` (`Adapters.<Service> HTTPClient`) and YAML.

---

## 12. Checklist — New Feature End-to-End

Use this before opening a PR for a new domain or feature.

**Domain:**
- [ ] Model created and GORM-tagged
- [ ] DTOs created — validate tags on all user-facing fields
- [ ] `NewXxxDTO(model)` constructor handles nil input

**Application:**
- [ ] Command/Query interface declared with `decorator.*Handler`
- [ ] Business logic in `Handle()` — no nil/zero re-checks (handler validated already)
- [ ] Infra errors mapped to `svcerr.*` — raw errors never returned
- [ ] Unexpected errors logged before returning `svcerr.ErrInternalServer`

**Handler:**
- [ ] `c.Bind` → on error return `svcerr.ErrBadRequest`
- [ ] `c.Validate` → on error return `err` as-is
- [ ] Success path: `res := response.SuccessRes; res.Data = dto; return res.JSON(c)`
- [ ] Route registered in `internal/router/routes.go`

**Worker (if applicable):**
- [ ] Implements `worker.Worker` interface (`Start(ctx) error`)
- [ ] Handler dispatched via `go w.handle(ctx, payload)` — MQTT loop never blocked
- [ ] Topic and QoS read from `config.WorkerConfig`, not hardcoded
- [ ] Registered in `server.go` via `workerGroup.Register(...)`

**Infrastructure:**
- [ ] Repository interface in `repository/<domain>/repository.go`
- [ ] Mock generated (`go generate`) for unit tests
- [ ] No validation in repository — trusts the app layer

**Config:**
- [ ] New keys added to `config/config.go` struct
- [ ] Defaults set in `config/config.yml`
- [ ] Sensitive fields empty in YAML with `# override: ENV_VAR` comment

**Tests:**
- [ ] App layer: unit tests with mock repository
- [ ] Handler layer: unit tests with fake repo + `newJSONContext`
- [ ] Repository layer: integration tests with in-process SQLite

---

## Replacing the Sample Domain

The `user` domain exists only as a reference implementation. To start fresh:

**Delete:**
```
internal/domain/models/user.go
internal/domain/dtos/userdtos/
internal/domain/msgs/user_event_msg.go
internal/app/user/
internal/handler/user/
internal/infras/repository/user/
internal/infras/adapter/productsvc/
internal/worker/event/user_event_worker.go
```

**Keep (framework core — do not delete):**
```
pkgs/                       ← all shared packages
config/                     ← modify, do not delete
internal/server.go          ← rewire with your domain
internal/http.go            ← do not modify
internal/router/routes.go   ← replace routes
internal/worker/worker_group.go
internal/infras/mqttpub/
```

After deleting, follow the steps in [Section 11](#11-implementing-a-new-domain) for your domain.
