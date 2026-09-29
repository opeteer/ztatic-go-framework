# Ztatic Full-Stack Go Framework

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Architecture](https://img.shields.io/badge/Architecture-HOTW-success?style=flat-square)](https://github.com/opeteer/ztatic-go-framework)
[![Security](https://img.shields.io/badge/Security-Built--in-blue?style=flat-square&logo=shield)](https://github.com/opeteer/ztatic-go-framework)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)

**Ztatic** is a modern full-stack web framework for Go built on top of **Echo v5**. It enables you to build fast, reactive, real-time web applications using the **HOTW (HTML Over The Wire)** stack: combining **Templ**, **Hotwire (Turbo 8)**, **Alpine.js**, native in-memory **esbuild** bundling, **interactive OpenAPI 3.0 docs**, and **Realtime Pub/Sub (SSE & WebSockets)**.

Write pure Go and HTML—**no Node.js, npm, or Webpack required**—and compile your entire application (templates, assets, database migrations, and server) into a **single, self-contained binary**.

---

## Why Ztatic?

- **Zero Node.js Overhead**: Go bindings to `esbuild` transpile TypeScript, bundle JavaScript, and compile CSS in sub-10ms directly in memory.
- **HTML Over The Wire (HOTW)**: Stream reactive HTML updates straight from Go using **Templ** templates and **Hotwire Turbo 8**, keeping client-side state minimal and simple.
- **Smart Layout Unwrapping**: Inside `<turbo-frame>` containers, Ztatic detects frame requests and renders only the requested HTML fragment, cutting network bandwidth by up to 80%.
- **Real-Time Push without JS Boilerplate**: Stream DOM mutations over Server-Sent Events (SSE) or WebSockets with `<turbo-stream-from>` and watch the page update live.
- **Automated REST & Interactive Docs**: Point `rapid.RegisterResource` at any repository to get full CRUD endpoints and interactive **Scalar UI** docs at `/docs` generated from struct tags.
- **Sensible Security Defaults**: Built-in WAF request inspection, Content Security Policy (CSP) nonces, Double-Submit CSRF, rate limiting, Argon2id password hashing, and AES-256 field encryption.
- **Type-Safe Config & Dotenv**: Strongly-typed struct configuration with 4-tier `.env` cascading, runtime profiles, and `SecretString` to prevent accidental credential leakage in logs.
- **Single-Binary Deployment**: Deploy everything—compiled templates, static assets, database migrations, and backend code—as a single static binary.

---

## Architecture Overview

```text
                  ┌────────────────────────────────────────────────────────┐
                  │                    Client Browser                      │
                  └───────┬────────────────────────┬───────────────────────┘
                          │                        │
               HTML /     │                        │ SSE / WebSocket Stream
         Turbo Streams    │                        │ (<turbo-stream-from>)
                          ▼                        ▼
       ┌────────────────────────────────────────────────────────────────┐
       │                      Ztatic Engine                             │
       │  ┌──────────────────────────────────────────────────────────┐  │
       │  │ Tracing & Logging (W3C Trace Context, Request ID, Slog)  │  │
       │  └────────────────────────────┬─────────────────────────────┘  │
       │                               ▼                                │
       │  ┌──────────────────────────────────────────────────────────┐  │
       │  │ Web Security (WAF, CSP Nonces, CSRF, Activity Ledger)    │  │
       │  └────────────────────────────┬─────────────────────────────┘  │
       │                               ▼                                │
       │  ┌──────────────────────────────────────────────────────────┐  │
       │  │ Echo v5 Core Router (Radix Compaction & TLS Proxy)       │  │
       │  └──────┬─────────────┬─────────────┬─────────────┬─────────┘  │
       │         │             │             │             │            │
       │         ▼             ▼             ▼             ▼            │
       │   ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐      │
       │   │ Fullstack │ │ Realtime  │ │   Data    │ │   Rapid   │      │
       │   │  (Templ + │ │ (Pub/Sub +│ │ (Engine + │ │ (OpenAPI  │      │
       │   │ esbuild)  │ │  WS/SSE)  │ │  Mapper)  │ │ + Scalar) │      │
       │   └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘      │
       │         └─────────────┼─────────────┼─────────────┘            │
       │                       ▼             ▼                          │
       │          ┌──────────────────────────────────────┐              │
       │          │   Config, Dotenv, Secrets & Errors   │              │
       │          └──────────────────────────────────────┘              │
       └────────────────────────────────────────────────────────────────┘
```

---

## Core Features & Packages

### 1. Rapid REST APIs & Scalar Docs (`rapid`)
- **Route Introspection:** Discovers routes and normalizes parameter syntax (`/articles/:id` $\rightarrow$ `/articles/{id}`).
- **OpenAPI 3.0 Generation:** Parses Go struct `json` and `validate` tags to generate OpenAPI 3.0.3 specs with support for nested/embedded structs.
- **Scaffolded Controllers:** `rapid.RegisterResource[T]` mounts standard REST CRUD endpoints (`GET`, `POST`, `PUT`, `DELETE`) with built-in validation and pagination.
- **Interactive Scalar UI:** Serves dynamic API docs at `/docs` and raw OpenAPI JSON at `/docs/openapi.json`.

### 2. Realtime Pub/Sub (`realtime`)
- **Event Brokers:** In-memory `MemoryBroker` for single instances, and distributed `RedisBroker` (`go-redis/v9`) for multi-node deployments.
- **Transports:** Stream Turbo mutations over Server-Sent Events (`SSEHandler`) or full-duplex WebSockets (`WebSocketHandler`).

### 3. Native Asset Pipeline (`fullstack`)
- **In-Memory esbuild:** Compiles and bundles JS, TS, and CSS on the fly in sub-10ms.
- **Dual-Mode Serving:** Disk-backed assets with live reload during development; embedded assets with SHA-256 content hashes and long-term caching in production.

### 4. Database & Data Tier (`data`)
- **DBEngine:** Wraps standard `database/sql` pools with automatic dialect placeholder detection (`$` for PostgreSQL, `?` for SQLite/MySQL).
- **Reflection Mapper:** Maps database rows directly to structs via `db` tags (`pk`, `auto`, `readonly`) and handles field-level AES-256-GCM encryption (`ztatic:"encrypt"`).
- **Fluent Query Helpers:** Compose Squirrel AST queries with clean helper functions (`WhereEq`, `WhereIn`, `WhereLike`, `OrderByDesc`, `Limit`, `Offset`).
- **ACID Transactions & Savepoints:** Run transactions with `dbEngine.TransactionCtx(ctx, fn)`. Ambient transactions propagate through `context.Context` and automatically use SQL Savepoints for nested blocks.
- **Generic BaseRepository:** `BaseRepository[T]` provides ready-to-use CRUD and pagination methods (`Find`, `FindPaginated`, `FindCursor`, `Insert`, `Save`, `DeleteByID`).
- **Embedded Migrations:** Embed Goose SQL migrations with `embed.FS` and apply them at startup.

### 5. Configuration & Environment (`config`)
- **Type-Safe Binding:** Binds environment variables into structs using `env`, `envDefault`, and `validate` tags.
- **4-Tier Dotenv Cascading:** Loads `.env` $\rightarrow$ `.env.local` $\rightarrow$ `.env.{profile}` $\rightarrow$ `.env.{profile}.local` with process environment overrides.
- **Safe Secrets:** `SecretString` and `Secret[T]` mask values as `[REDACTED]` in `fmt` output, logs, and JSON. Plaintext is only exposed through `.Expose()`.
- **Profiles:** Automatic profile detection (`development`, `test`, `staging`, `production`) via `ZTATIC_ENV` or `APP_ENV`.

### 6. Distributed Tracing & Request Correlation (`trace`)
- **W3C Trace Context:** Parses and generates standard `traceparent` (`00-<trace_id>-<span_id>-<flags>`) and `tracestate` headers.
- **Request IDs:** Automatically generates URL-safe timestamped request IDs (`req-...`) and sanitizes incoming `X-Request-ID` headers to prevent header injection.
- **HTTP Client Propagation:** Includes `trace.NewClient(nil)` and `trace.NewTransport(nil)` to automatically propagate tracing headers on outbound HTTP requests.

### 7. Structured Logging (`log`)
- **Built on `log/slog`:** Native Go structured logging with Echo v5 integration.
- **Dual-Mode Output:** Human-friendly colored terminal logs in development, and structured NDJSON in production.
- **Request Logging:** Automatically records request method, path, status, latency, client IP, and request ID.
- **Sensitive Data Masking:** Automatically scrubs passwords, tokens, and secret fields from log records.

### 8. Standardized Errors (`errors`)
- **Clean Architecture Errors:** Create domain errors (`ztatic.ErrNotFound`, `ztatic.ErrValidation`, `ztatic.ErrBadRequest`) without importing `net/http`.
- **Problem Details:** Supports both standard REST error envelopes and RFC 9457 Problem Details (`application/problem+json`).
- **Safe Production Errors:** In production, internal database errors and stack traces are hidden from clients while preserving request correlation IDs.

### 9. Response Envelopes & Pagination (`response`)
- **Canonical Envelope:** Generic `Envelope[T]` standardizes API responses with `success`, `data`, `pagination`, `meta`, and `links`.
- **Empty Slices:** Guarantees empty Go slices serialize as `[]` instead of JSON `null`.
- **Pagination & HATEOAS:** Offset and cursor pagination helpers that automatically generate navigation links and RFC 5988 `Link` headers.

### 10. Web Security (`security`)
- **WAF & Headers:** Payload inspection, multiline XSS filtering, SQL comment normalization, per-request CSP nonces (`fullstack.Nonce(c)`), and HSTS.
- **Hardened CSRF:** Double-Submit Cookie CSRF protection with automatic skipping for `/api/` and `/docs` routes.
- **Auth Tokens & Sessions:** Built-in JWT/AEAD token manager (`security/token`) with role/scope checks, and session stores (`security/session`) with session fixation protection.

### 11. Activity & Audit Logging (`security/audit`)
- **Event Ledger:** Captures actor, action, target resource, outcome, request context, and before/after state diffs.
- **Automatic Auditing:** Mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) and errors are audited automatically. Handlers can enrich entries with `ztatic.AuditFromContext(c)`.

### 12. Developer CLI (`cmd/ztatic`)
- **`ztatic new`**: Scaffolds a new application with recommended directory structure, configuration, and toolchain locks.
- **`ztatic dev`**: Starts the live-reloading dev server with file watching and sub-10ms asset bundling.
- **`ztatic build`**: Runs the 4-step production compiler to generate a standalone static binary.

### 13. File Testing Utility Suite (`filetest` & `fileassert`)
- **Fluent Multipart & Download Client:** `app.TestClient()` eliminates `mime/multipart.Writer` boilerplate, supporting `.Attach()`, `.AttachBytes()`, `.AttachReader()`, `.Field()`, `.WithBearerToken()`, `.WithCookie()`, and Range streaming with HTTP 206 Partial Content verification.
- **100% In-Memory Synthetic Fixtures:** Zero repository bloat. Generates valid images with true dimensions (`filetest.PNG`, `filetest.JPEG`, `filetest.GIF`, `filetest.WebP`, `filetest.SVG`), documents (`filetest.PDF`, `filetest.CSV`, `filetest.Text`, `filetest.JSON`), security payloads (`MaliciousSVG`, `WebShell`, `Executable`, `DisguisedExecutable`, `ZipSlip`, `ZipBomb`), virtual zero-RAM oversized streams (`filetest.Oversized`), and managed disk sandboxes (`filetest.TempSandbox`).
- **Expressive Multi-Level Assertions:** Standalone assertions (`fileassert.Exists`, `fileassert.ContentEquals`, `fileassert.MalwareBlocked`, `fileassert.PartialContent`) and chained fluent response assertions (`resp.AssertOK()`, `resp.AssertDownloaded()`, `resp.AssertNoSniff()`, `resp.AssertSHA256()`).
- **Thread-Safe Test Doubles & Fault Injection:** `MockStorage` and `MockScanner` provide call spy histories (`WasSaved`, `ScannedFiles`, `SaveCallCount`) and deterministic failure injection (`SimulateDiskFull`, `SimulatePermissionDenied`, timeout simulations).

---

## Quick Start

### Prerequisites
- **Go 1.21+** installed
- **templ CLI** installed:
  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```

### 1. Install the CLI

```bash
# Clone the repository
git clone https://github.com/opeteer/ztatic-go-framework.git
cd ztatic-go-framework

# Build the CLI
go build -o ztatic ./cmd/ztatic

# Optionally move to your PATH
sudo mv ztatic /usr/local/bin/
```

### 2. Scaffold a New Project

```bash
ztatic new myapp
cd myapp
```

Generated project layout:
```text
myapp/
├── .env                     # Local environment settings
├── .env.example             # Config blueprint
├── assets/                  # CSS & JS source files
├── cmd/server/main.go       # Application entrypoint
├── db/migrations/           # Goose SQL migrations
├── dist.go                  # Embedded asset filesystem
├── internal/
│   ├── config/config.go     # Config schema struct
│   ├── controllers/         # HTTP handlers
│   ├── models/              # Domain models & DB schemas
│   ├── repositories/        # Database repositories
│   └── views/               # Templ templates & layouts
├── tools.go                 # Toolchain dependency locks
└── go.mod
```

### 3. Start Development Server

```bash
ztatic dev
```

Your app will start at `http://localhost:8080` with automatic live-reloading whenever `.go`, `.templ`, `.css`, `.js`, or `.env` files change.

---

## Code Examples

### 1. Configuration & Dotenv Loading

```go
package config

import "ztatic-go-framework"

type AppConfig struct {
	AppName     string              `env:"APP_NAME" envDefault:"myapp" validate:"required"`
	Port        string              `env:"PORT" envDefault:"8080" validate:"required"`
	DatabaseDSN string              `env:"DATABASE_DSN" envDefault:"app.db" validate:"required"`
	AuthSecret  ztatic.SecretString `env:"AUTH_SECRET" validate:"required,min=32"`
}
```

In `main.go`:
```go
// Loads .env files, binds variables, and validates struct tags at startup
cfg := ztatic.MustLoadConfig[config.AppConfig]()
```

### 2. Database Models & Generic Repositories

Define a model with database tags and validation:

```go
package models

import "time"

type Product struct {
	ID        int       `json:"id" db:"id,primarykey,autoincrement"`
	Name      string    `json:"name" db:"name" validate:"required,min=3"`
	Price     float64   `json:"price" db:"price" validate:"required,min=0.01"`
	Notes     string    `json:"notes,omitempty" db:"notes" ztatic:"encrypt"` // AES-256 encrypted
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

Implement a repository using `BaseRepository[T]`:

```go
package repositories

import (
	"context"

	"ztatic-go-framework/data"
	"myapp/internal/models"
)

type ProductRepository struct {
	*data.BaseRepository[models.Product]
}

func NewProductRepository(db *data.DBEngine) *ProductRepository {
	return &ProductRepository{
		BaseRepository: data.NewBaseRepository[models.Product](db, "products"),
	}
}

// Custom query using fluent helpers
func (r *ProductRepository) FindAffordable(ctx context.Context, maxPrice float64) ([]models.Product, error) {
	return r.Find(ctx,
		data.WhereLtOrEq("price", maxPrice),
		data.OrderByAsc("price"),
	)
}
```

### 3. Rapid REST APIs & Interactive OpenAPI Docs

Mount full REST endpoints in one line:

```go
// Automatically provides:
// - GET    /api/products          (paginated with links)
// - GET    /api/products/:id      (single item)
// - POST   /api/products          (binds, validates, creates)
// - PUT    /api/products/:id      (binds, validates, updates)
// - DELETE /api/products/:id      (removes by ID)
rapid.RegisterResource(app.Group("/api"), "products", productRepo)

// Serve interactive Scalar documentation UI at /docs
rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")
```

### 4. Standard Response Envelopes & Pagination

```go
app.GET("/api/products", func(c *ztatic.Context) error {
	p := ztatic.ExtractPagination(c) // parses ?page=1&per_page=20&sort=price
	products, meta, err := productRepo.FindPaginated(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return ztatic.Paginated(c, products, meta)
})

app.GET("/api/products/:id", func(c *ztatic.Context) error {
	product, err := productRepo.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return ztatic.ErrNotFound("Product not found")
	}
	return ztatic.OK(c, product)
})
```

Example JSON response:
```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "Mechanical Keyboard", "price": 129.99 }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total_items": 45,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  },
  "links": {
    "self": "/api/products?page=1&per_page=20",
    "first": "/api/products?page=1&per_page=20",
    "next": "/api/products?page=2&per_page=20",
    "last": "/api/products?page=3&per_page=20"
  },
  "meta": {
    "request_id": "req-1923e5904b78-b1a4-0001",
    "timestamp": "2026-09-29T10:00:00Z",
    "duration": "1.25ms"
  }
}
```

### 5. Real-Time Updates via SSE & Turbo Streams

Broadcast DOM mutations to connected clients:

```go
broker := realtime.NewMemoryBroker()
app.GET("/sse", realtime.SSEHandler(broker))

