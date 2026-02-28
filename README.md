# RESTful API Go Fiber Template

![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)
[![Go Report Card](https://goreportcard.com/badge/github.com/tommynurwantoro/go-fiber-template)](https://goreportcard.com/report/github.com/tommynurwantoro/go-fiber-template)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
![Repository size](https://img.shields.io/github/repo-size/tommynurwantoro/go-fiber-template?color=56BEB8)
![Build](https://github.com/tommynurwantoro/go-fiber-template/workflows/Build/badge.svg)
![Test](https://github.com/tommynurwantoro/go-fiber-template/workflows/Test/badge.svg)
![Linter](https://github.com/tommynurwantoro/go-fiber-template/workflows/Linter/badge.svg)

A template/starter project for quickly building RESTful APIs using Go, Fiber, and PostgreSQL. Inspired by the Express boilerplate.

The app comes with many built-in features, such as authentication using JWT and Google OAuth2, request validation, unit and integration tests, docker support, API documentation, pagination, etc. For more details, check the features list below.

## Quick Start

To create a project, simply run:

```bash
go mod init <project-name>
```

## Manual Installation

If you would still prefer to do the installation manually, follow these steps:

Clone the repo:

```bash
git clone --depth 1 https://github.com/tommynurwantoro/go-fiber-template.git
cd go-fiber-template
rm -rf ./.git
```

Install the dependencies:

```bash
go mod tidy
```

Set up configuration:

```bash
cp .env.example .env
# Open .env and modify the environment variables (if needed)
# Config is primarily loaded from config.yaml; .env overrides defaults
```

## Table of Contents

- [Features](#features)
- [Commands](#commands)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Error Handling](#error-handling)
- [Validation](#validation)
- [Authentication](#authentication)
- [Authorization](#authorization)
- [Logging](#logging)
- [Linting](#linting)
- [Contributing](#contributing)

## Features

- **SQL database**: [PostgreSQL](https://www.postgresql.org) Object Relation Mapping using [Gorm](https://gorm.io)
- **Database migrations**: with [golang-migrate](https://github.com/golang-migrate/migrate)
- **Validation**: request data validation using [Package validator](https://github.com/go-playground/validator)
- **Logging**: using [golog](https://github.com/tommynurwantoro/golog)
- **Testing**: unit and integration tests using [Testify](https://github.com/stretchr/testify)
- **Error handling**: centralized error handling mechanism
- **API documentation**: with [Swag](https://github.com/swaggo/swag) and [Swagger](https://github.com/gofiber/swagger)
- **Sending email**: using [Gomail](https://github.com/go-gomail/gomail)
- **Configuration**: [config.yaml](config.yaml) with [Viper](https://github.com/spf13/viper) (env vars override defaults)
- **Security**: set security HTTP headers using [Fiber-Helmet](https://docs.gofiber.io/api/middleware/helmet)
- **CORS**: Cross-Origin Resource-Sharing enabled using [Fiber-CORS](https://docs.gofiber.io/api/middleware/cors)
- **Compression**: gzip compression with [Fiber-Compress](https://docs.gofiber.io/api/middleware/compress)
- **Dependency injection**: using [gontainer](https://github.com/tommynurwantoro/gontainer)
- **Docker support**
- **Linting**: with [golangci-lint](https://golangci-lint.run)

## Commands

Running locally:

```bash
make start
```

Or running with live reload:

```bash
air
```

> [!NOTE]
> Make sure you have `Air` installed.\
> See 👉 [How to install Air](https://github.com/air-verse/air)

Testing:

```bash
# run all tests
make tests

# run all tests with gotestsum format
make testsum

# run test for the selected function name
make tests-TestUserModel
```

> [!IMPORTANT]
> Tests use a **separate test database**.
>
> Make sure the test database already exists and all required tables (`users`, `tokens`, etc.) have been migrated before running the test commands.

Docker:

```bash
# run docker container
make docker

# run all tests in a docker container
make docker-test
```

Linting:

```bash
# run lint
make lint
```

Swagger:

```bash
# generate the swagger documentation
make swagger
```

Migration:

```bash
# Create migration
make migration-<table-name>

# Example for table users
make migration-users
```

```bash
# run migration up in local
make migrate-up

# run migration down in local
make migrate-down

# run migration up in docker container
make migrate-docker-up

# run migration down all in docker container
make migrate-docker-down
```

## Configuration

The app uses [config.yaml](config.yaml) as the primary configuration file. Environment variables in `.env` override the defaults. Add variables to `.env` only when you need to override the config.yaml values.

**config.yaml** structure:

```yaml
app_name: ""
app_version: ""
environment: ""  # prod || dev

http:
  host: "0.0.0.0"
  port: 8888

database:
  host: ""
  port: 5432
  name: ""
  user: ""
  password: ""

jwt:
  secret: ""
  expire: 30m
  refresh_expire: 7d

smtp:
  host: ""
  port: 587
  # ...

oauth2:
  google_client_id: ""
  google_client_secret: ""
  redirect_url: ""
```

**Environment variables** (`.env` – override config.yaml):

```bash
# server configuration
APP_NAME=app
APP_VERSION=v1.0.0
APP_PORT=8888
ENVIRONMENT=development

# database configuration
DATABASE_HOST=postgresdb
DATABASE_NAME=fiberdb
DATABASE_USER=postgres
DATABASE_PASSWORD=changemeinproduction

# for docker-compose
DB_USER=postgres
DB_PASSWORD=changemeinproduction
DB_NAME=fiberdb
DB_PORT=5432

# JWT
JWT_SECRET=changemeinproduction

# SMTP configuration
SMTP_HOST=email-server
SMTP_PORT=587
SMTP_USERNAME=changemeinproduction
SMTP_PASSWORD=changemeinproduction
SMTP_FROM=support@yourapp.com

# OAuth2 configuration
GOOGLE_CLIENT_ID=yourapps.googleusercontent.com
GOOGLE_CLIENT_SECRET=thisisasamplesecret
REDIRECT_URL=http://localhost:8888/auth/google-callback
```

## Project Structure

```
.
├── cmd/                    # Application entrypoints (Cobra commands)
│   ├── cmd.go
│   └── service.go
├── config/                 # Configuration (roles, tokens)
│   ├── model.go
│   ├── roles.go
│   └── tokens.go
├── config.yaml             # Primary configuration file
├── docs/                   # Swagger generated files
├── internal/
│   ├── adapter/            # External adapters (REST, DB, OAuth, Email)
│   │   ├── database/      # GORM connection & migrations
│   │   ├── oauth/
│   │   └── rest/          # Fiber app & error handling
│   ├── application/       # Application layer (handlers, services, models)
│   │   ├── handler/
│   │   ├── model/
│   │   ├── router/
│   │   └── service/
│   ├── bootstrap/         # App bootstrap & DI
│   ├── domain/            # Domain entities & errors
│   └── pkg/               # Shared packages (middleware, validator, crypto, etc.)
├── main.go                # Entry point
└── .env                   # Environment overrides (copy from .env.example)
```

## API Documentation

To view the list of available APIs and their specifications, run the server and go to `http://localhost:8888/v1/docs` in your browser.

![Auth](https://tommynurwantoro.github.io/assets/images/swagger1.png)
![User](https://tommynurwantoro.github.io/assets/images/swagger2.png)

This documentation page is automatically generated using the [Swag](https://github.com/swaggo/swag) definitions written as comments in the handler files.

See 👉 [Declarative Comments Format.](https://github.com/swaggo/swag#declarative-comments-format)

## API Endpoints

List of available routes:

**Auth routes** (`/auth`):\
`POST /auth/register` - register\
`POST /auth/login` - login\
`POST /auth/logout` - logout\
`POST /auth/refresh-tokens` - refresh auth tokens\
`POST /auth/forgot-password` - send reset password email\
`POST /auth/reset-password` - reset password\
`POST /auth/send-verification-email` - send verification email\
`POST /auth/verify-email` - verify email\
`GET /auth/google` - login with google account\
`GET /auth/google-callback` - OAuth2 callback

**User routes** (`/v1/users`):\
`POST /v1/users` - create a user\
`GET /v1/users` - get all users\
`GET /v1/users/:userId` - get user\
`PATCH /v1/users/:userId` - update user\
`DELETE /v1/users/:userId` - delete user

**Health check**:\
`GET /health-check` - health check endpoint

## Error Handling

The app includes a centralized error handling mechanism in `internal/adapter/rest/error_handler.go` and `internal/pkg/formatter/response.go`.

It also utilizes the `Fiber-Recover` middleware to gracefully recover from any panic that might occur in the handler stack, preventing the app from crashing unexpectedly.

The error handling process sends an error response in the following format:

```json
{
  "status": "error",
  "message": "Not found",
  "traceId": "request-id"
}
```

Fiber provides a custom error struct using `fiber.NewError()`, where you can specify a response code and a message. This error can then be returned from any part of your code, and Fiber's `ErrorHandler` will automatically catch it.

For example, if you are trying to retrieve a user from the database but the user is not found:

```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return fiber.NewError(fiber.StatusNotFound, "User not found")
}
```

## Validation

Request data is validated using [Package validator](https://github.com/go-playground/validator). Check the [documentation](https://pkg.go.dev/github.com/go-playground/validator/v10) for more details on how to write validations.

The validation is handled in the application layer. Custom validators and error mapping are in `internal/pkg/validator/`.

## Authentication

To require authentication for certain routes, you can use the `JWTAuth` middleware:

```go
import (
    "app/internal/application/handler"
    "app/internal/pkg/middleware"

    "github.com/gofiber/fiber/v2"
)

user := v1.Group("/users")
user.Post("/", r.AuthMiddleware.JWTAuth("manageUsers"), r.UserHandler.CreateUser)
```

These routes require a valid JWT access token in the Authorization request header using the Bearer schema. If the request does not contain a valid access token, an Unauthorized (401) error is thrown.

**Generating Access Tokens**:

An access token can be generated by making a successful call to the register (`POST /auth/register`) or login (`POST /auth/login`) endpoints. The response of these endpoints also contains refresh tokens.

Access and refresh token expiration times are configured in `config.yaml` under the `jwt` section.

**Refreshing Access Tokens**:

After the access token expires, a new access token can be generated by making a call to the refresh token endpoint (`POST /auth/refresh-tokens`) and sending along a valid refresh token in the request body.

## Authorization

The `JWTAuth` middleware supports role-based permissions. Pass the required rights as arguments:

```go
user.Get("/", r.AuthMiddleware.JWTAuth("getUsers"), r.UserHandler.GetUsers)
user.Post("/", r.AuthMiddleware.JWTAuth("manageUsers"), r.UserHandler.CreateUser)
```

In the example above, an authenticated user can access the route only if that user has the `manageUsers` permission.

The permissions are role-based. You can view the permissions/rights of each role in the `config/roles.go` file:

```go
var allRoles = map[string][]string{
    "user":  {},
    "admin": {"getUsers", "manageUsers"},
}
```

If the user making the request does not have the required permissions, a Forbidden (403) error is thrown.

## Logging

The app uses [golog](https://github.com/tommynurwantoro/golog) for structured logging. Import and use it in your handlers and services:

```go
import "github.com/tommynurwantoro/golog"

golog.Info("message")
golog.Error("message", err)
golog.Panic("message")
```

Logging configuration (file location, rotation, stdout) is defined in `config.yaml` under the `log` section.

## Linting

Linting is done using [golangci-lint](https://golangci-lint.run)

See 👉 [How to install golangci-lint](https://golangci-lint.run/welcome/install)

To modify the golangci-lint configuration, update the `.golangci.yml` file.

## Contributing

Contributions are more than welcome! Please check out the [contributing guide](CONTRIBUTING.md).

If you find this template useful, consider giving it a star! ⭐

## Inspirations

- [indravscode/go-fiber-boilerplate](https://github.com/indravscode/go-fiber-boilerplate)

## License

[MIT](LICENSE)

## Contributors

[![Contributors](https://contrib.rocks/image?c=6&repo=tommynurwantoro/go-fiber-template)](https://github.com/tommynurwantoro/go-fiber-template/graphs/contributors)
