---
name: ddd
description: Domain Driven Design (DDD) - Concepts, Patterns, and Best Practices. Use this anytime you are designing or working within a DDD architecture. You must follow these guidelines always.
metadata:
  version: "1.0"
---

# Domain-Driven Design (DDD) Best Practices

Separate your application into distinct layers and modules following DDD principles to ensure maintainability, scalability, and clear separation of concerns.

**Domain Layer:** Focuses on the core business logic and rules.
**Application Layer:** Orchestrates use cases and application workflows.
**Infrastructure Layer:** Handles technical details like persistence and external systems.

## Shared Kernel
**Rule:** Only stable, truly cross-cutting abstractions and infrastructure belong here. Context-specific code stays in its bounded context.

## References
- Eric Evans — Domain-Driven Design: Tackling Complexity in the Heart of Software
- Vaughn Vernon — Implementing Domain-Driven Design

## Key Principles

- **Ubiquitous Language:** Use a common language shared by developers and domain experts.
- **Bounded Contexts:** Isolate different parts of the domain into separate contexts with clear boundaries.
- **Entities:** Objects with a distinct identity that persists over time.
- **Value Objects:** Immutable objects defined by their attributes rather than identity.
- **Aggregates:** Clusters of related entities and value objects treated as a single unit.
- **Repositories:** Abstractions for data access, providing methods to retrieve and persist aggregates.
- **Domain Services:** Stateless operations that encapsulate domain logic not naturally fitting within entities or value objects.


## Domain Layer Rules

- **NO I/O, NO frameworks, NO infrastructure dependencies**
- **Aggregates:** Transactional boundaries. Only aggregate root is repository-accessible. Co-locate entities and events in `domain/aggregates/{name}/`
- **Entities:** Private `__init__`, factory classmethods (`create`, `reconstitute`), immutable IDs. Collect domain events via `_domain_events: list[DomainEvent]`, expose `pull_domain_events()`
- **Value Objects:** Immutable dataclasses, self-validating, `__eq__`/`__hash__` by value
- **Repository Protocols:** `domain/repositories/` — collection-like interface (get, save, find_by_*, delete)
- **Port Protocols:** `domain/ports/` — external system interfaces (email, SMS, payments, event publishers)
- **Domain Errors:** Inherit from `DomainError`, pure domain logic errors

## Application Layer Rules

- **Thin orchestration:** Delegates business logic to domain
- **Commands:** Intent to modify state (`{Action}{Entity}Command`)
- **Queries:** Read-only requests (`Get{Entity}Query`, `List{Entities}Query`)
- **Handlers:** One handler per use case. Opens UoW → loads aggregates → calls domain methods → collects events → commits → publishes events
- **Event Handlers:** `application/handlers/events/` — React to domain events with I/O (send email, update analytics, notify)
- **Ports:** Application-level interfaces for external systems (`application/ports/`)
- **Application Errors:** Inherit from `ApplicationError`, application logic errors

## Infrastructure Layer Rules

- **Implements domain/application interfaces**
- **Persistence:** `infrastructure/persistence/` — concrete repositories, mappers
- **Adapters:** `infrastructure/adapters/{category}/` — implements port interfaces
- **Messaging:** `infrastructure/messaging/` — event bus, publishers, subscribers
- **Clients:** `infrastructure/clients/` — external API clients (HTTP, DB drivers)
- **All implementations MUST explicitly inherit from their protocols**
- **Infrastructure Errors**: Inherit from `InfrastructureError`, technical errors (DB connection, HTTP failures)

## Event-Driven Architecture

- **Domain Events:** Raised by aggregates during state changes, past-tense names (`OrderPlaced`, `ShopCreated`)
- **Event Bus:** `shared/infrastructure/messaging/` — in-process pub/sub for application event handlers
- **Integration Events:** Cross-context communication via messaging infrastructure
- **Handler Flow:** Repository saves aggregate → pulls domain events → event bus publishes → event handlers execute side effects

## Naming Conventions

**Never add extra suffixes like `_handler` or `_service`, as the folder structure already provides context.**

### Application DTOs