app.POST("/products", func(c *ztatic.Context) error {
	var product models.Product
	if err := rapid.BindAndValidate(c, &product); err != nil {
		return err
	}
	_ = productRepo.Insert(c.Request().Context(), &product)

	// Broadcast DOM prepend to all subscribers listening to "products"
	broker.Publish(c.Request().Context(), "products", fullstack.TurboStreamItem{
		Action:    fullstack.StreamPrepend,
		Target:    "product-list",
		Component: components.ProductCard(product),
	})

	return fullstack.RenderTurboStream(c, fullstack.StreamPrepend, "product-list", components.ProductCard(product))
})
```

Subscribe in HTML without custom JavaScript:
```html
<turbo-stream-from src="/sse?topic=products"></turbo-stream-from>
<div id="product-list"></div>
```

### 6. Distributed Tracing & Request Correlation

```go
app.GET("/api/checkout", func(c *ztatic.Context) error {
	// Request ID and W3C Trace ID are automatically available on every request
	reqID   := ztatic.RequestIDFromContext(c)
	traceID := ztatic.TraceIDFromContext(c)

	// Make outbound HTTP requests while automatically propagating traceparent headers
	client := trace.NewClient(nil)
	req, _ := http.NewRequestWithContext(c.Request().Context(), "POST", "https://payment.internal/charge", nil)
	resp, err := client.Do(req)

	return ztatic.OK(c, ztatic.Map{"status": "paid", "trace_id": traceID, "request_id": reqID})
})
```

### 7. Structured Logging with Slog

```go
app.GET("/api/orders/:id", func(c *ztatic.Context) error {
	// Pre-populated with request_id, method, path, and client IP
	logger := ztatic.LogFromContext(c)
	logger.Info("fetching order", "order_id", c.Param("id"))

	// Sensitive values (passwords, tokens) are automatically masked in log records
	logger.Info("user authenticated", "token", "my-secret-token") // Output: token=[REDACTED]

	return ztatic.OK(c, ztatic.Map{"id": c.Param("id"), "status": "shipped"})
})
```

### 8. Authentication & Sessions

```go
// Issue tokens (JWT or AEAD-encrypted)
tokenMgr, _ := ztatic.NewTokenManager(token.DefaultConfig([]byte("32-byte-secret-key-goes-here!")))
app.SetTokenManager(tokenMgr)

