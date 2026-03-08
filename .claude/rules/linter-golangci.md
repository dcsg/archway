---
paths: "**/*.go"
version: "1.0.0"
---
<!-- keel:generated — synced from .golangci.yml -->

# Go Linter Rules (golangci-lint)

These rules reinforce what golangci-lint enforces, so violations are caught before linting.

## Enabled Linters

The following linters are active — write code that passes all of them:

- **errcheck**: Always check returned errors. Never discard with `_`.
- **govet**: Follow `go vet` rules — correct printf format strings, struct tags, etc.
- **ineffassign**: Don't assign to variables that are never read afterwards.
- **staticcheck**: Follow staticcheck rules. Note: `strings.Title` is excluded (deprecated but allowed).
- **unused**: Don't leave unused variables, functions, types, or constants.
- **gosimple**: Use simpler constructs when available (e.g., `strings.Contains` over manual loops).
- **gocritic**: Follow gocritic suggestions — use consistent style, avoid anti-patterns. Relaxed in test files.
- **misspell**: No typos in comments, strings, or identifiers.
- **gofmt**: Code must be `gofmt`-formatted. No manual formatting.
- **goimports**: Imports must be grouped and sorted (stdlib, external, internal).
- **bodyclose**: Always close HTTP response bodies (`defer resp.Body.Close()`).
- **noctx**: Always pass `context.Context` to HTTP requests — use `http.NewRequestWithContext`.
- **prealloc**: Pre-allocate slices when the size is known (`make([]T, 0, n)`).

## Test File Exceptions

In `_test.go` files, the following are relaxed:
- **errcheck**: Unchecked errors are acceptable in tests.
- **gocritic**: Style suggestions are not enforced in tests.
