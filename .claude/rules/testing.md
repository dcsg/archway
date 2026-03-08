---
paths: "**/*"
version: "1.0.0"
---
<!-- keel:generated -->

# Testing

Rules for writing reliable, maintainable tests. Every feature and bugfix must have tests.

## Test-Driven Development

Follow the Red-Green-Refactor cycle:

1. **RED** — Write one failing test that describes the expected behavior.
2. **Verify RED** — Run it. Confirm it fails for the RIGHT reason (missing feature, not a typo).
3. **GREEN** — Write the minimal code to make the test pass. Nothing more.
4. **Verify GREEN** — Run it. Confirm it passes. Confirm other tests still pass.
5. **REFACTOR** — Clean up. Keep tests green. Don't add behavior.
6. **Repeat** — Next failing test for next behavior.

If you wrote production code before writing a failing test, delete it and start over with TDD.

## Test Naming

Use descriptive names that document behavior:

```
// BAD
test('test1')
test('it works')
test('user test')

// GOOD
test('rejects empty email with validation error')
test('retries failed operations up to 3 times')
test('returns cached result when called within TTL')
```

If you need "and" in the test name, split it into two tests.

## One Behavior Per Test

Each test verifies exactly one behavior. Multiple assertions are fine if they all verify the same behavior.

```
// BAD — two behaviors in one test
test('validates email and saves user')

// GOOD — separate tests
test('rejects invalid email format')
test('saves user with valid data')
```

## Test Real Behavior, Not Mocks

- Test what the code DOES, not what the mocks DO.
- Never assert on mock element existence (e.g., `getByTestId('sidebar-mock')`).
- Use real implementations wherever possible. Only mock: external services (APIs, databases), time-dependent operations, and non-deterministic code.

## Mock Anti-Patterns

Mocks are tools to isolate, not things to test. These patterns lead to tests that pass but prove nothing.

### Don't test mock behavior
If your assertion checks that a mock was called or that a mock element exists, you're testing the test setup, not the code.

### Don't add test-only methods to production code
If a method only exists for tests (e.g., `destroy()`, `reset()`), put it in test utilities instead.

### Don't mock without understanding dependencies
Before mocking any method, ask: "What side effects does the real method have? Does this test depend on any of them?" If yes, mock at a lower level — mock the slow/external operation, not the high-level method.

### Don't use incomplete mocks
Mock the COMPLETE data structure as it exists in reality, not just the fields your immediate test uses. Partial mocks hide structural assumptions and cause silent failures.

## Edge Cases & Error Paths

Every feature needs tests for:
- Happy path (expected input, expected output)
- Invalid input (empty, null, wrong type, too large, too small)
- Boundary conditions (zero, one, max, off-by-one)
- Error paths (network failure, timeout, permission denied)
- Concurrent access (if applicable)

## Test Organization

- Test files live next to the code they test, or in a parallel `test/` directory — follow the project's existing convention.
- Group related tests with describe/context blocks.
- Share setup with beforeEach/setUp, not by copy-pasting.
- Test utilities and factories live in a dedicated test helpers directory.

## When Tests Are Hard to Write

If a test requires extensive setup, deep mocking, or access to private internals — the code under test has a design problem. Listen to the test:
- Hard to test = hard to use
- Needs many mocks = too many dependencies
- Needs private access = wrong abstraction boundary

Simplify the code, then test it.
