# Copilot instructions

## Build, test, lint
```sh
make run
make build
make test
go test ./path/to/pkg -run TestName
make lint
make fmt
make tidy
make docker-build
```

## High-level architecture
The entry point is `cmd/main.go`, which constructs `internal.Server`, loads config via Viper from `config/config.yml` (override with `CONFIG_PATH`), sets the JSON `slog` handler, and starts the Echo v5 HTTP server. HTTP setup lives in `internal/http.go` (middleware, validator, and error handler). Cross‑cutting utilities are in `pkgs/` (logging, Redis, GORM/Postgres, validation, response/error helpers), while shared error codes are in `internal/domain/enums`.

## Key conventions
1. Use `pkgs/ultis/svcerr.Error` (or the predefined errors in `pkgs/ultis/svcerr/common_err.go`) for application errors so the Echo error handler can map them to `response.BaseRes` with `internal/domain/enums` codes.
2. Request validation relies on `validate.GetValidator()` as the Echo validator; use `validator` struct tags for request DTOs.
3. Logging is structured JSON via `logger.NewJsonSlogHandler`, which injects `service_name`; prefer `slog` for logs.
