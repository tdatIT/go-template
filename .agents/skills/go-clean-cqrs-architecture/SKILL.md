---
name: go-clean-cqrs-architecture
description: Use when designing and implementing a clean architecture with CQRS in Go. This includes structuring your code into layers (e.g., domain, app, infrastructure), defining clear boundaries between read and write models, and ensuring that your code is modular, testable, and maintainable. Also use when discussing best practices for organizing code in a way that promotes separation of concerns and scalability in Go applications.
---

# Clean Architecture with CQRS in Go
This skill covers designing and implementing a clean architecture with Command Query Responsibility Segregation (CQRS) in Go projects. It includes structuring code into layers (domain, application, infrastructure)
# Layers
- Domain: core entities, rules — framework-agnostic.
- Application: use-case orchestration, interfaces only.
- Infrastructure: concrete implementations (DB, cache, external APIs).
- Interface: HTTP/gRPC/CLI handlers — depend on Application.
- Shared/Cmd/Config: utilities, entry points, and config.

Group code by domain (e.g., `user`) across layers to avoid cross-cutting confusion.

# Minimal Project Layout
```
service/
├─ cmd/main.go
├─ config/
├─ internal/
│  ├─ app/<domain>/{application.go,command/,query/}
│  ├─ domain/{models,dto,errors}
│  ├─ handler/<domain>/
│  └─ infras/{repository,adapter}
├─ pkgs/{decorator,logger,utils}
├─ Dockerfile
└─ go.mod
```

# Interfaces and Boundaries
- Define interfaces in Application; implement them in Infrastructure.
- Keep Application free of infrastructure and transport details.

# CQRS via Decorators
- Use `decorator.CommandHandler` and `decorator.QueryHandler` to separate write/read handlers.
- See `pkgs/decorator/command.go` and `pkgs/decorator/queries.go` for interfaces.

Example:
Define a typed query handler in `internal/app/user/query`:

```go
type IGetUserByIDQuery decorator.QueryHandler[*userdto.GetUserByIDReq, *userdto.UserDTO]

type getUserByIDQuery struct { userRepo repository.UserRepository }

func (q *getUserByIDQuery) Handle(ctx context.Context, req *userdto.GetUserByIDReq) (*userdto.UserDTO, error) {
	// implement business logic using userRepo
}
```

Initialize app services in `internal/app/user/application.go` and inject them into handlers.

# Application Scope
- Application: business logic and orchestration only.
- Do not access DB, external services, or HTTP directly here.

# Logging & Errors (short)
- Use structured logging (e.g., Zap) with contextual fields.
- Wrap lower-level errors with domain error types before returning.

```go
if err := repo.Delete(ctx, id); err != nil {
	logger.Error("delete failed", zap.Error(err))
	return domain.ErrDeleteFailed.WithError(err)
}
```

# Related Skills

- Refer to related skills for details: [Handle Logging (Go)](../handle-logging/SKILL.md) and [Writing Tests (Go)](../writing-tests-go-projects/SKILL.md).

- Refer to related skills for unit testing and mocking best practices in Go projects: [Writing Tests for Go Projects](../writing-tests-go-projects/SKILL.md).


