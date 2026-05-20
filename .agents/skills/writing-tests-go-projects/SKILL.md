---
name: writing-tests-go-projects

description: "Writes and fixes Go unit tests with table-driven tests, testify assertions, mocks, and coverage checks. Use when the user asks to run go test, add tests, debug failing tests, or improve test coverage in Go projects."
---


# Write Tests for Go Projects

Run tests, write new tests, fix failures, and improve coverage in Go projects. Read ./resources/mockup.md for mocking with Mockery and ./resources/test-table.md for table-driven tests.

## Workflow

1. Run existing tests first: `go test ./...`
2. Add or update tests in `*_test.go` files.
3. Prefer same-package tests for internal behavior; use external `_test` package for public API behavior.
4. Use table-driven tests when only inputs/outputs vary.
5. Use mocks/fakes for external dependencies to keep tests deterministic.
6. Re-run tests and coverage:
  ```
  go test -coverprofile=coverage.out ./...
  go tool cover -func=coverage.out
  ```

## Rules

- Use `require` for assertions that should stop a test immediately.
- Use `assert` for additional checks after required conditions pass.
- Do not add exported production APIs only to make testing easier.
- For DB tests, prefer lightweight isolated setups (for example, sqlite in-memory or test containers based on project context).
- For generated mocks, keep them in `mocks/` in root project and ignore them where appropriate for your repo workflow.


## Resources
- [Test Table Example](./resources/test-table.md)
- [Mocking with Mockery](./resources/mockup.md)