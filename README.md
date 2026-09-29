# Ztatic Full-Stack Go Framework

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Security](https://img.shields.io/badge/Security-First-red?style=flat-square&logo=shield)](https://github.com/opeteer/ztatic-go-framework)
[![Architecture](https://img.shields.io/badge/Architecture-HOTW-success?style=flat-square)](https://github.com/opeteer/ztatic-go-framework)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

**Ztatic** is a modern, open-source, full-stack web framework for Go built on top of **Echo v5**. It helps developers craft fast, secure, real-time web applications using the **HOTW (HTML Over The Wire)** paradigm: combining **Templ**, **Hotwire (Turbo)**, **Alpine.js**, native **esbuild** bundling, **OpenAPI 3.0 generation**, and **Realtime Pub/Sub (SSE & WebSockets)**.

Write pure Go and HTML—**zero Node.js or npm required**—and compile your entire application (templates, TypeScript/JS, CSS, database migrations) into a **single, statically-linked binary**.

---

## Why Choose Ztatic?

- **Zero Node.js Dependency**: Built-in Go bindings to `esbuild` transpile TypeScript, bundle JavaScript, and compile CSS in sub-10ms directly in memory.
- **HTML Over The Wire (HOTW)**: Stream reactive HTML updates straight from Go using **Templ** templates and **Hotwire Turbo 8**, keeping client-side state lightweight.
- **Automated REST & OpenAPI 3.0**: Introspects registered routes and Go struct validation tags to generate OpenAPI 3.0.3 documentation rendered interactively via **Scalar UI**, with full support for recursive embedded struct property flattening.
- **Realtime Multi-Node Events**: Push DOM mutations instantly over Server-Sent Events (SSE) or WebSockets using in-memory or distributed **Redis Pub/Sub** brokers.
- **Built-in Security Defaults**: Out-of-the-box WAF with deep payload inspection (URIs and request bodies up to 128KB), multiline XSS protection, HTML event-handler blocking, recursive URL unescaping, SQL comment stripping, nonces-based CSP, Double-Submit CSRF, Argon2id password hashing, and field-level AES-256 encryption.
- **Single Binary Artifact**: Deploy everything—assets, views, database migrations, and backend logic—as a single zero-dependency static executable.

---

## Core Modules & Features

### 1. Rapid REST API & OpenAPI 3.0 Engine (`rapid`)
* **Dynamic Route Introspection:** Introspects registered routes and normalizes Echo path parameter syntax (`/users/:id` → `/users/{id}`).
* **Reflection & Tag Parsing Engine:** Inspects Go struct `json` and `validate` tags (e.g., `required`, `email`, `min`, `max`, `len`) to produce OpenAPI 3.0.3 JSON schemas with cyclic pointer protection and automatic property merging for embedded anonymous structs.
* **Scaffolded Controllers:** `rapid.RegisterResource[T]` automatically mounts type-safe REST CRUD endpoints (`GET`, `POST`, `PUT`, `DELETE`) with pre-registered OpenAPI documentation.
* **Interactive Scalar UI:** Serves dynamic API documentation rendered by Scalar at `/docs` and raw JSON schemas at `/docs/openapi.json`.

### 2. Realtime Engine & Event Brokers (`realtime`)
* **Pub/Sub Brokers:** Includes a local `MemoryBroker` for single-node apps and a distributed `RedisBroker` (`go-redis/v9`) supporting standalone Redis, Sentinel, or Redis Cluster configurations.
* **SSE & WebSocket Transports:** Serves Turbo Stream updates over Server-Sent Events (`<turbo-stream-from src="/sse?topic=room">`) or full-duplex WebSockets (`gorilla/websocket`) with automated ping/pong keep-alives and connection teardown.

### 3. Native Asset Pipeline (`fullstack`)
* **In-Memory esbuild Bundling:** Compiles JS, TS, and CSS on the fly in sub-10ms with detailed terminal error reporting (`AssetBundleError`).
* **Dual-Mode Asset Mounting:** Serves disk-backed assets during development with live reloading, and Go `embed.FS` assets in production with SHA-256 content hashing and long-term caching.

### 4. Data & Persistence Engine (`data`)
* **AST Query Builder:** Integrates `Masterminds/squirrel` for fluid, type-safe SQL query generation with dynamic dialect placeholder resolution (PostgreSQL `$1` vs. MySQL/SQLite `?`).
* **Embedded Migrations:** Embeds `pressly/goose/v3` SQL migrations directly in Go binaries using `embed.FS`.
* **Panic-Safe Transactions:** `DBEngine.Transaction(ctx, fn)` automatically handles `COMMIT` on success, and `ROLLBACK` on explicit errors or runtime panics.

### 5. High-Performance Core Router & TLS Proxy (`echo`)
* **Radix Tree Node Compaction:** `DefaultRouter.Remove` dynamically merges single-child non-handler nodes upon route deletion, preventing tree fragmentation and guaranteeing $O(\text{URL length})$ routing speed.
* **HTTPS/TLS Reverse Proxy:** Proxy middleware supports custom `crypto/tls.Config` settings (`InsecureSkipVerify`, custom CAs) for both HTTP reverse proxying (`proxyHTTP`) and raw WebSocket TLS tunneling (`proxyRaw`).

### 6. Security Suite (`security`)
* **WAF & Security Headers:** Inspects incoming URIs and request payload bodies (up to 128KB) for SQLi and XSS with 413 oversized body enforcement, performs resilient URL unescaping and SQL comment normalization, injects per-request nonce-based CSP (`fullstack.Nonce(c)`), HSTS (`Strict-Transport-Security`), and hardened Double-Submit Cookie CSRF protection (`HttpOnly` with automatic `/api/` and `/docs` skipping).
* **WAF Body Limit:** The WAF enforces a default **128 KB** maximum request body for security inspection. For endpoints that accept larger payloads (document uploads, rich content), call `SetMaxBodySize` before starting the server:
  ```go
  app := ztatic.NewSecure()
  app.SetMaxBodySize(4 * 1024 * 1024) // Allow up to 4 MB
  log.Fatal(app.Start(":8080"))
  ```
* **Data Privacy:** AES-256-GCM database field encryption (`crypto.EncryptedString` and `ztatic:"encrypt"` via `BaseRepository`, initialized with `app.SetCipherKey` or `ZTATIC_CIPHER_KEY`), Argon2id password hashing, and PII scrubbing for `slog`.

### 7. Zero-Trust Audit Logging Engine (`security/audit`)
* **Compliance-Grade Event Ledger:** Captures Actor (ID, Role, Tenant, IP), Target resource, Action verb, Outcome status, Context (Request ID, route, latency), and State Diffs, complying with SOC 2, HIPAA, ISO 27001, and NIST SP 800-92 standards.
* **Pluggable Multi-Format Serialization:** Built-in formatters for **JSON/NDJSON**, **CNCF CloudEvents v1.0.2**, **Micro Focus CEF (Common Event Format)** for SIEM ingestion, and human-readable terminal output with ANSI colors.
* **Flexible Sinks & Schedulers:** Choose between deterministic `SyncLogger` (fail-closed mode for strict financial systems) or high-throughput `AsyncLogger` with non-blocking channel queues, worker pools, and overflow drop/block/fallback policies. Supports `WriterSink` (stdout/stderr), `FileSink` (append-only with fsync), `MemorySink` (testing & admin queries), `MultiSink` (fan-out broadcast), and `WebhookSink`.
* **Zero-Config Middleware & Deep Sanitization:** Automatically audits all state-mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) and error responses (`status >= 400`), while automatically redacting sensitive credentials and PII (passwords, tokens, keys) in metadata and diffs.

### 8. Structured Operational Logging Engine (`log`)
* **Standard-Aligned Architecture:** Built natively on Go standard library `log/slog` with seamless Echo v5 integration.
* **Dual-Mode Formatting:** Emits human-friendly colorized terminal logs with aligned badges in development (`APP_ENV=development`), and high-performance NDJSON in production for automated SIEM / log aggregator ingestion (Datadog, Loki, Splunk).
* **Automatic Request Logging:** Pre-wired into `NewSecure()` to record request latency, route pattern, client IP, HTTP status, and payload sizes with intelligent status-to-level mapping (2xx/3xx -> INFO, 4xx -> WARN, 5xx -> ERROR).
* **Context & Request Correlation:** Automatically generates or propagates `X-Request-ID` across HTTP response headers and context loggers, ensuring every log emitted inside a handler correlates to the request.
* **Dynamic Level Switching:** Dynamically adjust log severity at runtime via `app.SetLogLevel(...)` and `slog.LevelVar` without restarting the server.
* **Zero-Trust Privacy Scrubbing:** Seamlessly integrates with `security/privacy.LogMasker` to scrub passwords, tokens, API keys, and sensitive data from all emitted log records.

### 9. Standardized Error Handling & Resilience (`errors`)
* **Clean Architecture Domain Errors:** Decouples business logic from HTTP transport concerns via `errors.Error` / `ztatic.AppError` (supporting machine codes, user-facing safe messages, internal causes, field violations, and metadata).
* **Automatic HTTP & Runtime Mapping:** Maps domain codes (`NOT_FOUND` -> 404, `VALIDATION_FAILED` -> 422, etc.), standard library errors (`sql.ErrNoRows`, `data.ErrNotFound`, `os.ErrNotExist` -> 404), and `validator.ValidationErrors` into rich structured payloads automatically.
* **Dual Wire Representation:** Supports both standard REST API envelopes (`{"error": {...}}`) and **RFC 9457 / RFC 7807 Problem Details** (`application/problem+json`).
* **Zero-Trust Information Hiding:** In production (`APP_ENV=production`), 5xx internal database errors, SQL queries, and stack traces are scrubbed and replaced with reference-guided safe messages. In development, complete root causes and call stacks are exposed.
* **Centralized Interception & Panic Recovery:** Replaces standard error handlers with a unified pipeline that coordinates with `slog` structured logging and `security/audit` ledgers, while recovering gracefully from runtime panics.

### 10. Developer CLI (`ztatic`)
* **Cobra CLI Suite:** Built on `spf13/cobra` for scaffolding, running, and building applications.
* **Live Reload Engine:** `ztatic dev` monitors `.go`, `.templ`, `.css`, and `.js` files using `fsnotify` with a 100ms debouncer.
* **Single-Binary Compiler:** `ztatic build` executes a 4-step pipeline (`templ generate`, `esbuild`, manifest generation, `go build`) to create an optimized production binary.

### 11. Request/Response Envelope Standardization (`response`)
* **Canonical Type-Safe Envelope:** Generic `Envelope[T any]` encapsulates `success`, `data`, `error`, `pagination`, `meta`, and `links` with zero type assertion overhead.
* **Empty Slice Normalization:** Guarantees empty Go slices serialize as `[]` rather than JSON `null` for consistent collection semantics.
* **Dual-Model Pagination:**
  * **Offset / Page-Based Pagination:** Validates `page`, `per_page`, `sort`, `order` with strict bounds checking (default 20, max 100), automated SQL offset/limit math (`p.Offset()`, `p.Limit()`, `p.Apply(squirrel.SelectBuilder)`), and computed total pages and navigation flags.
  * **Cursor / Keyset-Based Pagination:** Provides opaque URL-safe Base64 token generation and decoding (`EncodeCursor`, `DecodeCursor[T]`) for high-throughput infinite feeds and audit log streaming.
* **HATEOAS Navigation & RFC 5988 Header:** Automatically generates navigation links (`self`, `first`, `prev`, `next`, `last`) preserving all custom query parameters, and injects standard RFC 5988 `Link` HTTP headers.
* **Error Unification & Zero-Trust Safety:** Harmonizes domain errors (`errors.Error`) with the unified envelope contract (`success: false`, `error: {...}`, `meta: {...}`), strictly scrubbing 5xx internals in production while preserving request IDs.
* **Rapid Resource Integration:** `rapid.RegisterResource` natively supports `PaginatedResource[T]` and standard response envelopes (`response.OK`, `response.Created`, `response.Paginated`, `response.NoContent`).
* **Ergonomic DX Helpers:** 1-import top-level helpers: `ztatic.OK(c, data)`, `ztatic.Created(c, data, loc)`, `ztatic.Paginated(c, items, meta)`, `ztatic.NoContent(c)`, `ztatic.ResponseError(c, err)`, and `response.Raw(c, status, data)` escape hatch.

### 12. Zero-Trust Environment Configuration Engine (`config`)
* **Type-Safe Struct Binding:** Binds environment variables into strongly-typed Go structs with `env`, `envDefault`, `envPrefix`, and `envSeparator` struct tags, supporting primitives, slices, `time.Duration`, `time.Time`, `url.URL`, and custom unmarshalers.
* **4-Tier Dotenv Cascading:** Automatically loads configuration cascading across `.env` $\rightarrow$ `.env.local` $\rightarrow$ `.env.{profile}` $\rightarrow$ `.env.{profile}.local` with process `os.Environ()` overriding all files, and nested `${VAR:-default}` variable expansion.
* **Fail-Fast Startup Validation:** Integrates directly with Ztatic's `validation.Engine` (`go-playground/validator/v10`) and zero-trust rules (`config_profile`, `secure_secret`, `https_url`, `db_dsn`, and `SelfValidator`) to reject misconfigured servers in sub-milliseconds with actionable startup diagnostics.
* **Zero-Trust Secrets Protection:** `SecretString` and `Secret[T]` prevent accidental credential leakage by masking sensitive values in `fmt.Printf`, `slog` logs, and JSON serialization (`[REDACTED]`), supporting memory zeroization (`privacy.Zeroize`), Docker/Kubernetes file secret resolution (`file:///run/secrets/...`), and AES-256-GCM encrypted secrets (`enc:aes-gcm:...`).
* **Multi-Environment Profiles:** Built-in profiles (`development`, `test`, `staging`, `production`) automatically detected via `ZTATIC_ENV`, `APP_ENV`, or `GO_ENV` with profile-specific engine defaults.
* **1-Import Ergonomic Helpers:** `ztatic.LoadConfig[T]()`, `ztatic.MustLoadConfig[T]()`, `ztatic.ActiveProfile()`, and `ztatic.NewSecretString()`.

### 13. File Testing Utility Suite (`filetest` & `fileassert`)
* **Fluent Multipart & Download Client:** `app.TestClient()` eliminates `mime/multipart.Writer` boilerplate, supporting `.Attach()`, `.AttachBytes()`, `.AttachReader()`, `.Field()`, `.WithBearerToken()`, `.WithCookie()`, and Range streaming with HTTP 206 Partial Content verification.
* **100% In-Memory Synthetic Fixtures:** Zero repository bloat. Generates valid images with true dimensions (`filetest.PNG`, `filetest.JPEG`, `filetest.GIF`, `filetest.WebP`, `filetest.SVG`), documents (`filetest.PDF`, `filetest.CSV`, `filetest.Text`, `filetest.JSON`), security payloads (`MaliciousSVG`, `WebShell`, `Executable`, `DisguisedExecutable`, `ZipSlip`, `ZipBomb`), virtual zero-RAM oversized streams (`filetest.Oversized`), and managed disk sandboxes (`filetest.TempSandbox`).
* **Expressive Multi-Level Assertions:** First-class standalone assertions (`fileassert.Exists`, `fileassert.ContentEquals`, `fileassert.MalwareBlocked`, `fileassert.PartialContent`) and chained fluent response assertions (`resp.AssertOK()`, `resp.AssertDownloaded()`, `resp.AssertNoSniff()`, `resp.AssertSHA256()`).
* **Thread-Safe Test Doubles & Fault Injection:** `MockStorage` and `MockScanner` provide call spy histories (`WasSaved`, `ScannedFiles`, `SaveCallCount`) and deterministic failure injection (`SimulateDiskFull`, `SimulatePermissionDenied`, `FailClosed` timeout simulations).

---

## Quick Start

### Installation & Prerequisites
Requires **Go 1.21+** and the **templ** CLI:
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

```bash
# Clone the repository
git clone https://github.com/opeteer/ztatic-go-framework.git
cd ztatic-go-framework

# Build the ztatic CLI tool
go build -o ztatic ./cmd/ztatic
```

### 1. Scaffold a New Project
```bash
./ztatic new myapp
cd myapp
```

Project Directory Layout:
```text
myapp/
├── assets/                  # Raw CSS and JavaScript source files
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint
├── db/
│   └── migrations/          # SQL schema migrations (Goose format)
├── internal/
│   ├── controllers/         # HTTP request handlers
│   ├── models/              # Domain models & struct schemas
│   ├── repositories/        # Database access layer
│   └── views/               # Templ templates and layout shells
└── go.mod                   # Go module definition
```

### 2. Basic Server Example (`cmd/server/main.go`)
```go
package main

import (
	"log"
	"os"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
)

func main() {
	// Initialize security-hardened engine
	app := ztatic.NewSecure()

	// Mount static asset pipeline (development live reload vs production)
	isDev := os.Getenv("APP_ENV") == "development"
	fullstack.MountAssets(app.Echo, os.DirFS("dist"), isDev)

	// Expose interactive OpenAPI documentation at /docs
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	// Define application routes
	app.GET("/", func(c *ztatic.Context) error {
		return c.String(200, "Welcome to Ztatic Framework!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Start(":" + port))
}
```

---

## Developer Usage Guide

### 1. Rapid REST API & Interactive OpenAPI Docs

Define a struct model with validation tags:
```go
type Product struct {
	ID    int    `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required,min=3"`
	Price int    `json:"price" validate:"required,min=1"`
}
```

Mount REST CRUD endpoints in one call:
```go
// Automatically registers GET /, GET /:id, POST /, PUT /:id, DELETE /:id
// and generates OpenAPI schemas in /docs/openapi.json
rapid.RegisterResource(app.Group("/api"), "products", productRepo)
```

### 2. Real-Time Updates via SSE & WebSockets

```go
package main

import (
	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
)

func SetupRealtime(app *ztatic.Engine) {
	broker := realtime.NewMemoryBroker()

	// Mount SSE or WebSocket handlers
	app.GET("/sse", realtime.SSEHandler(broker))
	app.GET("/ws", realtime.WebSocketHandler(broker))

	// Broadcast DOM updates from any handler
	app.POST("/messages/send", func(c *ztatic.Context) error {
		broker.Publish(c.Request().Context(), "chat-room", fullstack.TurboStreamItem{
			Action: fullstack.StreamRefresh,
			Target: "chat-box",
		})
		return c.NoContent(200)
	})
}
```

Connect in HTML without writing custom JavaScript:
```html
<turbo-stream-from src="/sse?topic=chat-room"></turbo-stream-from>
<div id="chat-box"></div>
```

### 3. Compliance & Audit Logging (`security/audit`)

Audit logging is pre-wired in `NewSecure()`. State-mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) and error responses (`status >= 400`) are automatically audited. You can enrich the audit ledger directly within request handlers:

```go
package main

import (
	"ztatic-go-framework"
	"ztatic-go-framework/security/audit"
)

func RegisterInvoiceRoutes(app *ztatic.Engine) {
	app.POST("/api/invoices/:id/pay", func(c *ztatic.Context) error {
		invoiceID := c.Param("id")

		// Enrich current request audit entry with target, category, diffs, and metadata
		if entry := ztatic.AuditFromContext(c); entry != nil {
			entry.WithTarget("invoice", invoiceID, "March Subscription").
				WithCategory(audit.CategoryData).
				WithDiff(
					map[string]any{"status": "unpaid", "paid_at": nil},
					map[string]any{"status": "paid", "paid_at": "2026-09-27T02:49:00Z"},
				).
				WithMetadata("gateway", "stripe")
		}

		return c.JSON(200, ztatic.Map{"status": "paid"})
	})
}
```

### 4. Structured Operational Logging & Context Correlation (`log`)

Structured logging is pre-configured and active in `NewSecure()`. Requests automatically receive an `X-Request-ID` and record latency and status codes. Inside handlers, retrieve the enriched request logger to correlate application logs:

```go
package main

import (
	"ztatic-go-framework"
	"ztatic-go-framework/log"
)

func RegisterUserRoutes(app *ztatic.Engine) {
	// Dynamically adjust log level at runtime without restarting
	app.SetLogLevel(log.LevelDebug)

	app.GET("/api/users/:id", func(c *ztatic.Context) error {
		// LogFromContext(c) provides a logger pre-populated with req_id, method, path, and client IP
		logger := ztatic.LogFromContext(c)
		logger.Debug("fetching user record", "user_id", c.Param("id"))

		// Sensitive attributes like passwords or tokens are automatically scrubbed:
		// Output: password=[REDACTED]
		logger.Info("user authenticated", "user_id", c.Param("id"), "password", "my-secret-pass")

		return c.JSON(200, ztatic.Map{"id": c.Param("id"), "status": "active"})
	})
}
```

### 5. Standardized Domain Errors & Struct Validation (`errors`)

Ztatic automatically catches handler errors, panics, and validation errors, transforming them into standardized API responses with correlated request IDs:

```go
package main

import (
	"ztatic-go-framework"
	"ztatic-go-framework/rapid"
)

type CreateUserInput struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
}

