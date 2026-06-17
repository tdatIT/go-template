---
name: go-clean-cqrs-architecture
description: Use when designing or implementing a Go HTTP service with clean architecture + CQRS, matching this repo's template (Echo v5, GORM, slog, generic decorator handlers, svcerr error model). Covers layer layout, where interfaces live, command/query handlers, repositories, DTOs, transport wiring, structured logging, and the service-error pattern. Use it to scaffold a new domain or a whole new service that looks like this base.
---

# Clean Architecture with CQRS in Go

This skill describes the exact conventions used in this template so an agent can
reproduce the same structure for a new domain or a new service. Follow the
patterns below verbatim — names, layering, and error/logging choices are
intentional. For tests and mocks see
[Writing Tests (Go)](../writing-tests-go-projects/SKILL.md); for logging detail
see [Handle Logging (Go)](../handle-logging/SKILL.md).

## Stack

- HTTP: **Echo v5** (`github.com/labstack/echo/v5`), handlers take `*echo.Context`.
- Persistence: **GORM** (Postgres) behind an `orm.ORM` interface.
- Cache: **go-redis** behind a `rdclient.RedisClient` interface.
- Logging: **stdlib `log/slog`** (JSON handler) — NOT zap/zerolog.
- Config: **viper** with `CONFIG_PATH` override.
- Validation: **go-playground/validator** via Echo's validator hook.

## Layers

- **Domain** (`internal/domain`): entities (`models`), request/response DTOs
  (`dtos`), enums/codes (`enums`). Framework-agnostic.
- **Application** (`internal/app/<domain>`): use-case orchestration only —
  command/query handlers. No DB, HTTP, or transport details here.
- **Infrastructure** (`internal/infras`): concrete implementations
  (repositories, adapters). Defines the repository **interface** next to its
  implementation and depends on `pkgs/db`.
- **Interface/Transport** (`internal/handler`, `internal/router`): Echo handlers
  and routes; depend on Application.
- **Shared** (`pkgs`): decorator, logger, db clients, utils (`ultis`).
- **Entry/Config** (`cmd`, `config`, `internal/server.go`, `internal/http.go`).

Group code by domain (e.g. `user`) across layers.

## Project Layout

```
service/
├─ cmd/main.go                         # builds server.NewServer(), starts Echo
├─ config/{config.go,config.yml}       # viper; AppConfig struct
├─ internal/
│  ├─ server.go                        # wiring: config→slog→DB/Redis→repo→app→handler→routes
│  ├─ http.go                          # Echo middleware, validator, global error handler
│  ├─ router/routes.go                 # RegisterRoutes(e, handlers...)
│  ├─ app/<domain>/
│  │  ├─ application.go                # aggregates Command/Query, NewXApplication(repo)
│  │  ├─ command/{create,update,delete}_<domain>.go
│  │  └─ query/{get_*,list_*}.go
│  ├─ domain/
│  │  ├─ models/<domain>.go            # GORM models (gorm tags)
│  │  ├─ dtos/<domain>dtos/<domain>_dto.go  # req/res DTOs (validate tags) + NewXDTO mappers
│  │  └─ enums/svc_code.go             # numeric service codes
│  ├─ handler/<domain>/handler.go      # Echo handlers (*echo.Context)
│  └─ infras/repository/<domain>/
│     ├─ repository.go                 # the Repository INTERFACE
│     └─ repos_impl.go                 # GORM implementation + NewXRepository(orm.ORM)
├─ pkgs/
│  ├─ decorator/{command.go,queries.go}# generic CommandHandler/QueryHandler
│  ├─ logger/slog.go                   # NewJsonSlogHandler
│  ├─ db/{orm,rdclient}/               # ORM + Redis interfaces & constructors
│  └─ ultis/{response,svcerr,validate}/
├─ .mockery.yaml                       # see Writing Tests skill
├─ Dockerfile
└─ go.mod
```

## Where Interfaces Live (important)

- The **repository interface** is defined in the infrastructure package next to
  its implementation: `internal/infras/repository/<domain>/repository.go`. The
  application layer imports this interface; the implementation
  (`repos_impl.go`) is wired in `internal/server.go`.
- Low-level infra interfaces (`orm.ORM`, `rdclient.RedisClient`) live in
  `pkgs/db/*` next to their constructors.

```go
// internal/infras/repository/user/repository.go
package user

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindAndCount(ctx context.Context, limit, offset int) ([]*models.User, int64, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
}
```

```go
// internal/infras/repository/user/repos_impl.go
type userReposImpl struct{ db orm.ORM }

func NewUserRepository(database orm.ORM) Repository { return &userReposImpl{db: database} }

func (r *userReposImpl) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.GormDB().WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err // return raw gorm error; the app layer maps it to svcerr
	}
	return &user, nil
}
```

The repository returns **raw GORM errors** (e.g. `gorm.ErrRecordNotFound`,
`gorm.ErrDuplicatedKey`). Mapping to service errors happens in the app layer.

## CQRS via Generic Decorators

`pkgs/decorator` defines three generic handler interfaces:

```go
type CommandHandler[T any]              interface{ Handle(ctx context.Context, req T) error }
type CommandReturnHandler[T any, E any] interface{ Handle(ctx context.Context, req T) (E, error) }
type QueryHandler[T any, E any]         interface{ Handle(ctx context.Context, req T) (E, error) }
```

Each use case declares a typed alias `I<Name>` over one of these, a private
struct holding only the repository, a `New<Name>` constructor returning the
interface, and a `Handle` method:

