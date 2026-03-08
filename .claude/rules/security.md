---
paths: "**/*"
version: "1.0.0"
---
<!-- keel:generated -->

# Security

Rules for writing secure code. Security is not optional — these rules apply to every change.

## Input Validation

- Validate ALL external input at system boundaries: HTTP handlers, CLI arguments, file parsers, message consumers, webhook handlers.
- Never trust input from: users, external APIs, file uploads, URL parameters, headers, cookies.
- Internal function calls between trusted modules do NOT need re-validation — validate once at the boundary.
- Reject invalid input early with clear error messages. Don't sanitize and silently continue.

## SQL & Database

- Use parameterized queries EXCLUSIVELY. No string interpolation, concatenation, or template literals in SQL.
- Use your ORM/query builder's parameter binding. If writing raw SQL, use placeholders (`$1`, `?`, `:name`).

```
// BAD — SQL injection vulnerability
query("SELECT * FROM users WHERE id = " + userId)
query(`SELECT * FROM users WHERE id = ${userId}`)

// GOOD
query("SELECT * FROM users WHERE id = $1", userId)
```

## Secrets & Sensitive Data

- NEVER log: passwords, tokens, API keys, credit card numbers, SSNs, personal health information.
- NEVER commit: `.env` files, credential files, private keys, service account JSON.
- NEVER hardcode: secrets, API keys, connection strings, passwords in source code.
- Use environment variables or secret management services for all credentials.
- If you suspect a secret was logged or committed, treat it as compromised — rotate immediately.

## Authentication & Authorization

- Check authorization BEFORE accessing or modifying any resource.
- Never rely on client-side checks alone — always enforce on the server.
- Use the principle of least privilege: grant minimum permissions needed.
- Validate that the authenticated user has permission for the SPECIFIC resource, not just the resource type.

```
// BAD — checks if user is authenticated, not if they own this resource
if (user.isAuthenticated) { return getOrder(orderId) }

// GOOD — checks ownership
if (user.isAuthenticated && order.userId === user.id) { return order }
```

## HTTP & API Security

- Set appropriate security headers (CORS, CSP, HSTS, X-Content-Type-Options).
- Use HTTPS for all external communication.
- Rate-limit authentication endpoints and expensive operations.
- Don't expose internal error details (stack traces, SQL errors) in API responses to clients.
- Validate Content-Type on incoming requests.

## File Handling

- Validate file types by content (magic bytes), not just extension.
- Set maximum file size limits.
- Never use user-provided filenames directly in file paths — sanitize or generate new names.
- Store uploaded files outside the web root.

## Dependencies

- Keep dependencies updated. Known vulnerabilities in outdated packages are the most common attack vector.
- Review new dependencies before adding them: check maintenance status, download count, and known vulnerabilities.
- Prefer well-maintained packages with active security response teams over abandoned or unknown packages.