func RegisterAccountRoutes(app *ztatic.Engine) {
	// 1. Struct validation errors automatically map to HTTP 422 with granular field details
	app.POST("/api/users", func(c *ztatic.Context) error {
		var input CreateUserInput
		if err := rapid.BindAndValidate(c, &input); err != nil {
			return err // Automatically serialized into standardized validation error JSON
		}
		return c.JSON(201, ztatic.Map{"status": "created"})
	})

	// 2. Clean Architecture: return domain errors without importing net/http
	app.GET("/api/accounts/:id", func(c *ztatic.Context) error {
		account, err := findAccount(c.Param("id"))
		if err != nil {
			return ztatic.ErrNotFound("account does not exist").
				WithMetadata("account_id", c.Param("id"))
		}
		return c.JSON(200, account)
	})
}
```

Standardized API Error Output (HTTP 422):
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed on 1 field(s)",
    "status": 422,
    "details": [
      {
        "field": "Email",
        "rule": "email",
        "message": "Field 'Email' must be a valid email address",
        "value": "invalid-address"
      }
    ],
    "request_id": "req-18d9110b48fd8d58-c28ef1ad-0001",
    "timestamp": "2026-09-27T10:52:00Z"
  }
}
```

### 6. Secure Token & Session Abstraction (`security/token` & `security/session`)

