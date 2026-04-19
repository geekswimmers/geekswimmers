# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Geek Swimmers is a Go web application designed to help swimmers analyze performance data and reach training goals. The application features user authentication, time standards, swimming records, and a data analysis platform focused on competitive swimming.

**Go Version:** 1.25.0  
**Module Name:** `geekswimmers`

## Architecture Overview

### High-Level Structure

This is a **monolithic web application** structured as a layered MVC-style architecture with the following pattern:

- **Request Entry Point:** `server.Router` (pat.PatternServeMux pattern router)
- **Handler Layer:** Controllers in `modules/<domain>/controller.go` 
- **Business Logic:** Repository pattern in `modules/<domain>/repository.go` + optional service files
- **Data Layer:** PostgreSQL via pgx/v5 connection pool (`storage/postgres.go`)
- **Session Management:** Gorilla sessions with cookie store (`storage/session.go`)
- **Configuration:** Viper-based config from TOML file with environment variable overrides (`config/config.go`)
- **Template Rendering:** Go's standard `html/template` with custom utilities in `utils/template.go`

### Module Organization

Each domain module (in `modules/`) follows this structure:
- `controller.go` - HTTP handlers (ViewName methods for page renders, API methods)
- `model.go` - Data structures (structs representing domain entities)
- `repository.go` - Database queries and data access
- `api.go` or `data.go` - API-specific or data transformation logic
- Optional: `service.go` for business logic, `partial.go` for HTMX partial responses

**Modules:**
- **user** - Authentication, sign-up, profiles, swimmer management, parent-child linking
- **admin** - Admin console, time standards management, meet/record imports, service updates
- **times** - Time standards, records, benchmarks for competitive times
- **swimming** - Swim styles and events data, team resources
- **content** - Blog articles, legal pages, content management
- **web** - Home page, legal documents, public pages

### Data Flow

1. HTTP request arrives at `server.Router` (pat pattern router)
2. Handler middleware checks authentication via `storage.SessionData` (from encrypted cookie)
3. Controller method executes, typically:
   - Initializes request context via `modules.InitializeRequestContext()`
   - Calls repository methods to fetch/mutate data from PostgreSQL
   - Renders template with data context map or returns JSON
4. Template rendering uses layout + page templates from `web/templates/`

### Key Dependencies

- **Database:** PostgreSQL with pgx/v5 connection pool
- **Routing:** pat (URL pattern router)
- **Configuration:** Viper (TOML files + env vars)
- **Session:** Gorilla sessions with secure cookies
- **Templating:** Go standard library + goldmark (markdown to HTML)
- **Security:** bcrypt for password hashing, JWT for Google OAuth
- **Migration:** golang-migrate with SQL migration files
- **CORS:** rs/cors for cross-origin requests

### Frontend Stack

- **CSS:** Bootstrap 5 (minified, no build step)
- **JS:** HTMX for dynamic interactions, vanilla JS in `web/static/js/`
- **Templates:** Go html/template (not a SPA framework)
- **Markdown Rendering:** goldmark (for blog content and articles)

The frontend is **server-rendered HTML with progressive enhancement via HTMX**. No JavaScript build pipeline.

## Build, Test & Development Commands

### Running the Application

```bash
# Development - simple run
go run main.go

# Development - with hot reload (requires wgo: go install github.com/bokwoon95/wgo@latest)
wgo run main.go

# Production - build binary
go build -v ./...
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests for specific package (e.g., utils)
go test geekswimmers/utils

# Run with verbose coverage
go test -v -cover ./...
```

### Code Quality

```bash
# Install golangci-lint (v2.6.0)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.6.0

# Run linter
golangci-lint run
```

### Debugging

```bash
# Install Delve debugger
go install -v github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger (VSCode has pre-configured launch.json in .vscode/)
dlv debug
```

### Docker & Composition

```bash
# Using docker-compose (requires PostgreSQL setup via .env variables)
docker-compose up
```

## Configuration

Configuration is loaded from `config.toml` (or provided via environment variables). Key sections:

- **database** - PostgreSQL connection URL, pool settings
- **server** - HTTP port, session encryption key
- **google** - OAuth2 credentials
- **email** - SMTP settings for transactional emails
- **recaptcha** - reCAPTCHA keys
- **monitoring** - Google Analytics ID
- **miscellaneous** - Feedback form URL

