---
paths: "**/*"
version: "1.0.0"
---
<!-- keel:generated -->

# Code Quality

Rules for writing clean, maintainable, production-grade code. These apply to all files in the project.

## Structure & Size

- Functions: max 50 lines. If longer, extract helpers. If helpers are only used once, keep them in the same file.
- Files: max 200 lines. If longer, split by responsibility.
- Nesting: max 3 levels deep. Use early returns to flatten.
- Classes/structs: single responsibility. If you need "and" to describe what it does, split it.

## Early Returns

Always use early returns over nested conditions:

```
// BAD
func process(input) {
  if input.valid {
    if input.authorized {
      // 20 lines of logic
    }
  }
}

// GOOD
func process(input) {
  if !input.valid { return error }
  if !input.authorized { return error }
  // 20 lines of logic
}
```

## Naming

- Use domain-specific names that reflect business concepts.
- NEVER use generic names: `utils`, `helpers`, `common`, `shared`, `misc`, `stuff`.
- Name files, functions, and variables after what they DO, not what they ARE.
- A `utils.js` with 50 unrelated functions is a code smell. Each function belongs in the module it serves.

```
// BAD
utils/helpers.go
common/shared.ts
lib/misc.py

// GOOD
pricing/calculator.go
auth/token_validator.ts
invoice/generator.py
```

## DRY — But Not Premature

- Three similar lines of code is better than a premature abstraction.
- Extract a helper only when: (1) the pattern repeats 3+ times, AND (2) the abstraction has a clear name, AND (3) the variations are parameterizable.
- When in doubt, duplicate. You can always extract later; undoing a bad abstraction is harder.

## SOLID Principles

- **Single Responsibility**: Each function/class/module does one thing.
- **Open/Closed**: Extend behavior via composition or new implementations, not by modifying existing code.
- **Liskov Substitution**: Subtypes must be substitutable for their base types without breaking behavior.
- **Interface Segregation**: Small, focused interfaces. Don't force implementers to depend on methods they don't use.
- **Dependency Inversion**: Depend on abstractions (interfaces), not concretions. Pass dependencies in, don't instantiate them internally.

## Library-First

- Before writing custom code, search for existing libraries that solve the problem.
- Custom code is justified ONLY for: domain-specific business logic, performance-critical paths with special requirements, or when external dependencies would be overkill.
- Every line of custom code is a liability that needs maintenance, testing, and documentation.

## Separation of Concerns

- Business logic must not live in UI components, HTTP handlers, or database layers.
- Database queries must not appear in controllers/handlers.
- HTTP/transport concerns (status codes, headers, serialization) must not leak into business logic.
- Each layer talks to the next through interfaces, not concrete implementations.

## Imports & Dependencies

- Order imports: standard library, external packages, internal packages. Separate groups with a blank line.
- No circular dependencies between packages/modules.
- A package should not import from a sibling's internal details — only from its public API.

## Comments

- Don't add comments that restate what the code does. Code should be self-documenting through naming.
- DO add comments for: why a non-obvious approach was chosen, business rules that aren't evident from code, workarounds with links to issues, and public API documentation.
- Never leave TODO/FIXME/HACK comments without a linked issue or ticket number.