Ztatic provides a unified identity and authorization system designed for both stateful HOTW web applications and stateless REST/real-time APIs:

#### Cryptographic Tokens (JWT, AEAD-Encrypted, and Refresh Tokens)
```go
package main

import (
	"ztatic-go-framework"
	"ztatic-go-framework/security/token"
)

func SetupAuth(app *ztatic.Engine) {
	// Initialize Token Manager with RFC 8725 hardened cryptographic defaults
	secret := []byte(os.Getenv("AUTH_SECRET")) // Minimum 32 bytes (256 bits)
	tokenMgr, _ := ztatic.NewTokenManager(token.DefaultConfig(secret))
	app.SetTokenManager(tokenMgr)

	// Issue token pair (access token + refresh token with single-use rotation)
	app.POST("/api/login", func(c *ztatic.Context) error {
		pair, err := tokenMgr.CreateTokenPair(c.Request().Context(), &ztatic.Claims{
			StandardClaims: ztatic.StandardClaims{Subject: "user-123"},
			Roles:          []string{"admin"},
			Scopes:         []string{"read:users", "write:users"},
			TenantID:       "tenant-alpha",
		})
		if err != nil {
			return ztatic.ErrInternal("Failed to generate token pair")
		}
		return c.JSON(200, pair)
	})

	// Guard endpoints using declarative role and scope middleware
	api := app.Group("/api", token.TokenAuth(tokenMgr))
	api.GET("/admin/stats", func(c *ztatic.Context) error {
		claims := ztatic.ClaimsFromContext(c)
		return c.JSON(200, ztatic.Map{"admin": claims.Subject})
	}, ztatic.RequireRole("admin"), ztatic.RequireScope("read:users"))
}
```

