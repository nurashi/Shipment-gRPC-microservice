Task for Vektor TMS company,
# Shipment gRPC Microservice

A gRPC microservice for managing shipments and tracking status changes during transportation. Built with Go, following Clean Architecture principles.

## How to Run

### Using Docker Compose

```bash
docker compose up --build
```

This starts PostgreSQL 17 and the application. The gRPC server listens on port `50051`.

### Running locally

1. Start a PostgreSQL instance (or use the one from Docker Compose):
   ```bash
   docker compose up db -d
   ```

2. Set environment variables (`.env`):

Example data in .env.example

Note: environment data is for local running.




Tests cover domain logic and application use cases with manual mock implementations of repository interfaces. No external dependencies (database, gRPC) required to run tests.

## Architecture Overview

The project follows Clean Architecture with three layers:

![Architecture Diagram](./doc/image.png)

Thanks for: https://mermaid.live/edit 


### Key Design Principle

Every layer boundary is crossed via interfaces:
- gRPC handler calls `application.ShipmentService` (interface)
- Application service calls `ports.ShipmentRepository` and `ports.ShipmentEventRepository` (interfaces)
- Domain objects (structs) are shared across layers; proto types never leak beyond the gRPC handler

## Design Decisions

### Shipment Status Lifecycle

The shipment status follows a strict state machine:

```
PENDING -> PICKED_UP -> IN_TRANSIT -> DELIVERED
   |         |            |
   V         V            V
CANCELLED  CANCELLED   CANCELLED
```

- Every shipment starts as `PENDING`
- `DELIVERED` and `CANCELLED` are terminal states — no further transitions allowed
- Cancellation is possible from any non-terminal state
- Invalid transitions (`PENDING -> DELIVERED`) are rejected with a descriptive error
- Duplicate transitions (`PICKED_UP -> PICKED_UP`) are rejected
- Each status change is recorded as an immutable `ShipmentEvent`

### Why Clean Architecture?

- Domain logic is testable without any infrastructure (no DB, no gRPC)
- Repositories can be swapped (PostgreSQL -> SQLite) by implementing the same interface
- Transport layer can be replaced (gRPC -> REST) without touching business logic
- Each layer has a single reason to change

### Manual Dependency Injection

Dependencies are wired manually in `cmd/main.go` with no DI framework. This keeps the dependency chain explicit and easy to trace:

```
config -> pgxpool -> postgres.Repos -> application.Service -> grpc.Handler -> grpc.Server
```

### Error Mapping

Domain errors are mapped to appropriate gRPC status codes at the transport boundary:

| Domain Error | gRPC Code |
|---|---|
| `ErrShipmentNotFound` | `NOT_FOUND` |
| `ErrInvalidStatusTransition` | `INVALID_ARGUMENT` |
| `ErrShipmentTerminated` | `FAILED_PRECONDITION` |
| `ErrDuplicateReferenceNumber` | `ALREADY_EXISTS` |
| `ErrMissingRequiredField` | `INVALID_ARGUMENT` |

### Testing Strategy

Tests focus on business behavior, not framework plumbing:
- Domain tests validate the state machine, shipment creation, and transition rules
- Application tests use manual mock structs implementing port interfaces to verify use case orchestration
- No mocking library needed - Go interfaces make this straightforward

## Assumptions

1. **Single-node deployment**: No distributed locking or event sourcing. The service assumes a single instance or relies on PostgreSQL for consistency.
2. **No authentication**: The gRPC service does not implement auth interceptors. In production, this would be handled by an API gateway or gRPC interceptor middleware.
3. **Migrations at startup**: The SQL migration runs automatically when the service starts. 
4. **Monetary fields as float64**: `Amount` and `DriverRevenue` use `float64` / `NUMERIC(12,2)`. A production system would use a decimal library to avoid floating-point precision issues.
5. **UTC timestamps**: All timestamps are stored as `TIMESTAMPTZ` and managed in UTC.
6. **Reference number uniqueness**: Each shipment must have a unique reference number, enforced at both the application and database level.

## Project Structure
```
❯ tree
.
├── cmd
│   └── main.go               # Entry point
├── config
│   └── config.go             # Environment configuration
├── docker-compose.yml        # App + PostgreSQL 17
├── Dockerfile                # Multi-stage build
├── gen
│   └── shipment
│       ├── shipment_grpc.pb.go
│       └── shipment.pb.go    # Generated Go code
├── go.mod
├── go.sum
├── internal
│   ├── application
│   │   ├── service.go        # Use cases
│   │   └── service_test.go   # Application unit tests
│   ├── domain                # Business entities & rules
│   │   ├── errors.go         # Domain errors
│   │   ├── event.go          # ShipmentEvent value object
│   │   ├── shipment.go       # Shipment aggregate
│   │   ├── shipment_test.go  # Domain unit tests
│   │   └── status.go         # Status state machine
│   ├── infrastructure
│   │   ├── grpc
│   │   │   ├── handler.go    # gRPC handler (proto <-> domain)
│   │   │   └── server.go     # gRPC server lifecycle
│   │   └── persistence
│   │       └── postgres
│   │           └── repository.go # PostgreSQL implementation
│   └── ports
│       └── repository.go     # Repository interfaces
├── Makefile                  # Build targets
├── migrations
│   └── 001_init.sql          # Database schema
├── proto
│   └── shipment.proto        # Service contract
└── README.md

15 directories, 23 files
```