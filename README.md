# Ztatic Full-Stack Go Framework

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Security](https://img.shields.io/badge/Security-First-red?style=flat-square&logo=shield)](https://github.com/opeteer/ztatic-go-framework)
[![Architecture](https://img.shields.io/badge/Architecture-HOTW-success?style=flat-square)](https://github.com/opeteer/ztatic-go-framework)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

**Ztatic** is a modern, high-performance, security-first full-stack framework built on top of **Echo v5**. It enables Go developers to build high-productivity full-stack web applications using the **HOTW (HTML Over The Wire)** stack: **Templ**, **Hotwire (Turbo)**, **Alpine.js**, native **esbuild**, and **Realtime Pub/Sub (SSE & WebSockets)**.

With Ztatic, you write pure Go and HTML—**zero Node.js or npm required**—and compile your entire application (templates, TypeScript/JS, CSS, database migrations) into a **single, statically-linked binary**.

---

## Key Features & Architecture

### 1. Rapid REST API & OpenAPI 3.0 Engine (`rapid`)
* **Dynamic Route Introspection:** Automatically inspects registered routes and normalizes Echo path parameters (`/users/:id` → `/users/{id}`).
* **Reflection-Based Struct Tag Parser:** Converts Go struct `json` and `validate` tags (e.g. `required`, `email`, `min`, `max`, `len`) into OpenAPI 3.0.3 JSON schemas with cyclic reference safety.
* **Auto-Scaffolded Resource Controllers:** `rapid.RegisterResource[T]` automatically generates type-safe REST CRUD endpoints (`GET`, `POST`, `PUT`, `DELETE`) with embedded OpenAPI documentation.
* **Interactive Scalar UI:** Serves dynamic interactive API documentation via Scalar at `/docs` and raw spec schemas at `/docs/openapi.json`.

### 2. Multi-Node Realtime Pub/Sub Engine (`realtime`)
* **Dual Event Brokers:** Thread-safe `MemoryBroker` for local development and distributed `RedisBroker` (`go-redis/v9`) supporting standalone Redis, Sentinel, or Redis Cluster configurations for multi-node deployments.
* **SSE & WebSocket Transports:** Serves native Hotwire Turbo Stream DOM updates over Server-Sent Events (`<turbo-stream-from src="/sse?topic=room">`) or full-duplex WebSockets (`gorilla/websocket`) with automated ping/pong keep-alive heartbeats and graceful disconnect handling.

### 3. Native Asset Pipeline (`fullstack`)
* **In-Memory esbuild Bundler:** Direct Go integration with `esbuild` compiles TypeScript, JavaScript, and CSS in sub-10ms with structured terminal error reporting (`AssetBundleError`).
* **Dual-Mode Asset Mounting:** Disk-backed live reloads during development; Go 1.16+ `embed.FS` with SHA-256 content hashing (`app.a8f9b2.js`) and long-term HTTP caching headers in production.

### 4. Data & Persistence Engine (`data`)
* **AST Query Building:** Integrates `Masterminds/squirrel` for type-safe query construction with automatic SQL dialect placeholder detection (PostgreSQL `$1`, MySQL/SQLite `?`).
* **Embedded Migrations:** Runs `pressly/goose/v3` SQL schema migrations directly from embedded `embed.FS` directories inside your binary.
* **Panic-Safe Transactions:** `DBEngine.Transaction(ctx, fn)` automatically handles `COMMIT` on success, and `ROLLBACK` on explicit errors or runtime panics.

### 5. High-Performance Core Router & TLS Proxy (`echo`)
* **Radix Tree Node Compaction:** `DefaultRouter.Remove` dynamically merges single-child non-handler nodes upon route deletion, preventing tree fragmentation and guaranteeing $O(\text{URL length})$ routing performance.
* **HTTPS/TLS Reverse Proxy:** Proxy middleware supports custom `crypto/tls.Config` settings (`InsecureSkipVerify`, custom root CAs) for both HTTP reverse proxying (`proxyHTTP`) and raw WebSocket TLS tunneling (`proxyRaw`).

### 6. Built-in Security Suite (`security`)
* **Web Security Engine:** Web Application Firewall (WAF) payload inspection, nonces-based Content Security Policy (CSP), HTTP Strict Transport Security (HSTS), and Double Submit Cookie CSRF protection.
* **Data Privacy & Encryption:** AES-256-GCM field-level struct tag encryption (`ztatic:"encrypt"`), Argon2id password hashing, and zero-allocation PII masking for `slog`.

### 7. Developer CLI (`ztatic`)
* **Cobra CLI Suite:** Simple developer tooling for scaffolding, building, and running applications.
* **Live Reload Engine:** `ztatic dev` monitors `.go`, `.templ`, `.css`, and `.js` files via `fsnotify`, automatically regenerating components, bundling assets, and restarting the server with a 100ms debouncer.
* **Single-Binary Build:** `ztatic build` runs a 4-step pipeline (`templ generate`, `esbuild`, asset manifest generation, `go build`) to create a zero-dependency static binary.

---

## Quick Start

### Installation & Prerequisites
Requires **Go 1.21+**.

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

This creates the standard Ztatic project layout:
```text
myapp/
├── assets/
│   ├── css/
│   └── js/
├── cmd/
│   └── server/
│       └── main.go
├── db/
│   └── migrations/
├── internal/
│   ├── controllers/
│   ├── models/
│   ├── repositories/
│   └── views/
│       ├── components/
│       └── layouts/
└── go.mod
```

### 2. Basic Server Example (`cmd/server/main.go`)
```go
package main

import (
	"log"
	
	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
)

func main() {
	// Initialize security-hardened engine
	app := ztatic.NewSecure()

	// Serve static assets
	fullstack.MountAssets(app.Engine, assets.FS, true)

	// Mount dynamic OpenAPI documentation at /docs
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Engine, "/docs")

	// Define routes
	app.GET("/", func(c ztatic.Context) error {
		return c.String(200, "Welcome to Ztatic Full-Stack Framework!")
	})

	log.Fatal(app.Start(":8080"))
}
```

---

## Developer Usage Examples

### Rapid REST API & OpenAPI Docs (`rapid`)

Define a struct with validation tags:
```go
type User struct {
	ID    int    `json:"id" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2"`
}
```

Register a full REST resource in one line:
```go
// Automatically mounts GET /, GET /:id, POST /, PUT /:id, DELETE /:id
// and registers OpenAPI 3.0 schema definitions in /docs/openapi.json
rapid.RegisterResource(app.Group("/api/users"), "users", userRepository)
```

### Real-Time Pub/Sub & WebSockets (`realtime`)

```go
package main

import (
	"github.com/labstack/echo/v5"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
)

func SetupRealtime(e *echo.Echo) {
	// Use MemoryBroker for single-instance or RedisBroker for multi-node clusters
	broker := realtime.NewMemoryBroker()

	// Mount SSE or WebSocket endpoints
	e.GET("/sse", realtime.SSEHandler(broker))
	e.GET("/ws", realtime.WebSocketHandler(broker))

	// Broadcast DOM mutations from anywhere in your backend
	e.POST("/chat/send", func(c *echo.Context) error {
		broker.Publish(c.Request().Context(), "chat-room-1", fullstack.TurboStreamItem{
			Action:    fullstack.StreamAppend,
			Target:    "chat-box",
			Component: components.ChatMessage("Hello from Go!"),
		})
		return c.NoContent(200)
	})
}
```

Connect natively in HTML with zero client JavaScript:
```html
<turbo-stream-from src="/sse?topic=chat-room-1"></turbo-stream-from>
<div id="chat-box"></div>
```

---

## Development & Production Workflow

### Development Mode (Live Reload)
```bash
./ztatic dev
```
Monitors `.templ`, `.go`, `.css`, and `.js` files, auto-compiling templates, bundling assets via `esbuild`, and live-reloading the application.

### Production Build
```bash
./ztatic build
```
Executes the production build pipeline:
1. `templ generate` compiles templates to Go code.
2. `esbuild` bundles and minifies JS/CSS assets targeting ES2022.
3. Content-hash manifest (`manifest.json`) is generated.
4. `go build -ldflags="-s -w" -trimpath` creates a zero-dependency static binary.

---

## License

Distributed under the MIT License. See `LICENSE` for more information.