#### Browser Sessions (Stateful & Stateless Encrypted)
```go
package main

import (
	"ztatic-go-framework"
)

func SetupSessions(app *ztatic.Engine) {
	// Choose MemoryStore, RedisStore, or stateless AES-256-GCM CookieStore
	store := ztatic.NewMemorySessionStore()
	app.UseSession(store)

	// Login and prevent Session Fixation attacks
	app.POST("/login", func(c *ztatic.Context) error {
		sess := ztatic.SessionFromContext(c)
		_ = sess.RegenerateID() // Issues new session ID, purges old ID from backend
		sess.Set("user_id", "user-123")
		sess.Flash("notice", "You have successfully signed in")
		return c.Redirect(303, "/dashboard")
	})

	// Read session and consume flash messages
	app.GET("/dashboard", func(c *ztatic.Context) error {
		sess := ztatic.SessionFromContext(c)
		flashes := sess.Flashes("notice")
		return c.Render(200, "dashboard.html", ztatic.Map{
			"user_id": sess.GetString("user_id"),
			"flashes": flashes,
		})
	})
}
```

### 7. Request/Response Envelope Standardization & Pagination (`response`)

Ztatic provides ergonomic helpers for returning standard API envelopes, paginated collections, and HATEOAS navigation links:

```go
package main

import (
	"ztatic-go-framework"
	"ztatic-go-framework/response"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price float64 `json:"price"`
}

func RegisterProductRoutes(app *ztatic.Engine) {
	// 1. Single Item Envelopes: OK (200), Created (201), NoContent (204)
	app.GET("/api/products/:id", func(c *ztatic.Context) error {
		product := Product{ID: 1, Name: "Mechanical Keyboard", Price: 129.99}
		return ztatic.OK(c, product, response.WithMeta("cached", true))
	})

	app.POST("/api/products", func(c *ztatic.Context) error {
		created := Product{ID: 2, Name: "Wireless Mouse", Price: 79.99}
		return ztatic.Created(c, created, "/api/products/2")
	})

	// 2. Offset / Page-Based Pagination with HATEOAS Navigation Links
	app.GET("/api/products", func(c *ztatic.Context) error {
		// Extracts and validates ?page=1&per_page=20&sort=price&order=desc
		p := ztatic.ExtractPagination(c)

		// Directly apply pagination to SQL query builders
		// builder := p.Apply(squirrel.Select("*").From("products"))
		products, totalCount := fetchProducts(p.Limit(), p.Offset())

		// Calculates TotalPages, HasNext, HasPrev, and builds RFC 5988 Link headers
		meta := p.WithTotal(totalCount)
		return ztatic.Paginated(c, products, meta)
	})

	// 3. Raw JSON Escape Hatch (for third-party webhooks)
	app.GET("/api/webhook-health", func(c *ztatic.Context) error {
		return response.Raw(c, 200, map[string]string{"status": "UP"})
	})
}
```

