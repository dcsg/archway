---
paths: "**/*"
version: "1.0.0"
---
<!-- keel:generated -->

# Architecture

Rules for projects with complex domain logic. Applies Domain-Driven Design and Clean Architecture principles to keep business logic independent, testable, and maintainable.

Enable this rule pack when your project has multiple bounded contexts, complex business rules, or domain logic that goes beyond simple CRUD.

## Layer Boundaries

Organize code into layers with strict dependency direction — inner layers never import from outer layers:

```
Domain (innermost)    → entities, value objects, domain events, repository interfaces
Application           → use cases, application services, DTOs
Infrastructure        → database, external APIs, file system, message queues
Transport (outermost) → HTTP handlers, CLI, GraphQL resolvers, gRPC
```

**Enforced import restrictions:**
- `domain/` must NOT import from `application/`, `infrastructure/`, or `transport/`
- `application/` must NOT import from `infrastructure/` or `transport/`
- `infrastructure/` must NOT import from `transport/`
- Dependencies always point inward

## Bounded Contexts

- Each bounded context gets its own package/module. No cross-context domain imports.
- Contexts communicate through: application-level services, domain events, or explicit anti-corruption layers.
- Never share domain entities between contexts. If two contexts need similar data, each defines its own representation.

```
// BAD — ordering context imports catalog's domain
import { Product } from '../catalog/domain/product'

// GOOD — ordering context has its own representation
import { OrderItem } from './domain/order-item'
```

## Ubiquitous Language

- Code names (classes, methods, variables) must match the domain language used by stakeholders.
- If the business says "order" don't call it "purchase" in code. If they say "shipment" don't call it "delivery".
- When domain terms are ambiguous, clarify with stakeholders and document in the bounded context's glossary.

## Entities & Value Objects

- **Entities** have identity (ID) and lifecycle. Two entities with the same data but different IDs are different.
- **Value Objects** are defined by their attributes, not identity. Two value objects with the same data are equal.
- Use value objects for domain concepts instead of primitives: `Money` instead of `float`, `EmailAddress` instead of `string`, `DateRange` instead of two dates.

```
// BAD — primitive obsession
func calculateTotal(price float64, currency string, quantity int)

// GOOD — value objects
func calculateTotal(price Money, quantity Quantity) Money
```

## Aggregates

- An aggregate is a cluster of entities and value objects treated as a single unit for data changes.
- All modifications go through the aggregate root. External code never modifies internal entities directly.
- Keep aggregates small. If an aggregate loads too much data, the boundary is wrong.
- Reference other aggregates by ID, not by direct object reference.

## Repository Pattern

- Repository interfaces live in the domain layer. They define what the domain needs (FindByID, Save, etc.).
- Repository implementations live in the infrastructure layer. They handle the actual database/storage.
- Repositories operate on aggregates, not individual entities or database tables.

```
// domain/repository.go — interface
type OrderRepository interface {
    FindByID(id OrderID) (*Order, error)
    Save(order *Order) error
}

// infrastructure/postgres/order_repo.go — implementation
type PostgresOrderRepository struct { db *sql.DB }
func (r *PostgresOrderRepository) FindByID(id OrderID) (*Order, error) { ... }
```

## Domain Events

- Use domain events to communicate that something significant happened within a bounded context.
- Events are past-tense facts: `OrderPlaced`, `PaymentReceived`, `ShipmentDispatched`.
- Events carry only the data needed by consumers — not the entire aggregate state.
- Events enable loose coupling between bounded contexts. The publisher doesn't know who listens.

## Application Services / Use Cases

- Each use case is a single class/function that orchestrates domain objects to accomplish one business operation.
- Use cases depend on domain interfaces (repositories, domain services), not infrastructure.
- Use cases handle transaction boundaries, authorization checks, and cross-aggregate coordination.
- Keep use cases thin — business logic belongs in domain objects, not in the use case orchestrator.

## Anti-Corruption Layer

When integrating with external systems or legacy code:
- Define your own domain model. Don't let external data structures leak into your domain.
- Create adapters that translate between the external model and your domain model.
- The rest of your code should never know the external system's data format.
