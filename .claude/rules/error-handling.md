---
paths: "**/*"
version: "1.0.0"
---
<!-- keel:generated -->

# Error Handling

Rules for handling errors consistently and safely.

## Never Silently Swallow Errors

Every catch/except/recover block must do ONE of:
1. Handle the error (retry, fallback, user-facing message)
2. Re-throw/propagate with added context
3. Log with sufficient context for debugging

Empty catch blocks are never acceptable. If you genuinely need to ignore an error, add a comment explaining why.

```
// BAD
try { sendEmail(user) } catch (e) {}

// BAD — logs but loses context
try { sendEmail(user) } catch (e) { console.log("error") }

// GOOD — handles with context
try {
  sendEmail(user)
} catch (e) {
  logger.warn("failed to send welcome email", { userId: user.id, error: e })
  // non-critical, continue without email
}
```

## Add Context When Propagating

When re-throwing or wrapping errors, add context about what operation was being attempted. The goal: someone reading the error in logs can trace what happened without looking at code.

```
// BAD — original error with no context
throw err

// GOOD — wrapped with operation context
throw new Error(`failed to process order ${orderId}: ${err.message}`, { cause: err })
```

## Use Typed/Structured Errors

- Define specific error types for different failure categories (ValidationError, NotFoundError, AuthorizationError, etc.).
- Don't throw generic Error/Exception with string messages for known failure modes.
- Error types enable callers to handle different failures differently.

```
// BAD
throw new Error("not found")
throw new Error("invalid email")

// GOOD
throw new NotFoundError("order", orderId)
throw new ValidationError("email", "invalid format")
```

## Validate at System Boundaries Only

- Validate at entry points: HTTP handlers, CLI parsers, message consumers, public API methods.
- Trust internal code. If a private function receives data that already passed validation, don't re-validate.
- Don't add defensive checks for scenarios that can't happen given the code path.

## Fail Fast

- Detect errors as early as possible. Don't let invalid state propagate through multiple layers before failing.
- Check preconditions at function entry. Return/throw immediately if they're not met.
- Prefer returning errors over panic/crash for recoverable situations. Reserve panics for truly unrecoverable states (corrupted data, violated invariants).

## Error Responses

- External-facing errors (API responses, user messages): clear, actionable, no internal details.
- Internal errors (logs, monitoring): detailed, with full context, stack traces, and correlation IDs.
- Never expose: stack traces, SQL queries, internal file paths, or configuration details to external clients.

```
// To the client
{ "error": "Order not found", "code": "ORDER_NOT_FOUND" }

// To the logs
{ "error": "order not found", "orderId": "abc-123", "userId": "user-456", "stack": "..." }
```

## Result Types (Where Available)

In languages that support it (Go, Rust, TypeScript with libraries), prefer Result/Either types over throwing exceptions for expected failure modes:

- Exceptions for truly exceptional situations (out of memory, disk full, programming errors)
- Result types for expected failures (not found, validation failed, permission denied)

This makes error handling explicit in the type signature — callers can't accidentally ignore errors.