Standardized Paginated Output (HTTP 200):
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
    "request_id": "req-18d91807c92fdff4-435a2085-0007",
    "timestamp": "2026-09-27T12:00:00Z",
    "duration": "1.25ms"
  }
}
```

### 8. File Testing Utility (Client, Fixtures, Assertions, and Mocks)

Ztatic provides a zero-dependency testing toolkit (`filetest` and `fileassert`) allowing developers to test complex file upload, download, malware scanning, and storage interactions without manual multipart boilerplate or committing binary test files to git:

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

## Development & Production Workflow

### Development Mode (Live Reload)
```bash
./ztatic dev
```
Monitors `.go`, `.templ`, `.css`, and `.js` files, auto-compiling templates, bundling assets via `esbuild`, and live-reloading the application.

### Production Build
```bash
./ztatic build
```
Executes the production build pipeline:
1. Compiles `.templ` files into Go code.
2. Bundles and minifies JS/CSS targeting ES2022 via `esbuild`.
3. Generates SHA-256 asset content hashes and `manifest.json`.
4. Compiles a zero-dependency static binary in `bin/server`.

---

## Air-Gapped & Corporate Environments

If you are developing in strict corporate, offline, or sandbox environments where outbound network access to public proxies is restricted:
- The Ztatic CLI (`ztatic new`) automatically seeds the project's `go.sum` and locks dependencies so initial generation and compilation work 100% offline.
- If you need to run `go mod tidy` or `go build` without contacting Go's public checksum database (`sum.golang.org`), configure Go to bypass sumdb lookups:
```bash
export GOSUMDB=off
```
Or set it per command:
```bash
GOSUMDB=off go mod tidy
```

---

## License

Distributed under the MIT License. See `LICENSE` for details.