// Guard routes with role/scope middleware
api := app.Group("/api", token.TokenAuth(tokenMgr))
api.GET("/admin/metrics", adminHandler, ztatic.RequireRole("admin"))

// Or use browser sessions with flash messages and CSRF protection
store := ztatic.NewMemorySessionStore()
app.UseSession(store)
```

### 9. File Testing Utility (Client, Fixtures, Assertions, and Mocks)

Ztatic provides a zero-dependency testing toolkit (`filetest` and `fileassert`) allowing you to test file uploads, downloads, range streaming, and storage without multipart boilerplate or committing test files to git:

```go
package main_test

import (
	"testing"

	"ztatic-go-framework"
	"ztatic-go-framework/fileassert"
	"ztatic-go-framework/filetest"
	"ztatic-go-framework/upload"
)

func TestAvatarUploadAndDownload(t *testing.T) {
	app := ztatic.NewSecure()
	mockStore := filetest.NewMockStorage()
	mockScan := filetest.NewMockScanner()

	// 1. Mount upload & download routes
	uploader, _ := ztatic.NewUploader(ztatic.UploadConfig{
		AllowedMIMEs: upload.ImageMIMEs(),
		Storage:      mockStore,
		Scanner:      mockScan,
	})

	app.POST("/api/avatar", func(c *ztatic.Context) error {
		fh, err := c.FormFile("avatar")
		if err != nil {
			return err
		}
		processed, err := uploader.ProcessFileHeader(c.Request().Context(), fh)
		if err != nil {
			return err
		}
		return ztatic.OK(c, ztatic.Map{"key": processed.Key})
	})

	app.GET("/api/files/:key", func(c *ztatic.Context) error {
		return ztatic.ServeFile(c, mockStore, c.Param("key"), upload.WithInline())
	})

	client := app.TestClient()

	// 2. Perform Multipart Upload with Synthetic Fixture
	resp := client.NewUpload("/api/avatar").
		Field("user_id", "usr_123").
		Attach("avatar", filetest.PNG("avatar.png", 200, 200)).
		WithBearerToken("auth-token-123").
		Send()

	resp.AssertOK(t).
		AssertContentType(t, "application/json")

	// 3. Storage & Scanner Assertions
	fileassert.SavedCount(t, mockStore, 1)
	fileassert.ScanCount(t, mockScan, 1)

	// 4. Test HTTP 206 Partial Content Range Streaming
	savedKey := mockStore.LastSaveCall().Key
	rangeResp := client.NewDownload("/api/files/" + savedKey).
		WithRange(0, 99).
		Send()

	rangeResp.AssertPartialContent(t, 0, 99, mockStore.LastSaveCall().Size).
		AssertContentLength(t, 100).
		AssertNoSniff(t)

	// 5. Test Fault Injection (Disk Full / Permission Denied)
	mockStore.SimulateDiskFull()
	failResp := client.NewUpload("/api/avatar").
		Attach("avatar", filetest.PNG("another.png", 50, 50)).
		Send()
	failResp.AssertStatus(t, 500)
}
```

---

## Production Build & Deployment

Compile your application into a single static binary:

```bash
ztatic build
```

This runs:
1. `templ generate` to compile templates to Go code.
2. `esbuild` to bundle and minify JavaScript and CSS.
3. Content hash and manifest generation for long-term browser caching.
4. `go build -ldflags="-s -w" -trimpath` to produce the final executable in `bin/server`.

Run the binary on your server:

```bash
APP_ENV=production PORT=80 ./bin/server
```

---

## Air-Gapped & Offline Environments

If you develop behind corporate proxies or in offline sandboxes where access to Go's public checksum database (`sum.golang.org`) is restricted:

```bash
# Bypass checksum server verification
export GOSUMDB=off
go mod tidy
```

The `ztatic new` command automatically seeds `go.sum` to support offline initialization.

---

## Contributing

Contributions are welcome! Please feel free to submit issues, open pull requests, or share feedback on GitHub.

```bash
# Run tests
go test ./...

# Run tests with race detection
go test -race ./...
```

---

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