```go
// internal/app/user/command/create_user.go
type ICreateUserCommand decorator.CommandReturnHandler[*userdtos.CreateUserReq, *userdtos.UserDTO]

type createUserCommand struct{ repo user.Repository }

func NewCreateUserCommand(repo user.Repository) ICreateUserCommand { return &createUserCommand{repo: repo} }

func (c *createUserCommand) Handle(ctx context.Context, req *userdtos.CreateUserReq) (*userdtos.UserDTO, error) {
	model := &models.User{Name: strings.TrimSpace(req.Name), Email: strings.TrimSpace(req.Email)}
	if model.Name == "" || model.Email == "" {
		return nil, svcerr.ErrInvalidParameters
	}
	if err := c.repo.Create(ctx, model); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, svcerr.ErrAlreadyExists
		}
		slog.Error("create user failed", slog.String("error", err.Error()))
		return nil, svcerr.ErrInternalServer
	}
	return userdtos.NewUserDTO(model), nil
}
```

- Commands with no payload back to the caller use `CommandHandler[T]` (delete).
- Commands that return data use `CommandReturnHandler[T, E]` (create/update).
- Reads use `QueryHandler[T, E]`.

Aggregate handlers in `application.go` and inject the repository once:

```go
// internal/app/user/application.go
type Application struct {
	Command commands
	Query   queries
}

func NewUserApplication(userRepo user.Repository) *Application {
	return &Application{
		Command: commands{
			CreateUser: command.NewCreateUserCommand(userRepo),
			UpdateUser: command.NewUpdateUserCommand(userRepo),
			DeleteUser: command.NewDeleteUserCommand(userRepo),
		},
		Query: queries{
			GetUser:   query.NewGetUserByIDQuery(userRepo),
			ListUsers: query.NewListUsersQuery(userRepo),
		},
	}
}
```

## DTOs & Models

- DTOs live in `internal/domain/dtos/<domain>dtos`, carry `json` + `validate`
  tags, and provide `NewXDTO`/`NewXDTOs` mappers from models. Path params use
  `json:"-"` and are set by the handler.
- Models live in `internal/domain/models` with GORM tags.

```go
type CreateUserReq struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email,max=255"`
}
func NewUserDTO(u *models.User) *UserDTO { if u == nil { return nil }; return &UserDTO{ID: u.ID, /* … */} }
```

## Transport (Echo v5 handlers)

Handlers bind+validate, delegate to the app, and return service errors directly
(the global error handler converts them). Success goes through `response`:

```go
func (h *Handler) CreateUser(c *echo.Context) error {
	var req userdtos.CreateUserReq
	if err := c.Bind(&req); err != nil {
		slog.Error("bind request failed", slog.String("error", err.Error()))
		return svcerr.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		slog.Error("validate request failed", slog.String("error", err.Error()))
		return err
	}
	data, err := h.app.Command.CreateUser.Handle(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	success := response.SuccessRes
	success.Data = data
	return success.JSON(c)
}
```

- The `Handler` struct holds `*<domain>.Application`; `NewXHandler(app)`.
- Parse path params with a small helper returning `svcerr.ErrInvalidIdParam`.

## Errors: the `svcerr` model

- All application/transport errors are `*svcerr.Error` (`pkgs/ultis/svcerr`).
- Use the predefined singletons in `common_err.go`
  (`ErrBadRequest`, `ErrInvalidParameters`, `ErrAlreadyExists`,
  `ErrRecordNotFound`, `ErrInternalServer`, …). Compare with `errors.Is` — they
  are pointer singletons.
- Map infra/GORM errors to a service error at the **app layer**, then log the
  raw cause with slog before returning a generic `ErrInternalServer`.
- The Echo global error handler (`svcerr.ErrorHandlerEchoFn`, registered in
  `internal/http.go`) renders any `*svcerr.Error` into the `response.BaseRes`
  shape `{code, message, data}` with the right HTTP status.
- Add new errors to `common_err.go`; never return raw infra errors from
  handlers.

```go
// shape of a service error
type Error struct {
	Status               int
	InternalErrorMessage string
	Code                 int    `json:"code"`
	Message              string `json:"message"`
}
```

## Logging

Use stdlib `slog` with typed attributes; never `fmt`-format log messages:

```go
slog.Error("get user failed",
	slog.Uint64("user_id", uint64(req.ID)),
	slog.String("error", err.Error()))
```

The JSON handler is built in `pkgs/logger.NewJsonSlogHandler` and installed as
the default logger in `internal/server.go`. Request logging uses Echo's
`RequestLoggerWithConfig` emitting `slog.Attr` values.

## Wiring order (`internal/server.go`)

`load config → set slog default → open ORM + Redis → NewXRepository(orm) →
NewXApplication(repo) → NewXHandler(app) → router.RegisterRoutes`. Close DB/Redis
on shutdown.

## Checklist: add a new domain `foo`

1. `internal/domain/models/foo.go` (+ AutoMigrate entry in `pkgs/db/orm`).
2. `internal/domain/dtos/foodtos/foo_dto.go` (req/res + mappers).
3. `internal/infras/repository/foo/{repository.go (interface), repos_impl.go}`.
4. `internal/app/foo/{application.go, command/*, query/*}` using the decorator aliases.
5. `internal/handler/foo/handler.go` (+ `NewFooHandler`).
6. Register routes in `internal/router/routes.go` and wire everything in `internal/server.go`.
7. Add any new errors to `pkgs/ultis/svcerr/common_err.go`.
8. Generate mocks + write black-box tests — see [Writing Tests (Go)](../writing-tests-go-projects/SKILL.md).
