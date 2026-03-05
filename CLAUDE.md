# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development
make start          # Run the application
air                 # Run with live reload (requires Air installed)

# Build
make build          # Build binary to ./bin/go-fiber-template

# Testing
make tests          # Run all tests
make tests-TestUserModel  # Run specific test (replace dots with underscores)
make testsum        # Run tests with gotestsum formatting

# Linting
make lint           # Run golangci-lint

# Swagger
make swagger        # Generate swagger docs (run after modifying handler annotations)

# Docker
make docker         # Start containers (app + postgres + adminer)
make docker-down    # Stop and remove containers, volumes, images

# Migrations (local)
make migrate-up     # Apply migrations
make migrate-down   # Rollback migrations
make migration-<table>  # Create new migration file

# Migrations (Docker)
make migrate-docker-up    # Apply migrations in Docker network
make migrate-docker-down  # Rollback migrations in Docker network
```

## Architecture Overview

This is a Go Fiber REST API using **layered architecture** with dependency injection and elements of hexagonal architecture:

```
cmd/                    # Cobra CLI entrypoints
internal/
  adapter/              # External adapters (implements ports)
    database/           # GORM database adapter
      repository/       # Repository implementations (adapters)
      mocks/            # Generated mocks
    email/              # Email adapter (gomail)
    oauth/              # OAuth2 adapters (Google)
    rest/               # Fiber app setup, error handling
  application/          # Application layer
    handler/            # HTTP handlers (controllers)
    model/              # Request/response DTOs
    router/             # Route registration
    service/            # Business logic (use cases)
      mocks/            # Generated mocks
  bootstrap/            # DI container setup, app startup
  domain/               # Domain layer (core)
    repository/         # Repository INTERFACES (ports)
    myerrors/           # Centralized domain errors
    user.go             # Domain entities
    token.go
  pkg/                  # Shared packages
    crypto/             # Password hashing
    formatter/          # Response formatting
    middleware/         # Auth middleware (JWT)
    token/              # Token utilities
    validator/          # Request validation
config/                 # Config models, roles, token types
```

### Key Patterns

**Dependency Injection**: Uses `gontainer` with struct tags:
```go
type UserServiceImpl struct {
    UserRepository repository.UserRepository `inject:"userRepository"`
    Validator      validator.Validator       `inject:"validator"`
}
```
Services are registered in `internal/bootstrap/` and auto-injected on `appContainer.Ready()`.

**Repository Pattern (Hexagonal)**: Data access through repository interfaces:
- Repository **interfaces** (ports) in `internal/domain/repository/`
- Repository **implementations** (adapters) in `internal/adapter/database/repository/`
- Implementations use GORM via `database.DatabaseAdapter` interface
- Services depend on interfaces, not implementations

**Framework-Free Services**: Services use `context.Context` instead of framework-specific types:
```go
func (u *UserServiceImpl) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*domain.User, error)
```
This allows services to be reused across different contexts (HTTP, gRPC, CLI, message queues).

**Domain-Driven Errors**: Centralized in `internal/domain/myerrors/`:
```go
var ErrUserNotFound = errors.New("user not found")
```

**Validation**: Request DTOs use go-playground/validator tags. Custom validators in `internal/pkg/validator/custom_validator.go`.

## Adding New Features

1. **Create domain entity** in `internal/domain/` (if needed)
2. **Create repository interface** in `internal/domain/repository/` with `//go:generate mockgen` directive
3. **Create repository implementation** in `internal/adapter/database/repository/`
4. **Register repository** in `internal/bootstrap/adapter.go`
5. **Create service** in `internal/application/service/` with `//go:generate mockgen` directive
6. **Register service** in `internal/bootstrap/application.go`
7. **Create handler** in `internal/application/handler/`
8. **Register handler** in `internal/bootstrap/application.go`
9. **Add routes** in `internal/application/router/router.go`
10. **Generate mocks**: `go generate ./...`
11. **Generate swagger docs**: `make swagger`

## Mock Generation

Mocks are generated with `go:generate` directives:
```go
//go:generate mockgen -source=user_service.go -destination=mocks/user_service.go -package=mocks
```

Run after modifying interfaces:
```bash
go generate ./...
```

## Configuration

Primary config: `config.yaml`. Override with `.env` file (copy from `.env.example`).

Key config sections:
- `http`: Server host/port
- `database`: PostgreSQL connection
- `jwt`: Token settings
- `smtp`: Email settings
- `oauth2`: Google OAuth

## Testing

Tests use `testify/suite` with in-memory SQLite:
```go
type userRepositoryTestSuite struct {
    suite.Suite
    mockCtrl *gomock.Controller
    gormDB   *gorm.DB
    repo     *UserRepositoryImpl
}

func (s *userRepositoryTestSuite) SetupTest() {
    // Setup in-memory DB per test
}
```

**Important**: Tests require a separate test database with migrated tables when running integration tests.

## Authentication

JWT-based auth via `AuthMiddleware.JWTAuth()`:
```go
user.Get("/", r.AuthMiddleware.JWTAuth("getUsers"), r.UserHandler.GetUsers)
```

Role-based permissions defined in `config/roles.go`:
```go
var allRoles = map[string][]string{
    "user":  {},
    "admin": {"getUsers", "manageUsers"},
}
```