Environment variables override TOML values:
- `DATABASE_URL`, `DATABASE_MAXOPENCONNS`, `DATABASE_CONNMAXLIFETIME`
- `EMAIL_*`, `GOOGLE_*`, `RECAPTCHA_*`
- `PORT`, `SERVER_SESSION_KEY`, `SERVER_URL`, `MONITORING_GOOGLE_ANALYTICS`, `MISCELLANEOUS_FEEDBACKFORM`

**To generate a secure session key:**
```bash
go run cmd/security/security.go
```

## Database

### Migrations

Database migrations are SQL files in `storage/migrations/` using the golang-migrate format:
- `00000X_*.up.sql` - Up migrations (numbered sequentially)
- `00000X_*.down.sql` - Down migrations (if implemented)

Migrations run automatically on application startup via `storage.MigrateDatabase()` before connection pool initialization.

### Connection Pool

Configured via:
- `database.maxopenconns` (default connection limit)
- `database.connmaxlifetime` (idle connection lifespan in minutes)

The pool uses pgx/v5 with a simple interface: `Database` interface in `storage/postgres.go` wraps pgxpool methods.

## Common Patterns & Conventions

### Error Handling

Errors are typically wrapped with context:
```go
return fmt.Errorf("functionName.%v", err)
```
This creates an error chain showing the call path.

### Template Function Helpers

Available in `utils/template.go`:
- `Title(str)` - Converts "FREESTYLE_RELAY" → "Freestyle Relay"
- `Lowercase(str)` - Lowercase conversion
- `MarkdownToHTML(str)` - Markdown to HTML with target="_blank" on links and img-fluid class
- `GetTemplate(layout, page)` - Loads layout + page templates

### Session Data Structure

`storage.SessionData` holds user session values extracted from encrypted cookie. Check authentication with:
```go
if !sessionData.IsAuthenticated() { /* user not logged in */ }
```

Role-based access controlled in server middleware:
- `handleAuthRequest()` - Requires authentication
- `handleAuthApiRequest()` - Requires auth + returns JSON errors
- `handleAdminAuthRequest()` - Requires admin role

### Repository Query Pattern

Repositories use raw SQL with pgx:
```go
stm := `SELECT ... FROM ... WHERE ...`
rows, err := db.Query(context.Background(), stm, args...)
// or
row := db.QueryRow(context.Background(), stm, args...)
```

No ORM is used; all queries are hand-written SQL for control and clarity.

### Time Handling

Swimming times are stored as integers (milliseconds). Utilities in `utils/time.go` handle conversions:
- Text format: "MM:SS.MS" (e.g., "01:24.99")
- Internal: milliseconds since event start

## Testing Patterns

Tests are located alongside code (e.g., `config/config_test.go`, `modules/times/model_test.go`).

Table-driven test pattern is used (see `utils/template_test.go`):
```go
testCases := []struct {
    name     string
    input    string
    expected string
}{...}
for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

## Important Routing Rules

The router order in `server/server.go` **must be respected** - patterns are matched in order, and some routes must come before others (e.g., specific paths before parameterized paths).

Route groups by function:
- **Web/Public** - Home, legal docs, signup, auth
- **Admin** - Console, standards, meets, service updates (require admin role)
- **Profile** - User profiles, swimmers, best times (require auth)
- **Content** - Blog articles, styling guides
- **Times** - Standards, records, benchmarks (public)
- **API** - Teams, events, best times (some require auth)
- **Static** - CSS, JS, images from `web/static/`

## CI/CD

GitHub Actions runs on every push and PR:
- **Go tests:** `go test -v ./...`
- **Go build:** `go build -v ./...`
- **Linter:** golangci-lint-action (configured for v2.6.0)

## Key Files to Know

- `main.go` - Entry point: loads config, migrates DB, initializes pool, starts HTTP server
- `server/server.go` - Router setup and all route definitions
- `config/config.go` - Configuration loading and Viper integration
- `storage/postgres.go` - Database connection pool and migration setup
- `storage/session.go` - Session management with Gorilla
- `utils/template.go` - Template loading and helper functions
- `modules/controller.go` - Base context initialization helper
