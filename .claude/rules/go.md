---
paths: "**/*.go"
version: "1.0.0"
---
<!-- keel:generated -->

# Go

Rules for writing idiomatic, production-grade Go code.

## Error Handling

- Always check errors. Never use `_` to discard an error return.
- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`.
- Use `errors.Is()` and `errors.As()` for error comparison, not string matching.
- Define sentinel errors or custom error types for errors callers need to handle differently.
- Return errors, don't panic. Reserve `panic` for truly unrecoverable programmer errors.

```go
// BAD
result, _ := doSomething()

// BAD — no context
if err != nil { return err }

// GOOD
result, err := doSomething()
if err != nil {
    return fmt.Errorf("processing order %s: %w", orderID, err)
}
```

## Naming

- Follow Go conventions: `MixedCaps` for exported, `mixedCaps` for unexported. No underscores in names.
- Receivers: short (1-2 letters), consistent across methods of the same type. `(o *Order)` not `(order *Order)`.
- Interfaces: single-method interfaces use the method name + "er" suffix (`Reader`, `Writer`, `Closer`).
- Packages: short, lowercase, singular. `order` not `orders`, `user` not `userService`.
- Avoid stutter: `order.Order` is fine, `order.OrderService` stutters.

## Package Design

- Each package has a single, clear purpose described by its name.
- No `utils`, `helpers`, `common`, or `base` packages.
- Avoid circular imports — if two packages need each other, extract the shared concept into a third.
- Keep the public API surface small. Only export what other packages need.
- `internal/` for packages that should not be imported outside the module.

## Interfaces

- Define interfaces where they are USED, not where they are implemented.
- Keep interfaces small — one or two methods. Compose larger interfaces from smaller ones.
- Accept interfaces, return concrete types.
- Don't define an interface until you have two or more implementations, or need it for testing.

```go
// BAD — interface defined next to implementation, too large
type UserStore interface {
    Create(u *User) error
    Update(u *User) error
    Delete(id string) error
    FindByID(id string) (*User, error)
    FindByEmail(email string) (*User, error)
    List(opts ListOpts) ([]*User, error)
}

// GOOD — small interface, defined where used
type UserFinder interface {
    FindByID(id string) (*User, error)
}
```

## Concurrency

- Don't start goroutines without a plan for how they stop. Use `context.Context` for cancellation.
- Always use `sync.WaitGroup` or channels to wait for goroutines to complete.
- Never share memory between goroutines without synchronization. Prefer channels for communication.
- Use `errgroup.Group` for running multiple operations concurrently and collecting errors.
- If a function starts goroutines, document it. Callers need to know about concurrent behavior.

## Structs & Methods

- Use pointer receivers when the method modifies the receiver, when the struct is large, or for consistency across the type's method set.
- Use value receivers for small, immutable types.
- Don't mix pointer and value receivers on the same type.
- Initialize structs with field names: `User{Name: "alice", Age: 30}` not `User{"alice", 30}`.

## Context

- `context.Context` is always the first parameter. Never store it in a struct.
- Propagate context through the call chain. Don't create new contexts in the middle.
- Use context for cancellation and deadlines, not for passing business data (with rare exceptions like request IDs and auth tokens in middleware).

## Testing

- Use table-driven tests for functions with multiple input/output combinations.
- Use `testify/assert` or `testify/require` for readable assertions.
- Name test cases descriptively: `"returns error when order not found"`.
- Use `t.Helper()` in test helper functions so failure line numbers point to the caller.
- Test files live in the same package for white-box tests, or `_test` package for black-box tests.

## Project Layout

Follow the standard Go project layout conventions:
- `cmd/` — main applications (one dir per binary)
- `internal/` — private packages
- `pkg/` — public libraries (only if building a reusable library)
- Don't create `pkg/` unless the code is genuinely intended for external use.
