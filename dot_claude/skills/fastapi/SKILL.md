---
name: fastapi
description: FastAPI best practices and conventions with Pydantic v2. Use this anytime you are working with FastAPI. Follow these guidelines always.
metadata:
  version: "2.0"
  updated: "2025-01"
---

# FastAPI Best Practices

Modern FastAPI with **Pydantic v2** for building production-ready RESTful APIs.

Enforce idiomatic patterns, correct code smells, and maintain high-quality FastAPI applications.

**Remember**: FastAPI + Pydantic v2 provides automatic validation, serialization, and documentation. Leverage these features—don't fight them.

## Project Structure

```
project/
├── main.py              # App factory and startup
├── config.py            # Settings with Pydantic BaseSettings
├── dependencies.py      # Shared dependencies
├── api/
│   └── v1/
│       ├── __init__.py
│       └── routes/      # APIRouter modules by domain
├── services/            # Business logic layer
├── models/              # SQLAlchemy/DB models
└── schemas/             # Pydantic schemas
```

## Application Factory & Settings

- **Always use application factory pattern** to create FastAPI instances
- **Use Pydantic v2 `BaseSettings`** for configuration (never raw `os.environ`)
- Load settings with `model_config = SettingsConfigDict(env_file=".env")`

```python
from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", case_sensitive=False)
    
    database_url: str
    debug: bool = False
```

## Dependency Injection

- **Use `Annotated[Type, Depends(...)]`** for all dependencies (Pydantic v2 style)
- **Never use default arguments** like `= Depends(...)`
- Inject database sessions, services, and current user through dependencies
- Avoid global state—dependencies provide scoped resources

```python
from typing import Annotated

async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_session() as session:
        yield session

DBSession = Annotated[AsyncSession, Depends(get_db)]

@router.get("/users/{user_id}")
async def get_user(user_id: int, db: DBSession):
    ...
```

## Routes & Endpoints

- **Type everything** with Pydantic models: request bodies, responses, query params, path params
- **Use async** for all endpoints and I/O operations (database, HTTP, file operations)
- **Organize with `APIRouter`** by feature/domain (users, products, orders)
- **Validate input** with Pydantic models—never trust raw data
- Use status codes from `fastapi.status` module

```python
from typing import Annotated
from fastapi import Query, Path

@router.get("/items", response_model=ItemListResponse)
async def list_items(
    params: Annotated[ItemQueryParams, Query()],
    db: DBSession,
):
    return await item_service.list_items(db, params)
```

## Pydantic v2 Schema Conventions

### Naming Patterns

| Type | Pattern | Example |
|------|---------|---------|
| Request | `{Action}{Entity}Request` | `CreateUserRequest`, `UpdateProductRequest` |
| Response | `{Entity}Response` | `UserResponse`, `ProductResponse` |
| List Response | `{Entity}ListResponse` | `UserListResponse` |
| Query Params | `{Entity}QueryParams` | `ProductQueryParams` |
| Path Params | `{Entity}PathParams` | `UserPathParams` |

### Base Schemas (DRY)

Create base schemas to share common fields across related models:

```python
class UserBase(BaseModel):
    email: EmailStr
    username: str

class CreateUserRequest(UserBase):
    password: str

class UserResponse(UserBase):
    id: int
    created_at: datetime
```

## Error Handling

- **Custom exception handlers** that convert domain exceptions → HTTP responses
- **Consistent error format**: `{"error": str, "code": str, "details": dict | None}`
- **Include tracebacks only in development** (never in production)
- Use `HTTPException` for expected errors, custom handlers for domain exceptions

```python
@app.exception_handler(DomainException)
async def domain_exception_handler(request: Request, exc: DomainException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": exc.message, "code": exc.code}
    )
```

## OpenAPI Documentation

- **Customize OpenAPI schema** with meaningful metadata and descriptions
- **Use `tags`** to group related endpoints in documentation
- **Add `summary` and `description`** to all endpoints for clarity
- **Document response models** with examples using `response_model` and `responses`
- **Include examples** in Pydantic schemas with `model_config` and `json_schema_extra`

```python
app = FastAPI(
    title="My API",
    description="Production-ready API with FastAPI",
    version="1.0.0",
    openapi_tags=[
        {"name": "users", "description": "User management"},
        {"name": "products", "description": "Product operations"},
    ]
)

@router.post("/users", tags=["users"], response_model=UserResponse)
async def create_user(
    request: CreateUserRequest,
    db: DBSession,
):
    """
    Create a new user account.
    
    - **email**: Valid email address
    - **username**: Unique username (3-50 chars)
    - **password**: Min 8 characters
    """
    return await user_service.create(db, request)

# Add examples to schemas
class CreateUserRequest(BaseModel):
    email: EmailStr
    username: str
    password: str
    
    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "email": "user@example.com",
                    "username": "john_doe",
                    "password": "securepass123"
                }
            ]
        }
    )
```

## Code Architecture

- **Layered architecture**: Routes → Services → Repository/Models
- **Routes are thin**: Just call service methods, handle HTTP concerns
- **Services contain business logic**: Validation, calculations, orchestration
- **Use DDD principles** where complexity warrants it, keep simple otherwise

## Configuration & Observability

- **Structured logging** with context (request_id, user_id, correlation_id)
- **OpenTelemetry** for tracing and metrics in production
- **CORS middleware** configured for allowed origins (never `allow_origins=["*"]` in prod)
- **Health check endpoints** (`/health`, `/ready`) for orchestration

## Prohibited Patterns

- ❌ Raw `os.environ` access (use Settings)
- ❌ `= Depends(...)` as default arguments (use `Annotated`)
- ❌ Global database sessions or state
- ❌ Synchronous I/O in async endpoints
- ❌ Missing type annotations on route parameters
- ❌ Exposing tracebacks in production errors

## Development Server

```bash
# With uv (recommended)
uv run uvicorn project.main:app --reload --host 0.0.0.0 --port 8000

# With standard Python
uvicorn project.main:app --reload --host 0.0.0.0 --port 8000
```

## Testing

- Use `TestClient` from `fastapi.testclient` for integration tests
- Override dependencies for testing (database, external services)
- Use `pytest` with `pytest-asyncio` for async test support

