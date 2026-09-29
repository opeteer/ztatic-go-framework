# Ztatic Framework Tutorial: Building High-Performance Web Applications in Go

Welcome to the step-by-step tutorial for building modern, fast, full-stack web applications using the **Ztatic Framework**.

Ztatic is an open-source, full-stack Go framework built on top of **Echo v5**. It enables you to craft reactive, real-time web applications using the **HOTW (HTML Over The Wire)** stack—combining **Templ**, **Hotwire (Turbo 8)**, **Alpine.js**, native in-memory **esbuild** bundling, **OpenAPI 3.0 generation**, and **Realtime Pub/Sub (SSE & WebSockets)**—all without requiring Node.js, npm, or frontend build tools.

Under the hood, Ztatic provides everything you need to build real-world applications: type-safe cascading configuration, W3C distributed tracing, structured logging, activity audit trails, standardized error handling, clean response envelopes, and a fluid data tier powered by reflection mapping, Squirrel AST queries, Goose migrations, and generic repositories.

---

## Table of Contents

1. [Understanding the Ztatic Architecture](#1-understanding-the-ztatic-architecture)
2. [Ztatic Framework Directory & Module Structure](#2-ztatic-framework-directory--module-structure)
3. [Step 1: Environment Setup & Installing the `ztatic` CLI](#step-1-environment-setup--installing-the-ztatic-cli)
4. [Step 2: Scaffolding a New Application](#step-2-scaffolding-a-new-application)
5. [Step 3: Configuration & Environment Management (`config`)](#step-3-configuration--environment-management-config)
6. [Step 4: Database & Data Tier Setup (Engine, Mapper, Migrations & Transactions)](#step-4-database--data-tier-setup-engine-mapper-migrations--transactions)
7. [Step 5: Crafting Views (Layouts, Templ Components & Micro-Interactions)](#step-5-crafting-views-layouts-templ-components--micro-interactions)
8. [Step 6: Implementing Rapid REST APIs, Input Sanitization & OpenAPI Docs](#step-6-implementing-rapid-rest-apis-input-sanitization--openapi-docs)
9. [Step 7: Adding Real-Time Updates (SSE, WebSockets & Pub/Sub)](#step-7-adding-real-time-updates-sse-websockets--pubsub)
10. [Step 8: Observability: Tracing, Logging & Activity Ledger](#step-8-observability-tracing-logging--activity-ledger)
11. [Step 9: Asset Management & Client Integration](#step-9-asset-management--client-integration)
12. [Step 10: Assembling the Full Application Server (`cmd/server/main.go`)](#step-10-assembling-the-full-application-server-cmdservermaingo)
13. [Step 11: Development Workflow (Live Reload)](#step-11-development-workflow-live-reload)
14. [Step 12: Production Build & Single-Binary Deployment](#step-12-production-build--single-binary-deployment)
15. [Summary Checklist](#summary-checklist)

---

## 1. Understanding the Ztatic Architecture

Before writing code, let's look at how Ztatic operates under the hood.

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

### Key Technical Pillars

1. **Zero Node.js Overhead**: Native Go bindings run `esbuild` directly in memory to transpile TypeScript, bundle JavaScript, and compile CSS in under 10 milliseconds without requiring npm or Node.js runtimes.
2. **HTML Over The Wire (HOTW)**: Instead of sending heavy JSON payloads and managing complex client-side state, Ztatic renders type-safe Go components using **Templ** and streams DOM mutations over HTTP using **Hotwire Turbo 8**.
3. **Smart Layout Unwrapping**: Inside `<turbo-frame>` containers, Ztatic automatically detects frame request headers (`Turbo-Frame`) and bypasses outer HTML layout rendering, saving bandwidth and improving responsiveness.
4. **Real-Time Push without JS Boilerplate**: By linking Server-Sent Events (SSE) or WebSockets directly to Hotwire Turbo Streams (`<turbo-stream-from src="/sse?topic=room">`), the backend mutates client DOM elements instantly without custom frontend JavaScript.
5. **Type-Safe Configuration**: Strongly-typed struct configuration with 4-tier dotenv cascading (`.env` $\rightarrow$ `.env.local` $\rightarrow$ `.env.{profile}` $\rightarrow$ `.env.{profile}.local`), nested variable expansion, and `SecretString` to prevent accidental credential leakage in logs.
6. **W3C Distributed Tracing & Request Correlation**: Conforms to W3C Trace Context specifications (`traceparent`, `tracestate`), sanitizes incoming correlation headers (`X-Request-ID`), binds trace IDs to the request context, and provides outbound HTTP transport propagation.
7. **Modern Data Tier**: High-performance persistence featuring `data.NewDBEngine` with automatic dialect placeholder resolution (PostgreSQL `$` vs. MySQL/SQLite `?`), reflection mapping with struct tags (`db`, `ztatic:"encrypt"`), fluent Squirrel AST query helpers, nested transactions with Savepoints, and a generic `BaseRepository[T]` implementing standard REST and paginated operations.
8. **Automated REST APIs & Interactive OpenAPI 3.0**: Dynamic route introspection and reflection tag parsing produce OpenAPI 3.0.3 specifications rendered interactively via **Scalar UI** at `/docs`, with recursive struct property merging.
9. **Standardized Responses & Domain Errors**: Canonical `Envelope[T]` contracts guarantee type safety and normalize empty slices to `[]` instead of `null`. Clean architecture domain errors map to RFC 9457 Problem Details with safe error messages in production.
10. **Built-in Security Defaults**: Out-of-the-box WAF payload inspection (URIs and request bodies up to 128KB), multiline XSS filtering, per-request nonce-based Content Security Policy (`fullstack.Nonce(c)`), HSTS, hardened CSRF (`HttpOnly` with automatic `/api/` and `/docs` skipping), Argon2id password hashing, and AES-256-GCM database field encryption (`ztatic:"encrypt"`).

---

## 2. Ztatic Framework Directory & Module Structure

When you scaffold a Ztatic project, you work with a standard architectural layout designed for clarity and separation of concerns.

### Standard Scaffolding Structure (`ztatic new`)

```text
mywebsite/
├── .env                     # Local environment configuration
├── .env.example             # Configuration blueprint template
├── assets/                  # Frontend raw source files
│   ├── css/                 # Global stylesheets & CSS source files
│   └── js/                  # Alpine.js modules, controllers, TypeScript/JS source
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint & HTTP server bootstrapping
├── db/
│   └── migrations/          # Embedded SQL schema migrations (Goose format)
├── dist/                    # Bundled, minified, content-hashed static output
│   └── .gitkeep
├── dist.go                  # Root package embed.FS for single-binary distribution
├── go.mod                   # Go module definition
├── internal/
│   ├── config/              # Type-safe environment configuration schema
│   │   └── config.go
│   ├── controllers/         # HTTP Route Handlers & Controller logic
│   ├── models/              # Business domain structs & database schemas
│   ├── repositories/        # Data access layer (Generic BaseRepository & Squirrel)
│   └── views/               # Type-safe Templ view templates
│       ├── components/      # Modular, reusable UI components
│       └── layouts/         # Master layout wrappers (HTML head, shell, nav)
└── tools.go                 # Go toolchain runtime dependencies & generator locks
```

### Core Framework Modules Breakdown

| Module Package | Path | Responsibilities |
| :--- | :--- | :--- |
| **`ztatic`** | [`ztatic.go`](file:///home/opeteer/ztatic-go-framework/ztatic.go) | Central engine constructor (`NewSecure()`), exporting convenience type aliases (`Context`, `HandlerFunc`, `Map`, `Group`), top-level DX helpers (`OK`, `Created`, `Paginated`, `MustLoadConfig`), and pre-wiring the full security and observability pipeline onto Echo v5. |
| **`config`** | [`config/`](file:///home/opeteer/ztatic-go-framework/config) | Configuration management: 4-tier dotenv cascading, profiles (`development`, `test`, `staging`, `production`), struct tag binding (`env`, `envDefault`), fail-fast validation, and `SecretString` / `Secret[T]` wrappers. |
| **`trace`** | [`trace/`](file:///home/opeteer/ztatic-go-framework/trace) | W3C distributed tracing (`traceparent`, `tracestate`), URL-safe timestamped request ID generation (`GenerateRequestID`), request ID sanitization, Echo tracing middleware, context accessors, and outbound HTTP `Transport` / `Client` propagation. |
| **`data`** | [`data/`](file:///home/opeteer/ztatic-go-framework/data) | Database pool wrapper (`DBEngine`), reflection struct mapper (`ScanOne`, `ScanAll`, `ExtractValues`), fluent Squirrel AST query helpers, Goose embedded SQL migrations (`RunMigrations`), nested transactions with savepoints (`TransactionCtx`), and generic `BaseRepository[T]`. |
| **`validation`** | [`validation/`](file:///home/opeteer/ztatic-go-framework/validation) | Input sanitization engine (`sanitize` tags), built-in security rules (`strong_password`, `xss_safe`, `no_sql_injection`, `safe_path`), domain validators (`slug`, `phone`, `unique`, `exists`), custom error messages (`message` tags), and `rapid.BindAndValidate`. |
| **`response`** | [`response/`](file:///home/opeteer/ztatic-go-framework/response) | Canonical type-safe envelope (`Envelope[T]`), empty slice normalization (`[]`), offset and cursor pagination, HATEOAS link generator, RFC 5988 `Link` HTTP header injection, and raw JSON escape hatches. |
| **`errors`** | [`errors/`](file:///home/opeteer/ztatic-go-framework/errors) | Domain application errors (`errors.Error` / `ztatic.AppError`), RFC 9457 Problem Details formatting, automatic HTTP status code mapping, and safe error messages in production. |
| **`log`** | [`log/`](file:///home/opeteer/ztatic-go-framework/log) | Standard Go `log/slog` structured logging, dual-mode formatting (colorized terminal in development, NDJSON in production), request logging middleware with latency and status recording, dynamic log level switching, and sensitive data masking. |
| **`security`** | [`security/`](file:///home/opeteer/ztatic-go-framework/security) | Web security: WAF request inspection, nonces-based CSP (`fullstack.Nonce(c)`), hardened Double-Submit CSRF, Argon2id hashing, AES-256-GCM field encryption, activity audit logging (`security/audit`), JWT/AEAD tokens (`security/token`), and browser sessions (`security/session`). |
| **`fullstack`** | [`fullstack/`](file:///home/opeteer/ztatic-go-framework/fullstack) | Layout rendering (`RenderLayout`), Turbo Frame detection (`IsTurboFrame`), Turbo Stream responses (`RenderTurboStream`), asset pipeline (`MountAssets`, `esbuild`), and content hashing manifest. |
| **`realtime`** | [`realtime/`](file:///home/opeteer/ztatic-go-framework/realtime) | Pub/Sub messaging: local `MemoryBroker` for single instances, distributed `RedisBroker` for multi-node setups, Server-Sent Events (`SSEHandler`), and full-duplex WebSockets (`WebSocketHandler`). |
| **`rapid`** | [`rapid/`](file:///home/opeteer/ztatic-go-framework/rapid) | Dynamic RESTful route registration (`RegisterResource`), struct validation reflection, automated OpenAPI 3.0.3 JSON schema generation, and interactive Scalar UI docs rendering at `/docs`. |
| **`echo`** | [`echo/`](file:///home/opeteer/ztatic-go-framework/echo) | Core router with radix tree node compaction (`Remove`), fast-path query binding, and HTTPS/TLS reverse proxying (`proxyHTTP`, `proxyRaw`). |
| **`cmd/ztatic`** | [`cmd/ztatic/`](file:///home/opeteer/ztatic-go-framework/cmd/ztatic) | Developer CLI: project scaffolding (`ztatic new`), live-reloading dev server with debounced watching (`ztatic dev`), and single-binary production compiler (`ztatic build`). |

---

## Step 1: Environment Setup & Installing the `ztatic` CLI

### Prerequisites
- **Go 1.21+** installed on your operating system.
- **templ CLI** installed globally:
  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```

### Installing the Ztatic CLI

Clone the repository and compile the unified `ztatic` CLI binary:

```bash
# Clone the repository
git clone https://github.com/opeteer/ztatic-go-framework.git
cd ztatic-go-framework

# Compile the ztatic CLI binary
go build -o ztatic ./cmd/ztatic

# Optionally move to your PATH for global usage
sudo mv ztatic /usr/local/bin/
```

Verify installation:

```bash
ztatic --help
```

> [!TIP]
> **Offline / Corporate Environments:** If you develop behind strict proxies or offline networks where access to Go's checksum server (`sum.golang.org`) is restricted, configure Go to bypass sumdb checks:
> ```bash
> export GOSUMDB=off
> ```

---

## Step 2: Scaffolding a New Application

Create a new application named `mywebsite`:

```bash
ztatic new mywebsite
cd mywebsite
```

### What `ztatic new` Generates:
1. **Directory Tree**: Standard folders for assets, controllers, models, views, repositories, and migrations.
2. **Configuration Blueprint**: Pre-configured `.env` and `.env.example` files containing defaults for ports, application names, and secrets.
3. **Type-Safe Config Model**: `internal/config/config.go` with environment struct binding.
4. **Single-Binary Root Embed**: `dist.go` at the root package exposing `//go:embed all:dist` as `mywebsite.DistFS`.
5. **Toolchain Locks**: `tools.go` pre-wiring Templ and runtime dependencies.
6. **Application Entrypoint**: `cmd/server/main.go` ready for development and deployment.

---

## Step 3: Configuration & Environment Management (`config`)

Applications should never hardcode configuration values or leak credentials. Ztatic provides a clean configuration engine built into `ztatic-go-framework/config`.

### 1. Define the Configuration Schema (`internal/config/config.go`)

Open `internal/config/config.go` and define your configuration struct:

```go
package config

import (
	"ztatic-go-framework"
)

// AppConfig defines the application's environment configuration schema.
type AppConfig struct {
	AppName     string              `env:"APP_NAME" envDefault:"mywebsite" validate:"required"`
	Port        string              `env:"PORT" envDefault:"8080" validate:"required"`
	DatabaseDSN string              `env:"DATABASE_DSN" envDefault:"app.db" validate:"required"`
	CipherKey   ztatic.SecretString `env:"ZTATIC_CIPHER_KEY" envDefault:"01234567890123456789012345678901"`
	AuthSecret  ztatic.SecretString `env:"AUTH_SECRET" envDefault:"dev-secret-key-must-be-changed-in-production-min-32-bytes" validate:"required,min=32"`
}
```

### 2. Configure Local Environment Variables (`.env`)

Edit `.env` in your project root:

```ini
APP_NAME=mywebsite
PORT=8080
DATABASE_DSN=app.db
ZTATIC_CIPHER_KEY=01234567890123456789012345678901
AUTH_SECRET=dev-secret-key-must-be-changed-in-production-min-32-bytes
```

### 3. How Configuration Works in Ztatic

- **Dotenv Cascading**: Loads configuration in cascading order:
  $$\text{.env} \longrightarrow \text{.env.local} \longrightarrow \text{.env.\{profile\}} \longrightarrow \text{.env.\{profile\}.local}$$
  Process `os.Environ()` variables always take final precedence.
- **Nested Variable Expansion**: Supports shell syntax such as `${DATABASE_PATH:-data/app.db}`.
- **Safe `SecretString` & `Secret[T]`**:
  - Automatically serializes as `"[REDACTED]"` when printed via `fmt.Printf`, logged with `slog`, or serialized to JSON.
  - Access the raw plaintext only when intentionally passing to database drivers or crypto modules using `secret.Expose()` or `secret.Value()`.
  - Securely wipe secrets from memory when discarded using `secret.Destroy()`.
  - Supports Docker/Kubernetes volume secrets via `file:///run/secrets/api_key` or encrypted values via `enc:aes-gcm:<ciphertext>`.
- **Profiles**: Automatically detected via `ZTATIC_ENV`, `APP_ENV`, or `GO_ENV` (`development`, `test`, `staging`, `production`).
- **Fail-Fast Validation**: Loading configuration via `ztatic.MustLoadConfig[AppConfig]()` validates every struct tag at startup, refusing to boot if mandatory variables are missing.

---

## Step 4: Database & Data Tier Setup (Engine, Mapper, Migrations & Transactions)

Ztatic provides a unified persistence layer via `ztatic-go-framework/data`.

### 1. Define the Domain Model (`internal/models/article.go`)

Create `internal/models/article.go` with database mapping and validation tags:

```go
package models

import "time"

type Article struct {
	ID          int       `json:"id" db:"id,primarykey,autoincrement"`
	Title       string    `json:"title" db:"title" validate:"required,min=3,max=100,xss_safe" message:"Title is required and must be between 3 and 100 characters"`
	Content     string    `json:"content" db:"content" validate:"required"`
	Author      string    `json:"author" db:"author" validate:"required"`
	SecretNotes string    `json:"secret_notes,omitempty" db:"secret_notes" ztatic:"encrypt"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

> [!TIP]
> **Field-Level Encryption:** Fields tagged with `ztatic:"encrypt"` (such as `SecretNotes`) are automatically encrypted before writing to the database and decrypted upon scanning by `BaseRepository[T]`. This uses the 32-byte key set via `app.SetCipherKey(...)` or the `ZTATIC_CIPHER_KEY` environment variable.

### 2. Embedded SQL Migrations (`db/migrations/00001_create_articles_table.sql`)

Create `db/migrations/00001_create_articles_table.sql`:

```sql
-- +goose Up
CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    secret_notes TEXT,
    created_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE articles;
```

Create `db/migrations/embed.go` to bundle migrations directly into your binary:

```go
package migrations

import "embed"

// MigrationFS embeds all SQL migration files for single-binary deployment
//go:embed *.sql
var MigrationFS embed.FS
```

### 3. Implement Generic Repository with AST Query Helpers (`internal/repositories/article_repository.go`)

Leverage Ztatic's generic `BaseRepository[T]`:

```go
package repositories

import (
	"context"

	"ztatic-go-framework/data"
	"mywebsite/internal/models"
)

type ArticleRepository struct {
	*data.BaseRepository[models.Article]
}

func NewArticleRepository(db *data.DBEngine) *ArticleRepository {
	return &ArticleRepository{
		BaseRepository: data.NewBaseRepository[models.Article](db, "articles"),
	}
}

// Custom query using Ztatic fluent query helpers and Squirrel AST
func (r *ArticleRepository) FindByAuthor(ctx context.Context, author string) ([]models.Article, error) {
	return r.Find(ctx,
		data.WhereEq("author", author),
		data.OrderByDesc("created_at"),
	)
}

// Search articles by title keyword
func (r *ArticleRepository) Search(ctx context.Context, term string) ([]models.Article, error) {
	return r.Find(ctx,
		data.WhereLike("title", "%"+term+"%"),
		data.OrderByDesc("created_at"),
		data.Limit(50),
	)
}
```

### 4. Built-in Capabilities of `BaseRepository[T]`

Ztatic's `BaseRepository[T]` eliminates boilerplate data access code. It out-of-the-box implements:

- **Full REST & Pagination Contracts**: Implements `rapid.Resource[T]` (`FindAll`, `FindByID`, `Create`, `Update`, `Delete`) and `rapid.PaginatedResource[T]` (`FindAllPaginated`).
- **CRUD Operations**:
  - `r.GetByID(ctx, id)`: Fetches a single record by primary key with automatic struct mapping and field decryption.
  - `r.Find(ctx, opts...)`: Fetches records using fluent query options (`WhereEq`, `WhereIn`, `WhereBetween`, `OrderByDesc`, `Limit`, etc.).
  - `r.FindPaginated(ctx, pageParams, opts...)`: Executes count and select queries, returning results with pagination metadata.
  - `r.FindCursor(ctx, cursorParams, cursorCol, opts...)`: Executes keyset/cursor queries for infinite feeds.
  - `r.Insert(ctx, entity)`: Inserts a record and populates auto-increment primary keys back onto the struct.
  - `r.InsertMany(ctx, entities)`: Batch multi-row insert.
  - `r.Save(ctx, entity)`: Updates a record identified by its primary key.
  - `r.UpdateColumns(ctx, id, valuesMap)`: Updates specific columns without re-writing the full entity.
  - `r.DeleteByID(ctx, id)` / `r.DeleteWhere(ctx, pred, args...)`: Deletes matching records.
  - `r.Count(ctx, opts...)` / `r.Exists(ctx, opts...)`: Efficient count and existence checks.
- **Transactions & Savepoints**:
  ```go
  err := dbEngine.TransactionCtx(ctx, func(txCtx context.Context) error {
      // Any repository method passed txCtx automatically participates in the transaction!
      if err := repo.Insert(txCtx, &article1); err != nil {
          return err // Automatic rollback
      }
      return repo.Insert(txCtx, &article2) // Automatic commit
  })
  ```
  If `TransactionCtx` is invoked when a transaction already exists on `ctx`, Ztatic automatically creates an internal SQL **Savepoint** (`SAVEPOINT sp_ztatic_N`) for safe nested transactions.

> [!TIP]
> When using SQLite with Go `database/sql`, import a pure-Go driver in your application entrypoint:
> ```go
> import _ "modernc.org/sqlite"
> ```

---

## Step 5: Crafting Views (Layouts, Templ Components & Micro-Interactions)

Ztatic uses **Templ** for type-safe, compiled HTML templates in Go.

### 1. Main Layout Shell (`internal/views/layouts/app_layout.templ`)

Define the master layout shell in `internal/views/layouts/app_layout.templ`:

```html
package layouts

import "ztatic-go-framework/fullstack"

templ AppLayout(content templ.Component) {
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
		<title>My Ztatic Website</title>
		
		<!-- Loaded from native esbuild pipeline with content hash -->
		<link rel="stylesheet" href={ fullstack.AssetURL("app.css") }/>
		<script src="https://cdn.jsdelivr.net/npm/@hotwired/turbo@8.0.0-beta.2/dist/turbo.es2017-umd.js"></script>
		<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
	</head>
	<body class="bg-gray-100 text-gray-900 font-sans">
		<header class="bg-blue-600 text-white p-4 shadow-md">
			<div class="container mx-auto flex justify-between items-center">
				<h1 class="text-xl font-bold">Ztatic Site</h1>
				<nav class="space-x-4">
					<a href="/" class="hover:underline">Home</a>
					<a href="/docs" class="hover:underline">API Docs</a>
				</nav>
			</div>
		</header>

		<main class="container mx-auto my-8 p-4 bg-white rounded shadow-sm">
			@content
		</main>

		<footer class="text-center p-4 text-gray-500 text-sm">
			Built with Ztatic Framework & HOTW Stack
		</footer>
	</body>
	</html>
}
```

> [!NOTE]
> **Smart Layout Unwrapping:** When navigating inside `<turbo-frame>` containers, `fullstack.RenderLayout` automatically inspects the `Turbo-Frame` request header and strips the outer layout shell, transmitting only the inner component over the wire.

### 2. Article Components & Alpine.js Modal (`internal/views/components/article_card.templ`)

Create `internal/views/components/article_card.templ`:

```html
package components

import (
	"fmt"
	"mywebsite/internal/models"
)

templ ArticleCard(article models.Article) {
	<div id={ fmt.Sprintf("article-%d", article.ID) } class="p-4 border-b border-gray-200 hover:bg-gray-50">
		<h2 class="text-lg font-semibold text-blue-700">{ article.Title }</h2>
		<p class="text-gray-600 mt-1">{ article.Content }</p>
		<div class="text-xs text-gray-400 mt-2">
			By { article.Author } on { article.CreatedAt.Format("Jan 02, 2006") }
		</div>
	</div>
}

templ CreateArticleModal() {
	<div x-data="{ open: false }">
		<button @click="open = true" class="bg-blue-600 text-white px-4 py-2 rounded shadow hover:bg-blue-700 transition">
			+ New Article
		</button>

		<div x-show="open" x-cloak class="fixed inset-0 bg-black/50 flex items-center justify-center">
			<div @click.away="open = false" class="bg-white p-6 rounded-lg w-96 shadow-xl">
				<h3 class="text-lg font-bold mb-4">Create New Article</h3>
				<form action="/articles" method="POST" @submit="open = false">
					<input type="text" name="title" placeholder="Title" required class="w-full mb-3 p-2 border rounded"/>
					<textarea name="content" placeholder="Content" required class="w-full mb-3 p-2 border rounded"></textarea>
					<input type="text" name="author" placeholder="Author Name" class="w-full mb-3 p-2 border rounded"/>
					<input type="text" name="secret_notes" placeholder="Private Notes (Encrypted at Rest)" class="w-full mb-3 p-2 border rounded"/>
					<div class="flex justify-end space-x-2">
						<button type="button" @click="open = false" class="px-4 py-2 border rounded">Cancel</button>
						<button type="submit" class="px-4 py-2 bg-blue-600 text-white rounded">Post</button>
					</div>
				</form>
			</div>
		</div>
	</div>
}

templ ArticleList(articles []models.Article) {
	<div>
		<div class="flex justify-between items-center mb-4">
			<h1 class="text-2xl font-bold">Latest Articles</h1>
			@CreateArticleModal()
		</div>

		<!-- Turbo Stream SSE Listener for Real-Time Updates -->
		<turbo-stream-from src="/sse?topic=articles"></turbo-stream-from>

		<div id="articles-container" class="divide-y divide-gray-200">
			for _, article := range articles {
				@ArticleCard(article)
			}
		</div>
	</div>
}
```

---

## Step 6: Implementing Rapid REST APIs, Input Sanitization & OpenAPI Docs

Ztatic makes building REST APIs and generating documentation straightforward.

### 1. Mounting REST Endpoints (`rapid.RegisterResource`)

In `cmd/server/main.go`, register your repository directly as a REST resource:

```go
// Automatically mounts:
// - GET    /api/articles          (offset pagination, sort, RFC 5988 Link headers)
// - GET    /api/articles/:id      (single item envelope)
// - POST   /api/articles          (validation & creation)
// - PUT    /api/articles/:id      (validation & update)
// - DELETE /api/articles/:id      (deletion)
rapid.RegisterResource(app.Group("/api"), "articles", articleRepo)

// Serve interactive Scalar API documentation UI at /docs
rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")
```

### 2. Standardized Response Envelopes & Pagination

REST endpoints registered through `rapid` automatically output clean response envelopes:

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Getting Started with Ztatic",
      "content": "Building high performance apps in Go.",
      "author": "Alice",
      "created_at": "2026-09-29T10:00:00Z"
    }
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
    "self": "/api/articles?page=1&per_page=20",
    "first": "/api/articles?page=1&per_page=20",
    "next": "/api/articles?page=2&per_page=20",
    "last": "/api/articles?page=3&per_page=20"
  },
  "meta": {
    "request_id": "req-1923e5904b78-b1a4-0001",
    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
    "timestamp": "2026-09-29T10:00:01Z",
    "duration": "1.42ms"
  }
}
```

For custom endpoints, use Ztatic's top-level DX helpers:
- `ztatic.OK(c, data)`
- `ztatic.Created(c, data, "/api/articles/42")`
- `ztatic.Paginated(c, items, meta)`
- `ztatic.NoContent(c)`
- `ztatic.ResponseError(c, err)`

### 3. Interactive Scalar API Documentation

Navigate to `http://localhost:8080/docs` in your browser. The Scalar documentation UI dynamically generates interactive documentation directly from your Go struct tags (`json`, `validate`, `db`) with live "Try It" capabilities and zero manual YAML maintenance.

---

## Step 7: Adding Real-Time Updates (SSE, WebSockets & Pub/Sub)

Ztatic natively pushes DOM mutations over Server-Sent Events (SSE) and WebSockets using in-memory or Redis Pub/Sub brokers.

### 1. Choosing an Event Broker (`MemoryBroker` vs. `RedisBroker`)

- **`MemoryBroker`**: In-memory broker ideal for single-instance setups (`realtime.NewMemoryBroker()`).
- **`RedisBroker`**: Distributed broker for multi-node clusters (`realtime.NewRedisBroker(redisClient)` using `go-redis/v9`).

### 2. Creating the Article Controller (`internal/controllers/article_controller.go`)

Create `internal/controllers/article_controller.go`:

```go
package controllers

import (
	"fmt"
	"time"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
	"ztatic-go-framework/security/audit"
	"mywebsite/internal/models"
	"mywebsite/internal/repositories"
	"mywebsite/internal/views/components"
)

type ArticleController struct {
	Repo   *repositories.ArticleRepository
	Broker realtime.EventBroker
}

func (ac *ArticleController) Create(c *ztatic.Context) error {
	newArticle := &models.Article{
		Title:       c.FormValue("title"),
		Content:     c.FormValue("content"),
		Author:      c.FormValue("author"),
		SecretNotes: c.FormValue("secret_notes"),
		CreatedAt:   time.Now(),
	}

	if newArticle.Author == "" {
		newArticle.Author = "Anonymous"
	}

	// 1. Validate payload with struct rules
	if err := ztatic.Validate(newArticle); err != nil {
		return err // Automatically mapped to HTTP 422 with field details
	}

	// 2. Persist to database
	ctx := c.Request().Context()
	if err := ac.Repo.Insert(ctx, newArticle); err != nil {
		return ztatic.ErrInternal("Failed to save article").WithInternal(err)
	}

	// 3. Record activity audit entry
	if entry := ztatic.AuditFromContext(c); entry != nil {
		entry.WithTarget("article", fmt.Sprint(newArticle.ID), newArticle.Title).
			WithCategory(audit.CategoryData)
	}

	// 4. Broadcast Turbo Stream DOM mutation to all SSE subscribers on "articles" topic
	ac.Broker.Publish(ctx, "articles", fullstack.TurboStreamItem{
		Action:    fullstack.StreamPrepend,
		Target:    "articles-container",
		Component: components.ArticleCard(*newArticle),
	})

	// 5. If request came from Hotwire Turbo, return direct stream fragment; otherwise redirect
	if c.Request().Header.Get("Accept") == fullstack.MIMETurboStream {
		return fullstack.RenderTurboStream(
			c,
			fullstack.StreamPrepend,
			"articles-container",
			components.ArticleCard(*newArticle),
		)
	}

	return c.Redirect(303, "/")
}
```

---

## Step 8: Observability: Tracing, Logging & Activity Ledger

Ztatic provides end-to-end request visibility out of the box with zero external dependencies.

### 1. W3C Distributed Tracing & Request ID Correlation (`trace`)

- Every request automatically receives a timestamped, URL-safe Request ID (`req-...`) and a 16-byte W3C Trace ID.
- The `trace` middleware enforces W3C recommendations:
  - Parses incoming `traceparent` headers (`00-<trace_id>-<span_id>-<flags>`) and `tracestate`.
  - Sanitizes untrusted user-supplied `X-Request-ID` headers to prevent header injection.
  - Automatically injects `X-Request-ID` and `traceparent` into HTTP response headers.
- **Context Accessors**:
  ```go
  reqID   := ztatic.RequestIDFromContext(c)
  traceID := ztatic.TraceIDFromContext(c)
  spanID  := ztatic.SpanIDFromContext(c)
  tc      := ztatic.TraceFromContext(c)
  ```
- **Outbound HTTP Calls**: Propagate trace context to downstream services using the tracing HTTP client:
  ```go
  client := trace.NewClient(nil) // Wraps http.DefaultClient with tracing RoundTripper
  req, _ := http.NewRequestWithContext(c.Request().Context(), "GET", "https://api.internal/data", nil)
  resp, err := client.Do(req)
  ```

### 2. Structured Operational Logging (`log`)

- Built on standard library `log/slog`.
- Dual-mode output: formatted color terminal output during development (`APP_ENV=development`), and structured NDJSON in production for log aggregators (Datadog, Loki, Splunk).
- Requests automatically log latency, method, path, IP, and status with level mapping (2xx/3xx $\rightarrow$ `INFO`, 4xx $\rightarrow$ `WARN`, 5xx $\rightarrow$ `ERROR`).
- Enriched request logger:
  ```go
  logger := ztatic.LogFromContext(c)
  logger.Info("processing order", "order_id", 123)
  ```
- Sensitive attributes (passwords, tokens, keys) are automatically masked as `[REDACTED]`.
- Dynamically adjust log levels at runtime without restarting:
  ```go
  app.SetLogLevel(log.LevelDebug)
  ```

### 3. Activity & Audit Logging (`security/audit`)

- Automatically audits state-mutating requests (`POST`, `PUT`, `PATCH`, `DELETE`) and error responses (`status >= 400`).
- Enrich audit events inside your handlers using `ztatic.AuditFromContext(c)` or `ztatic.AuditRecord(c, ...)`:
  ```go
  if entry := ztatic.AuditFromContext(c); entry != nil {
      entry.WithTarget("user", "usr-42", "admin@company.com").
            WithCategory(audit.CategoryAuth).
            WithMetadata("mfa_verified", true)
  }
  ```

### 4. Standardized Domain Errors (`errors`)

- Decouple business logic from HTTP status codes:
  ```go
  return ztatic.ErrNotFound("Article does not exist").WithMetadata("article_id", id)
  ```
- Standard error responses conform to RFC 9457 Problem Details (`application/problem+json`).
- In production (`APP_ENV=production`), internal database errors and stack traces are hidden from users while preserving request correlation IDs.

---

## Step 9: Asset Management & Client Integration

Ztatic eliminates Node.js and Webpack by bundling assets in Go via `esbuild`.

### 1. Global CSS Styling (`assets/css/app.css`)

Create `assets/css/app.css`:

```css
body {
    -webkit-font-smoothing: antialiased;
}

.turbo-progress-bar {
    height: 3px;
    background-color: #2563eb;
}
```

### 2. Asset Pipeline

`fullstack.MountAssets(app.Echo, fs, isDev)` serves assets with live reloading during development, and from Go's embedded filesystem with SHA-256 content hashing in production.

---

## Step 10: Assembling the Full Application Server (`cmd/server/main.go`)

Now tie all components together in `cmd/server/main.go`:

```go
package main

import (
	"context"
	"io/fs"
	"log"
	"os"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver

	"ztatic-go-framework"
	"ztatic-go-framework/data"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/realtime"
	"mywebsite"
	"mywebsite/db/migrations"
	"mywebsite/internal/config"
	"mywebsite/internal/controllers"
	"mywebsite/internal/repositories"
	"mywebsite/internal/views/components"
	"mywebsite/internal/views/layouts"
)

func main() {
	// 1. Load configuration with startup validation
	cfg := ztatic.MustLoadConfig[config.AppConfig]()

	// 2. Initialize security-hardened engine
	app := ztatic.NewSecure()

	// Configure field-level AES-256 encryption key
	if !cfg.CipherKey.IsEmpty() {
		_, _ = app.SetCipherKey([]byte(cfg.CipherKey.Expose()))
	}

	// 3. Initialize Database Engine & run embedded schema migrations
	dbEngine, err := data.NewDBEngine("sqlite", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbEngine.Close()

	migrator := data.NewMigrationEngine(dbEngine.SQL)
	if err := migrator.RunMigrations(migrations.MigrationFS, ".", "sqlite3"); err != nil {
		log.Fatalf("Database schema migration failed: %v", err)
	}

	// 4. Initialize Data Repositories & Real-Time Broker
	articleRepo := repositories.NewArticleRepository(dbEngine)
	broker := realtime.NewMemoryBroker()
	articleCtrl := &controllers.ArticleController{
		Repo:   articleRepo,
		Broker: broker,
	}

	// 5. Mount static assets (development live-reload vs production single-binary embed)
	isDev := ztatic.ActiveProfile().IsDevelopment()
	if isDev {
		fullstack.MountAssets(app.Echo, os.DirFS("dist"), true)
	} else if distSub, err := fs.Sub(mywebsite.DistFS, "dist"); err == nil {
		fullstack.MountAssets(app.Echo, distSub, false)
	} else {
		fullstack.MountAssets(app.Echo, os.DirFS("dist"), false)
	}

	// 6. Register HOTW Web Routes
	app.GET("/", func(c *ztatic.Context) error {
		ctx := c.Request().Context()
		articles, err := articleRepo.Find(ctx, data.OrderByDesc("created_at"), data.Limit(20))
		if err != nil {
			return err
		}
		return fullstack.RenderLayout(c, 200, layouts.AppLayout, components.ArticleList(articles))
	})
	app.GET("/sse", realtime.SSEHandler(broker))
	app.POST("/articles", articleCtrl.Create)

	// 7. Mount Scaffolder REST API & Interactive OpenAPI Scalar Docs
	rapid.RegisterResource(app.Group("/api"), "articles", articleRepo)
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	// 8. Start HTTP Server
	log.Printf("🚀 %s starting on :%s (profile: %s)...\n", cfg.AppName, cfg.Port, ztatic.ActiveProfile())
	log.Fatal(app.Start(":" + cfg.Port))
}
```

---

## Step 11: Development Workflow (Live Reload)

Start active development with:

```bash
ztatic dev
```

### What Happens in Dev Mode?
1. Monitors `.go`, `.templ`, `.css`, `.js`, and `.env` files using `fsnotify` with a 100ms debouncer.
2. Compiles modified Templ components (`templ generate`).
3. Executes sub-10ms `esbuild` bundling for CSS and JavaScript into `dist/`.
4. Auto-rebuilds and restarts the Go application binary seamlessly.
5. Emits a live reload event over SSE to refresh browser tabs automatically.

---

## Step 12: Production Build & Single-Binary Deployment

To build an optimized, standalone executable for production:

```bash
ztatic build
```

### The 4-Step Production Build Pipeline
1. **Templ Compilation**: Compiles `.templ` files into optimized Go code.
2. **Asset Minification**: `esbuild` bundles and minifies JS/CSS targeting ES2022.
3. **Content Hashing**: Generates SHA-256 asset content hashes and `dist/manifest.json`.
4. **Single-Binary Compilation**: Executes `go build -ldflags="-s -w" -trimpath` embedding assets via root `dist.go` (`//go:embed all:dist`).

### Deploying the Binary

Deploying requires zero external runtime dependencies, node packages, or static asset folders:

```bash
# Copy binary to production server
scp bin/server user@your-server.com:/opt/mywebsite/

# Run the single binary in production
APP_ENV=production PORT=80 /opt/mywebsite/server
```

---

## Summary Checklist

- [x] Installed `ztatic` CLI and verified environment.
- [x] Scaffolded project with `ztatic new mywebsite`.
- [x] Learned the 13 core framework packages (`ztatic`, `config`, `trace`, `data`, `rapid`, `validation`, `response`, `errors`, `log`, `security`, `fullstack`, `realtime`, `echo`).
- [x] Set up type-safe environment configuration with 4-tier dotenv cascading and `SecretString`.
- [x] Initialized database connection pool with `data.NewDBEngine`, embedded Goose SQL migrations, and AES-256 field encryption.
- [x] Implemented data access using `data.BaseRepository[T]`, fluent query helpers, and nested transactions with savepoints.
- [x] Created type-safe Templ views with smart layout unwrapping and Alpine.js micro-interactions.
- [x] Auto-generated OpenAPI 3.0 documentation and Scalar UI at `/docs`.
- [x] Added real-time DOM mutations via SSE and Turbo Streams without client JavaScript.
- [x] Leveraged W3C distributed tracing, structured logging, and activity ledgers.
- [x] Built and deployed a single, self-contained static binary with `ztatic build`.

Congratulations! You have mastered the Ztatic framework ecosystem and are ready to build modern, high-performance web applications in Go!
