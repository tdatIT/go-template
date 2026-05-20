# go-template

Simple Go service template using Echo v5, Viper config, GORM/Postgres, Redis, and structured JSON logging.

## Requirements

- Go 1.26+
- (Optional) PostgreSQL and Redis for the backing services

## Configuration

Default config file: `config/config.yml`  
Override path: set `CONFIG_PATH` to a custom file.

## Run

```sh
make run
# or
go run ./cmd/main.go
```

## Build

```sh
make build
```

## Test

```sh
make test
```

## Lint and format

```sh
make lint
make fmt
make tidy
```

With golangci-lint installed, you can run:

```sh
golangci-lint run ./...
```

## Docker

```sh
make docker-build
```
