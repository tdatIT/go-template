# AGENTS — Quick Guide for AI coding agents

Purpose: give an AI (or new developer) the minimum, concrete repository knowledge needed to be productive: where to look first, how things are wired, developer commands, and project-specific conventions to follow.

1) Big picture (read first)
 - Entry: `cmd/main.go` creates `server.NewServer()` and starts Echo.
 - Server wiring: `internal/server.go` builds the app: loads config, sets slog handler, opens DB/Redis, constructs repositories → application → handlers and calls `router.RegisterRoutes`.
 - HTTP layer: `internal/http.go` configures Echo middleware, request logging, validator and the global Echo error handler (`pkgs/ultis/svcerr.ErrorHandlerEchoFn`).
 - Router → handlers → app → repository: routes are defined in `internal/router/routes.go`, handlers live in `internal/handler/*`, application logic in `internal/app/*`, and persistence in `internal/infras/repository/*`.

2) Commands & developer workflows (exact)
 - Run locally: `make run` or `go run ./cmd/main.go` (README/Makefile)
 - Build: `make build` (binary in `build/` by default)
 - Test: `make test` (runs `go test ./...`); run single test: `go test ./pkg/path -run TestName`
 - Lint / fmt / tidy: `make lint` (runs `go vet`), `make fmt`, `make tidy` (`go mod tidy`)
 - Docker image: `make docker-build` (runs `docker build -t go-service:local .`)
 - Config override: set `CONFIG_PATH` env var to point to your config (default is `config/config.yml` per `config.NewConfig()` / `getDefaultConfig()`).

3) Project-specific conventions (must-follow)
 - Error handling: use `pkgs/ultis/svcerr.Error` (or predefined errors in `pkgs/ultis/svcerr/common_err.go`) for application errors. Handlers return these errors directly so `ErrorHandlerEchoFn` maps them into the `pkgs/ultis/response.BaseRes` shape.
 - Example: handler returns `svcerr.ErrBadRequest` on bind failures (`internal/handler/user/handler.go`).
 - Responses: use `response.SuccessRes` / `response.ErrorRes` and call `.JSON(c)` to send consistent payloads (`pkgs/ultis/response/api_res_structure.go`).
 - Validation: Echo validator set in `internal/http.go` uses `validate.GetValidator()`; request DTOs use `validator` tags (see `internal/domain/dtos/userdtos/user_dto.go`).
 - Logging: structured JSON `slog` is configured in `internal/server.go` via `pkgs/logger.NewJsonSlogHandler`. Request logging is done with Echo's `RequestLoggerWithConfig` and a custom `LogValuesFunc` that emits `slog.Attr`.

4) Integrations & wiring patterns
 - Postgres/GORM: created via `pkgs/db/orm.NewDBConnection(cfg)` in `internal/server.go`; repositories accept an `orm.ORM` instance (see `internal/infras/repository/user/repository.go`).
 - Redis: created via `pkgs/db/rdclient.NewRedisClient(cfg)` and closed on shutdown.
 - Config: viper is used in `config/config.go`. Environment overrides enabled via `v.AutomaticEnv()` and `CONFIG_PATH` controls the file to load.
 - Docker: `Dockerfile` and `docker-compose.yml` present for container-based integration.

5) Quick triage checklist (when changing behaviour)
 - Add route: update `internal/router/routes.go` and add handler + app + repo wiring in `internal/server.go`.
 - Add config key: update `config/config.yml` and check `config.AppConfig` fields in `config/config.go` (viper uses env key replacer `.`→`_`).
 - New errors: add to `pkgs/ultis/svcerr/common_err.go` and return them from handlers/apps so the Echo error handler maps them automatically.

6) Known quirks & gotchas (discoverable)
 - `config.getDefaultConfig()` returns `"/config/config"` and `NewConfig()` calls `v.SetConfigName(path)` — be careful with viper config name vs path (use `CONFIG_PATH` if unsure).
 - `pkgs/ultis/svcerr/handle_err_fn.go` inspects `SERVER_DEBUG` to decide how verbose validation messages are.
 - Small code smells to watch for: `cmd/main.go` uses `sync.WaitGroup` with `wg.Go(...)` which is non-standard — tests/CI may surface issues.

7) Where to look next (first files to open)
 - `cmd/main.go`, `internal/server.go`, `internal/http.go`, `config/config.go`, `config/config.yml`, `internal/router/routes.go`, `internal/handler/*`, `internal/app/*`, `pkgs/ultis/svcerr/*`, `pkgs/ultis/response/*`, `pkgs/logger/slog.go`, `pkgs/db/*`.

If you need a follow-up, I can: (A) produce a checklist PR template for new routes/features, (B) extract a dependency diagram, or (C) add unit-test scaffolding examples.