| Type | Pattern | Example |
| ---- | ------- | ------- |
| Command | `{Action}{Entity}Command` | `CreateShopCommand` |
| Command Result | `{Action}{Entity}Result` | `CreateShopResult` |
| Query | `Get{Entity}Query` | `GetUserQuery` |
| Query Result | `Get{Entity}Result` | `GetUserResult` |
| List Query | `List{Entities}Query` | `listUsersQuery` |
| List Result | `List{Entities}Result` | `ListUsersResult` |


## Example Project and component structure and details

```

project_root/
│
├─ shared/
│  │  # SHARED KERNEL: Cross-cutting types and interfaces used across bounded contexts
│  │  # RULE: Only stable, well-established abstractions that rarely change
│  │  # CONTAINS: Base classes, common value objects, result types
│  │
│  ├─ api/
│  │  # Shared API concerns
│  │  # PUT HERE: Common schemas, dependencies, error handlers
│  │  # AVOID: Context-specific request/response types
│  │
│  ├─ domain/
│  │  │  # Shared domain abstractions
│  │  │
│  │  ├─ value_objects/
│  │  │  # Common value objects shared across multiple aggregates
│  │  │  # PUT HERE: Money, Email, UserId, Timestamp, PhoneNumber, Address
│  │  │  # AVOID: Domain-specific value objects (those belong in context domain/)
│  │  │
│  │  ├─ events.py
│  │  │  # DomainEvent base class/protocol
│  │  │  # PUT HERE: Base event interface, event metadata
│  │  │
│  │  ├─ errors/
│  │  │  # Base domain errors
│  │  │  # PUT HERE: DomainError, common error types
│  │  │
│  │  └─ ports/
│  │     # Shared port interfaces used by multiple contexts
│  │     # PUT HERE: StorageService, common external system interfaces
│  │     # RULE: Only ports needed by 2+ bounded contexts
│  │
│  └─ infrastructure/
│     │  # SHARED INFRASTRUCTURE: Cross-cutting technical implementations
│     │  # RULE: Only truly cross-cutting concerns that serve multiple contexts
│     │  # PRINCIPLE: Avoid duplication, but don't force sharing
│     │
│     ├─ config/
│     │  # Application configuration and settings
│     │  # PUT HERE: Environment variables, feature flags, secrets management
│     │  # PURPOSE: Single source of truth for app config
│     │
│     ├─ persistence/
│     │  # Database clients and shared persistence utilities
│     │  # PUT HERE: DB connection pools, base repository helpers, common mappers
│     │  # PURPOSE: Single connection pool shared across contexts
│     │  # AVOID: Context-specific repository implementations
│     │
│     ├─ logging/
│     │  # Structured logging infrastructure
│     │  # PUT HERE: Logger configuration, formatters, log context
│     │  # PURPOSE: Consistent logging format across contexts
│     │
│     ├─ telemetry/
│     │  # Observability infrastructure
│     │  # PUT HERE: Tracing, metrics, distributed context propagation
│     │  # PURPOSE: Unified observability across all contexts
│     │
│     ├─ event_bus/
│     │  # In-process event publishing
│     │  # PUT HERE: EventBus implementation, subscriber registry
│     │  # PURPOSE: Domain event dispatch to application handlers
│     │  # PATTERN: Publish-subscribe for decoupled event handling
│     │
│     └─ adapters/
│        # Shared adapter implementations
│        # PUT HERE: Storage adapters, common external integrations
│        # RULE: Only adapters used by multiple contexts
│
├─ contexts/
│  │  # BOUNDED CONTEXTS: Each context is a self-contained module
│  │  # RULE: One folder per bounded context
│  │  # STRUCTURE: Each context follows the same internal structure
│  │  # COMMUNICATION: Contexts communicate via integration events only
│  │
│  └─ {context_name}/
│     │  # Example: order_management, inventory, billing, user_management
│     │  # PRINCIPLE: High cohesion within context, loose coupling between contexts
│     │
│     ├─ api/
│     │  # ENTRY LAYER: HTTP endpoints, dependency injection, serialization
│     │  # RESPONSIBILITY: Translate HTTP ↔ application layer
│     │  # RULE: No business logic here, only routing and validation
│     │
│     │  ├─ routes/
│     │  │  # HTTP endpoints organized by resource or use case
│     │  │  # PUT HERE: FastAPI route definitions, path operations
│     │  │  # PATTERN: Receive request → validate → dispatch command/query → return response
│     │  │  # AVOID: Business logic, direct DB access, domain knowledge
│     │  │
│     │  ├─ dependencies/
│     │  │  # Dependency injection providers (FastAPI Depends)
│     │  │  # PUT HERE: Factory functions for UoW, repositories, ports, use cases
│     │  │  # PURPOSE: Wire up concrete implementations at runtime
│     │  │
│     │  ├─ schemas/
│     │  │  # Pydantic models for HTTP serialization/deserialization
│     │  │  # PUT HERE: Request/response DTOs, validation rules
│     │  │  # RULE: These are NOT domain models, only transport format
│     │  │  # PATTERN: Named *Request, *Response
│     │  │
│     │  └─ middleware/
│     │     # Cross-cutting HTTP concerns
│     │     # PUT HERE: Authentication, error translation, request logging, CORS
│     │     # RESPONSIBILITY: Convert application errors → HTTP status codes
│     │     # AVOID: Business logic, direct DB access
│     │
│     ├─ domain/
│     │  # PURE BUSINESS LOGIC: Framework-independent, side-effect-free
│     │  # RULE: NO I/O, no frameworks, no infrastructure dependencies
│     │  # GOAL: Testable without mocks, expresses ubiquitous language
│     │
│     │  ├─ aggregates/
│     │  │  # Aggregate roots: transactional consistency boundaries
│     │  │  # STRUCTURE: One folder per aggregate, co-locate entities and events
│     │  │  # RESPONSIBILITY: Enforce invariants, coordinate entities, raise domain events
│     │  │  # RULE: Only aggregate root is repository-accessible
│     │  │
│     │  │  └─ {aggregate_name}/
│     │  │     │  # Self-contained aggregate with all its parts
│     │  │     │
│     │  │     ├─ {aggregate_root}.py
│     │  │     │  # The aggregate root entity
│     │  │     │  # PUT HERE: Public methods that enforce business rules
│     │  │     │  # CONTAINS: Invariant validation, event raising, entity coordination
│     │  │     │  # PATTERN: Methods are verbs (place_order, cancel, ship)
│     │  │     │
│     │  │     ├─ {entity}.py
│     │  │     │  # Entities within the aggregate (have identity but not aggregate roots)
│     │  │     │  # PUT HERE: Supporting entities that exist only within aggregate lifecycle
│     │  │     │  # RULE: Never accessed directly by repositories, always through root
│     │  │     │
│     │  │     └─ events.py
│     │  │        # Domain events for this aggregate
│     │  │        # PUT HERE: Immutable event classes (past-tense names)
│     │  │        # STRUCTURE: Data classes with timestamp, aggregate_id, event data
│     │  │        # EXAMPLES: OrderPlaced, OrderCancelled, PaymentReceived
│     │  │
│     │  ├─ value_objects/
│     │  │  # Immutable objects compared by value, not identity
│     │  │  # PUT HERE: Money, Address, OrderStatus, Quantity, DateRange
│     │  │  # CHARACTERISTICS: No identity, immutable, validated on creation
│     │  │  # PATTERN: Implement __eq__, __hash__ based on all attributes
│     │  │
│     │  ├─ specifications/
│     │  │  # Reusable business rule objects (Specification pattern)
│     │  │  # PUT HERE: Complex eligibility rules, validation logic, queries
│     │  │  # PURPOSE: Make business rules first-class, testable, composable
│     │  │  # PATTERN: is_satisfied_by(candidate) → bool
│     │  │  # EXAMPLES: CustomerIsPremium, OrderCanBeShipped, ProductIsInStock
│     │  │  # COMPOSITION: Support and/or/not operations
│     │  │
│     │  ├─ services/
│     │  │  # Domain services: operations that span multiple aggregates
│     │  │  # PUT HERE: Logic that doesn't naturally fit in a single aggregate
│     │  │  # CHARACTERISTICS: Stateless, pure functions, no side effects
│     │  │  # WHEN TO USE: Rarely! Most logic belongs in aggregates
│     │  │  # EXAMPLES: PricingService (uses Product + Customer + Promotions)
│     │  │  # AVOID: CRUD operations, orchestration (those are application concerns)
│     │  │
│     │  ├─ repositories/
│     │  │  # Repository interfaces (contracts, not implementations)
│     │  │  # PUT HERE: Abstract base classes defining aggregate persistence
│     │  │  # PATTERN: get(id), save(aggregate), find_by_*, delete(id)
│     │  │  # RULE: One repository per aggregate root
│     │  │  # CHARACTERISTIC: Collection-like interface, hides persistence details
│     │  │
│     │  └─ errors/
│     │     # Domain-specific exceptions for business rule violations
│     │     # PUT HERE: Exceptions representing domain concepts
│     │     # EXAMPLES: InsufficientInventory, InvalidOrderState, AccountLocked
│     │     # RULE: Never reference infrastructure (no DB errors, no HTTP codes)
│     │     # PATTERN: Inherit from base DomainError class
│     │
│     ├─ application/
│     │  # USE CASE ORCHESTRATION: Coordinate domain + infrastructure
│     │  # RESPONSIBILITY: Transaction boundaries, workflow coordination, policies
│     │  # RULE: Thin layer, delegates to domain for business logic
│     │
│     │  ├─ commands/
│     │  │  # Command messages: intent to modify system state
│     │  │  # PUT HERE: Immutable data classes representing user actions
│     │  │  # STRUCTURE: All data needed to execute command
│     │  │  # EXAMPLES: PlaceOrderCommand, CancelOrderCommand, UpdateProfileCommand
│     │  │  # PATTERN: Imperative names (verbs), include user_id/correlation_id
│     │  │
│     │  ├─ queries/
│     │  │  # Query messages: read-only data requests
│     │  │  # PUT HERE: Immutable data classes for fetching data
│     │  │  # EXAMPLES: GetOrderSummaryQuery, ListCustomerOrdersQuery
│     │  │  # RULE: No side effects, no business logic
│     │  │
│     │  ├─ handlers/
│     │  │  # Message handlers: receive and dispatch commands/queries/events
│     │  │  # RESPONSIBILITY: Thin dispatch layer
│     │  │
│     │  │  ├─ commands/
│     │  │  │  # Command handlers
│     │  │  │  # PUT HERE: Validation, authorization check, call use case
│     │  │  │  # AVOID: Business logic (delegate to use cases/domain)
│     │  │  │  # PATTERN: handle(command) → Result[DTO, Error]
│     │  │  │
│     │  │  ├─ queries/
│     │  │  │  # Query handlers
│     │  │  │  # PUT HERE: Fetch read models, simple transformations
│     │  │  │  # RULE: No business logic, direct read from read models/repos
│     │  │  │  # PATTERN: handle(query) → Result[DTO, Error]
│     │  │  │
│     │  │  └─ events/
│     │  │     # Application event handlers: react to domain events
│     │  │     # PUT HERE: Side-effect coordination after domain events occur
│     │  │     # EXAMPLES: Send email when order placed, update analytics
│     │  │     # PATTERN: handle(event) → trigger ports (async, can fail)
│     │  │     # RULE: ALL event handling with I/O happens here, not in domain
│     │  │
│     │  ├─ use_cases/
│     │  │  # Use case implementations (transaction scripts)
│     │  │  # PUT HERE: Application workflow orchestration
│     │  │  # RESPONSIBILITY: Open UoW → load aggregates → call domain methods
│     │  │  #                 → collect events → commit → publish events
│     │  │  # PATTERN: execute() → Result[aggregate_id, Error]
│     │  │  # CONTAINS: Transaction boundaries, aggregate loading, event publishing
│     │  │  # AVOID: Business rules (delegate to domain)
│     │  │
│     │  ├─ policies/
│     │  │  # Application-level policies and cross-cutting concerns
│     │  │  # PUT HERE: Authorization, rate limiting, idempotency, retries
│     │  │  # EXAMPLES: RBACPolicy, IdempotencyPolicy, RateLimitPolicy
│     │  │  # CHARACTERISTIC: Reusable, configurable, non-domain logic
│     │  │
│     │  ├─ ports/
│     │  │  # Interfaces for external systems (hexagonal architecture)
│     │  │  # PUT HERE: Abstract definitions of external dependencies
│     │  │  # EXAMPLES: IEmailSender, IPaymentGateway, IEventPublisher
│     │  │  # RULE: Domain/application defines interface, infrastructure implements
│     │  │  # PATTERN: Interface methods return Result or raise port-specific errors
│     │  │
│     │  ├─ uow.py
│     │  │  # Unit of Work interface
│     │  │  # DEFINES: Transaction boundary, repository access, event collection
│     │  │  # PATTERN: Context manager (__enter__/__exit__), commit(), rollback()
│     │  │  # RESPONSIBILITY: Aggregate repository access, collect domain events
│     │  │
│     │  └─ errors/
│     │     # Application-level errors
│     │     # PUT HERE: Orchestration failures, wrapped domain/infra errors
│     │     # EXAMPLES: UseCaseExecutionFailed, ResourceNotFound
│     │     # RESPONSIBILITY: Map lower-level errors to application semantics
│     │     # CONTAINS: Error aggregation, context enrichment
│     │
│     └─ infrastructure/
│        │  # TECHNICAL IMPLEMENTATIONS: I/O, persistence, external systems
│        │  # RULE: Implements interfaces defined in domain/application
│        │  # CHARACTERISTIC: Replaceable, framework-specific, has side effects
│        │
│        ├─ persistence/
│        │  │  # Data persistence layer
│        │  │
│        │  ├─ repositories/
│        │  │  # Concrete repository implementations
│        │  │  # PUT HERE: ORM queries, aggregate reconstruction from DB
│        │  │  # IMPLEMENTS: Domain repository interfaces
│        │  │  # RESPONSIBILITY: Load/save aggregates, maintain aggregate boundaries
│        │  │  # AVOID: Business logic, return ORM models (return domain models)
│        │  │
│        │  ├─ models/
│        │  │  # ORM models (SQLAlchemy, Django ORM, etc.)
│        │  │  # PUT HERE: Database schema definitions, table mappings
│        │  │  # RULE: Separate from domain models, use mappers to convert
│        │  │  # PURPOSE: Optimize for storage, not domain logic
│        │  │
│        │  ├─ mappers/
│        │  │  # Bidirectional mapping: Domain ↔ ORM
│        │  │  # PUT HERE: to_domain(orm_model), to_persistence(domain_model)
│        │  │  # RESPONSIBILITY: Data transformation, handle relationships
│        │  │  # PATTERN: One mapper per aggregate
│        │  │
│        │  ├─ uow.py
│        │  │  # Concrete Unit of Work implementation
│        │  │  # PUT HERE: Database session management, transaction control
│        │  │  # IMPLEMENTS: application/uow.py interface
│        │  │  # CONTAINS: Session/connection lifecycle, event collection
│        │  │
│        │  ├─ outbox/
│        │  │  # (OPTIONAL) Transactional outbox pattern for reliable event publishing
│        │  │  # PUT HERE: Outbox table, event storage, polling mechanism
│        │  │  # PURPOSE: Atomically save aggregate + events in same transaction
│        │  │  # COMPONENTS: store.py (save), processor.py (poll + publish)
│        │  │
│        │  └─ errors/
│        │     # Persistence-specific errors
│        │     # PUT HERE: DBConnectionError, UniqueViolation, DeadlockDetected
│        │
│        ├─ adapters/
│        │  │  # External system adapters (implement application ports)
│        │  │
│        │  └─ {adapter_category}/
│        │     # Group by external system type
│        │     # PUT HERE: Concrete implementations of port interfaces
│        │     # EXAMPLES: email/, payment/, notifications/, storage/
│        │     # PATTERN: Multiple implementations per port (SMTP vs SendGrid)
│        │     # RESPONSIBILITY: API calls, error translation, retries
│        │
│        ├─ messaging/
│        │  # Message bus / integration events (cross-context communication)
│        │  # PUT HERE: Event bus implementation, message routing
│        │  # COMPONENTS: Publisher, subscribers, retry logic, dead letter queue
│        │  # PROTOCOLS: RabbitMQ, Kafka, Redis Streams, AWS SQS
│        │  # PURPOSE: Publish domain events to other bounded contexts
│        │
│        ├─ workers/
│        │  # Background workers and async processors
│        │  # PUT HERE: Outbox consumers, scheduled jobs, retry handlers
│        │  # EXAMPLES: Celery tasks, cron jobs, message queue consumers
│        │  # RESPONSIBILITY: Poll for work, process async, handle failures
│        │
│        └─ config/
│           # Configuration and settings
│           # PUT HERE: Environment variables, secrets, feature flags
│           # PATTERN: Pydantic Settings, python-decouple
│           # ORGANIZE: (Optional) By environment (dev, staging, prod)

```
