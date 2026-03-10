---
paths: "**/*.go"
version: "1.0.0"
source: linter
linter: golangci-lint
generated-by: keel:sync
---
<!-- keel:generated — synced from .golangci.yml -->

# Go Linter Rules (golangci-lint)

Rules translated from `.golangci.yml` — these mirror your linter config so Claude produces compliant code.

## Formatters

The following formatters are active — code must pass them:

- **gofmt**: Code must be `gofmt`-formatted. No manual formatting. Struct field tags must be aligned.
- **goimports**: Imports must be grouped and sorted (stdlib, external, internal). Separate groups with a blank line.

## Linters

The following linters are active — write code that passes all of them:

- **errcheck**: Always check returned errors. Never discard with `_`.
- **govet**: Follow `go vet` rules — correct printf format strings, struct tags, etc.
- **ineffassign**: Don't assign to variables that are never read afterwards.
- **staticcheck**: Follow staticcheck rules (includes gosimple). Note: `strings.Title` is excluded (deprecated but allowed).
- **unused**: Don't leave unused variables, functions, types, or constants.
- **gocritic**: Follow gocritic suggestions — use switch statements instead of if-else chains, avoid anti-patterns. Relaxed in test files.
- **misspell**: No typos in comments, strings, or identifiers.
- **bodyclose**: Always close HTTP response bodies (`defer resp.Body.Close()`).
- **noctx**: Always pass `context.Context` to HTTP requests — use `http.NewRequestWithContext`.
- **prealloc**: Pre-allocate slices when the size is known (`make([]T, 0, n)`).

## Test File Exceptions

In `_test.go` files, the following are relaxed:
- **errcheck**: Unchecked errors are acceptable in tests.
- **gocritic**: Style suggestions are not enforced in tests.

## Build Hygiene

- Always run `go mod tidy` after adding or removing imports. CI will fail if go.mod is out of sync.
