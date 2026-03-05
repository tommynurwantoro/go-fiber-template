# Project Initialization Prompt

This file is a structured prompt for AI coding agents (Claude Code, Cursor, Copilot, Windsurf, etc.). Run it immediately after cloning the template to rename the project and configure all files. Once initialization is complete, this file deletes itself.

---

## Step 1: Collect User Inputs

Ask the user for the following three values before proceeding:

1. **Go module path** — the full Go module path for the new project (e.g. `github.com/username/my-project`)
2. **App display name** — a human-readable name for the application (e.g. `My Project API`)
3. **Database name** — the PostgreSQL database name to use (e.g. `myprojectdb`)

Do not proceed until all three values are provided.

---

## Step 2: Derive Values

Compute the following values from the user inputs. Do not ask the user for these.

| Derived Value          | Rule                                                        | Example                  |
|------------------------|-------------------------------------------------------------|--------------------------|
| `binary-name`          | Last segment of the module path                             | `my-project`             |
| `docker-network-name`  | `app-network`                                               | `app-network`            |
| `docker-service-name`  | `<binary-name>-app`                                         | `my-project-app`         |
| `cobra-command-name`   | Same as `binary-name`                                       | `my-project`             |

---

## Step 3: Apply Changes

Apply each change below in order. Verify every replacement is made before moving to the next file.

### a. `go.mod`

Change the module declaration from:

```
module app
```

to:

```
module <module-path>
```

### b. All `.go` files

In every `.go` file in the repository, replace the import prefix `"app/` with `"<module-path>/`. This applies to all import statements referencing internal packages.

### c. `config.yaml`

- Set `app_name` to `<app-display-name>`
- Set `database.name` to `<database-name>`

### d. `.env.example`

Update the following values:

- `APP_NAME=<app-display-name>`
- `DATABASE_NAME=<database-name>`

The `docker-compose.yml` file also requires `DATABASE_NAME`, `DATABASE_USER`, `DATABASE_PASSWORD`, and `DATABASE_PORT` variables. If these do not exist in `.env.example`, add them in a separate "docker-compose configuration" section:

```
# docker-compose configuration
DATABASE_USER=postgres
DATABASE_PASSWORD=changemeinproduction
DATABASE_NAME=<database-name>
DATABASE_PORT=5432
```

If `DATABASE_NAME` already exists, update it to `<database-name>`.

### e. `main.go`

Update the Swagger `@title` annotation from:

```go
// @title go-fiber-template API documentation
```

to:

```go
// @title <app-display-name> API documentation
```

### f. `cmd/cmd.go`

Change the Cobra root command `Use` field from:

```go
Use: "app",
```

to:

```go
Use: "<binary-name>",
```

### g. `Makefile`

Replace `go-fiber-template` with `<binary-name>` in the following targets:

- **build target**: `go build -o ./bin/go-fiber-template` becomes `go build -o ./bin/<binary-name>`
- **migrate-docker-up target**: `--network go-fiber-template_go-network` becomes `--network <binary-name>_<docker-network-name>`
- **migrate-docker-down target**: `--network go-fiber-template_go-network` becomes `--network <binary-name>_<docker-network-name>`

> Note: Docker Compose prefixes network names with the project directory name. This assumes the directory is named `<binary-name>`. For example, if binary-name is `my-project`, the full network becomes `my-project_app-network`.

### h. `docker-compose.yml`

Apply all of the following replacements:

- Rename service `go-app` to `<docker-service-name>` (the service key under `services:`)
- Update `image: go-app` to `image: <docker-service-name>`
- Update volume mount path `.:/usr/src/go-app` to `.:/usr/src/<docker-service-name>`
- Rename network `go-network` to `<docker-network-name>` (in the `networks:` definition and all `networks:` references throughout the file)

### i. `Dockerfile`

No changes needed. The Dockerfile uses a generic `main` binary name.

### j. `internal/adapter/database/init/init.sql`

No changes needed. Keep `testdb` as-is; it is used for the test database.

### k. `.air.toml`

No changes needed. It uses generic paths.

### l. `.github/workflows/build.yml`

Update the following values:

- `APP_NAME: go-fiber-template` becomes `APP_NAME: <app-display-name>`
- `DATABASE_NAME: fiberdb` becomes `DATABASE_NAME: <database-name>`
- `POSTGRES_DB: fiberdb` becomes `POSTGRES_DB: <database-name>`
- In the `pg_isready` command, replace `fiberdb` with `<database-name>`

---

## Step 4: Cleanup

After all changes are applied:

1. Delete `setup.md` (this file) — it is no longer needed after initialization.
2. Delete the `docs/plans/` directory if it exists.
3. Delete the `.git` directory to remove the template's git history.
4. Run `git init` to initialize a fresh repository.
5. Run `go mod tidy` to synchronize dependencies.
6. Remove the "Quick Start with AI Agent" section from `README.md` if it exists.

---

## Step 5: Verify

Run the following command and confirm it compiles successfully with no errors:

```bash
make build
```

Report the result to the user. If the build fails, diagnose and fix the issue before finishing.
