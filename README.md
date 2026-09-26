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
* **WAF & Security Headers:** Inspects incoming URIs and request payload bodies (up to 128KB) for SQLi and XSS, performs recursive URL unescaping and SQL comment normalization, and injects nonces-based CSP, HSTS, and Double-Submit Cookie CSRF protection.
* **Data Privacy:** AES-256-GCM struct tag encryption (`ztatic:"encrypt"`), Argon2id password hashing, and zero-allocation PII masking for `slog`.

### 7. Developer CLI (`ztatic`)
* **Cobra CLI Suite:** Built on `spf13/cobra` for scaffolding, running, and building applications.
* **Live Reload Engine:** `ztatic dev` monitors `.go`, `.templ`, `.css`, and `.js` files using `fsnotify` with a 100ms debouncer.
* **Single-Binary Compiler:** `ztatic build` executes a 4-step pipeline (`templ generate`, `esbuild`, manifest generation, `go build`) to create an optimized production binary.

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

	// Mount static asset pipeline
	fullstack.MountAssets(app.Echo, os.DirFS("dist"), true)

	// Expose interactive OpenAPI documentation at /docs
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	// Define application routes
	app.GET("/", func(c *ztatic.Context) error {
		return c.String(200, "Welcome to Ztatic Framework!")
	})

	log.Fatal(app.Start(":8080"))
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

## License

Distributed under the MIT License. See `LICENSE` for details.
