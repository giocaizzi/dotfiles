---
name: sqlalchemy
description: SQLAlchemy v2 best practices and conventions. Use this anytime you are using SQLAlchemy. You must follow these guidelines always.
metadata:
  version: "2.0"
---

# SQLAlchemy v2 Best Practices

## Core Principles
- **Always use v2 API patterns** - avoid deprecated v1.x query methods
- Use `Mapped[T]` and `mapped_column()` for type-safe ORM models
- Use async sessions only when your application is async (FastAPI, async web servers)
- Use sync sessions for traditional applications, scripts, and simpler use cases
- Use `select()` statements instead of legacy `Query` interface

## Model Definition

```python
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy import String, ForeignKey
from typing import Optional

class Base(DeclarativeBase):
    pass

class User(Base):
    __tablename__ = "users"
    
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(50))
    email: Mapped[Optional[str]]  # Nullable column
    
    posts: Mapped[list["Post"]] = relationship(back_populates="user")

class Post(Base):
    __tablename__ = "posts"
    
    id: Mapped[int] = mapped_column(primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id"))
    title: Mapped[str]
    
    user: Mapped["User"] = relationship(back_populates="posts")
```

## Session Management

### Sync Sessions (Default)
```python
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, Session

engine = create_engine("postgresql://user:pass@localhost/db")
SessionLocal = sessionmaker(engine)

def get_session() -> Session:
    with SessionLocal() as session:
        yield session
```

### Async Sessions (For async applications only)
```python
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession

engine = create_async_engine("postgresql+asyncpg://user:pass@localhost/db")
AsyncSessionLocal = async_sessionmaker(engine, expire_on_commit=False)

async def get_session() -> AsyncSession:
    async with AsyncSessionLocal() as session:
        yield session
```

## Query Patterns (v2 Style)

### Select Queries
```python
from sqlalchemy import select

# Simple select (sync)
stmt = select(User).where(User.name == "Alice")
result = session.execute(stmt)
user = result.scalar_one_or_none()

# Multiple results (sync)
stmt = select(User).where(User.email.isnot(None))
result = session.execute(stmt)
users = result.scalars().all()

# For async, just add await:
result = await session.execute(stmt)

# Joins
stmt = select(User).join(User.posts).where(Post.title.like("%SQL%"))
result = session.execute(stmt)
users = result.scalars().unique().all()
```

### Insert/Update/Delete
```python
# Insert (sync)
new_user = User(name="Bob", email="bob@example.com")
session.add(new_user)
session.commit()

# Update (sync)
stmt = select(User).where(User.id == 1)
result = session.execute(stmt)
user = result.scalar_one()
user.email = "newemail@example.com"
session.commit()

# Delete (sync)
session.delete(user)
session.commit()

# For async, add await to commit/execute:
# await session.commit()
```

## Relationships and Loading

### Eager Loading
```python
from sqlalchemy.orm import selectinload, joinedload

# selectinload (separate query)
stmt = select(User).options(selectinload(User.posts))

# joinedload (single query with JOIN)
stmt = select(User).options(joinedload(User.posts))
```

## Anti-Patterns to Avoid

**Don't use legacy Query API:**
```python
session.query(User).filter_by(name="Alice")  # Old v1 style
```

**Use select() instead:**
```python
select(User).where(User.name == "Alice")  # v2 style
```

**Don't use implicit type inference:**
```python
name = Column(String(50))  # Old style
```

**Use Mapped with explicit types:**
```python
name: Mapped[str] = mapped_column(String(50))  # v2 style
```

## Error Handling

```python
from sqlalchemy.exc import IntegrityError, NoResultFound

try:
    session.commit()  # or await session.commit() for async
except IntegrityError:
    session.rollback()  # or await session.rollback() for async
    # Handle constraint violation
```

## Key Reminders
- Always use `select()` for queries, never `session.query()`
- Use `Mapped[T]` annotations for all model attributes
- Prefer `scalar_one()`, `scalar_one_or_none()`, `scalars().all()` for results
- Use `relationship()` with `back_populates` for bidirectional relationships
- Use sync sessions by default; only use async if your entire application is async
- Async requires `await` for `execute()`, `commit()`, `rollback()`, etc.