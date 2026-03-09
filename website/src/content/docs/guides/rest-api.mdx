---
title: Building a REST API
description: Step-by-step guide to scaffolding a production REST API
---

import { Aside, Steps } from '@astrojs/starlight/components';

This guide walks you through creating a production-ready REST API with Archway.

<Steps>

1. **Scaffold the project**

   ```bash
   archway new orders-api \
     --arch hexagonal \
     --cap platform,bootstrap,http-api,mysql,migrations,health,docker,linting \
     --module github.com/myorg/orders-api \
     --no-wizard
   ```

   This creates a project with hexagonal architecture, REST API on Chi, MySQL with migrations, health checks, Docker Compose, and linting.

2. **Explore the structure**

   ```
   orders-api/
   ├── cmd/orders-api/main.go           # Thin entry point (15 lines)
   ├── internal/bootstrap/bootstrap.go  # All dependency wiring
   ├── domain/
   │   ├── errors.go                    # ErrNotFound, ValidationError, etc.
   │   └── clock.go                     # Testable time
   ├── port/
   │   ├── inbound.go                   # Use case interfaces
   │   └── outbound.go                  # Repository interfaces
   ├── service/service.go               # Business logic
   ├── adapter/
   │   ├── httphandler/
   │   │   ├── router.go                # Chi routes
   │   │   ├── handler.go               # HTTP handlers
   │   │   ├── response.go              # RFC 7807 errors
   │   │   ├── middleware.go            # Custom middleware
   │   │   └── pagination.go           # Cursor pagination
   │   └── mysqlrepo/
   │       └── connection.go            # MySQL setup
   ├── config/config.go                 # Config loading
   ├── docs/PROJECT.md                  # Project anatomy
   └── archway.yaml                     # Architecture rules
   ```

3. **Define your domain**

   Start in `domain/` — this is your innermost layer with zero dependencies.

   ```go
   // domain/order.go
   package domain

   import "time"

   type OrderID string

   type Order struct {
       ID        OrderID
       Customer  string
       Items     []OrderItem
       Total     Money
       CreatedAt time.Time
   }

   type OrderItem struct {
       Product  string
       Quantity int
       Price    Money
   }

   type Money struct {
       Amount   int64  // cents
       Currency string
   }
   ```

4. **Define ports**

   Ports are interfaces that connect your domain to the outside world.

   ```go
   // port/inbound.go — what the service offers
   type OrderService interface {
       CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error)
       GetOrder(ctx context.Context, id domain.OrderID) (*domain.Order, error)
   }

   // port/outbound.go — what the service needs
   type OrderRepository interface {
       Save(ctx context.Context, order *domain.Order) error
       FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
   }
   ```

5. **Implement the service**

   ```go
   // service/order.go
   func (s *Service) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error) {
       order := &domain.Order{
           ID:        domain.OrderID(uuid.New().String()),
           Customer:  cmd.Customer,
           Items:     cmd.Items,
           CreatedAt: s.clock.Now(),
       }
       if err := s.orderRepo.Save(ctx, order); err != nil {
           return nil, fmt.Errorf("save order: %w", err)
       }
       return order, nil
   }
   ```

6. **Wire the HTTP handler**

   ```go
   // adapter/httphandler/handler.go
   func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
       var req CreateOrderRequest
       if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
           MapError(w, r, h.logger, &domain.ValidationError{
               Field: "body", Message: "invalid JSON",
           })
           return
       }

       order, err := h.orderSvc.CreateOrder(r.Context(), toCommand(req))
       if err != nil {
           MapError(w, r, h.logger, err)
           return
       }

       JSON(w, http.StatusCreated, toResponse(order))
   }
   ```

   <Aside type="tip">
   Errors automatically map to RFC 7807 responses:
   - `domain.ErrNotFound` → `404`
   - `domain.ValidationError` → `400` with field details
   - `domain.ErrConflict` → `409`
   - Unknown errors → `500` (logged, not exposed to clients)
   </Aside>

7. **Run it**

   ```bash
   # Start MySQL
   docker compose up -d

   # Run the service
   go run ./cmd/orders-api
   ```

8. **Validate architecture**

   ```bash
   archway check
   ```

   This ensures your domain code doesn't import from adapters, your service doesn't bypass ports, and all components follow the dependency rules.

</Steps>
