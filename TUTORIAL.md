# Ztatic Framework Tutorial: Building High-Performance Web Applications in Go

Welcome to the official step-by-step tutorial for building modern, secure, full-stack web applications using the **Ztatic Framework**.

Ztatic is a modern, open-source, full-stack Go framework built on top of **Echo v5**. It enables developers to build rich, dynamic, real-time web applications using the **HOTW (HTML Over The Wire)** stack—combining **Templ**, **Hotwire (Turbo 8)**, **Alpine.js**, native **esbuild** bundling, **OpenAPI 3.0 generation**, and **Realtime Pub/Sub (SSE & WebSockets)**—all without requiring Node.js or npm.

---

## Table of Contents

1. [Understanding the Ztatic Architecture](#1-understanding-the-ztatic-architecture)
2. [Ztatic Framework Directory & Module Structure](#2-ztatic-framework-directory--module-structure)
3. [Step 1: Environment Setup & Installing the `ztatic` CLI](#step-1-environment-setup--installing-the-ztatic-cli)
4. [Step 2: Scaffolding a New Application](#step-2-scaffolding-a-new-application)
5. [Step 3: Database & Data Tier Setup (Squirrel & Goose)](#step-3-database--data-tier-setup-squirrel--goose)
6. [Step 4: Crafting Views (Layouts & Templ Components)](#step-4-crafting-views-layouts--templ-components)
7. [Step 5: Implementing Rapid REST APIs & Interactive OpenAPI Docs](#step-5-implementing-rapid-rest-apis--interactive-openapi-docs)
8. [Step 6: Adding Real-Time Updates (SSE, WebSockets & Redis Pub/Sub)](#step-6-adding-real-time-updates-sse-websockets--redis-pubsub)
9. [Step 7: Asset Management & Client Micro-Interactions](#step-7-asset-management--client-micro-interactions)
10. [Step 8: Development Workflow (Live Reload)](#step-8-development-workflow-live-reload)
11. [Step 9: Production Build & Single-Binary Deployment](#step-9-production-build--single-binary-deployment)

---

## 1. Understanding the Ztatic Architecture

Before writing code, it is essential to understand how Ztatic operates under the hood.

```
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
       │  │ Security Pipeline (WAF, CSP, Hardened CSRF, HSTS)        │  │
       │  └────────────────────────────┬─────────────────────────────┘  │
       │                               ▼                                │
       │  ┌──────────────────────────────────────────────────────────┐  │
       │  │ Echo v5 Core Router (Radix Compaction & TLS Proxy)       │  │
       │  └──────┬─────────────┬─────────────┬─────────────┬─────────┘  │
       │         │             │             │             │            │
       │         ▼             ▼             ▼             ▼            │
       │   ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐      │
       │   │ Fullstack │ │ Realtime  │ │   Data    │ │   Rapid   │      │
       │   │ (Templ +  │ │(Pub/Sub + │ │(Squirrel+ │ │ (OpenAPI  │      │
       │   │ esbuild)  │ │WS / SSE)  │ │  Goose)   │ │ + Scalar) │      │
       │   └───────────┘ └───────────┘ └───────────┘ └───────────┘      │
       └────────────────────────────────────────────────────────────────┘
```

### Key Technical Pillars

1. **Zero Node.js Overhead**: Native Go bindings run `esbuild` directly in memory to transpile TypeScript, bundle JavaScript, and compile CSS in under 10 milliseconds.
2. **HTML Over The Wire (HOTW)**: Instead of sending JSON and executing heavy client-side JavaScript frameworks, Ztatic renders type-safe Go components using **Templ** and streams DOM mutations over HTTP via **Hotwire Turbo 8**.
3. **Smart Layout Unwrapping**: During navigation inside `<turbo-frame>` containers, Ztatic automatically detects frame request headers (`Turbo-Frame`) and bypasses outer HTML layout rendering, reducing bandwidth consumption by up to 80%.
4. **Real-Time Push without JS Boilerplate**: By linking Server-Sent Events (SSE) or WebSockets directly to Hotwire Turbo Streams (`<turbo-stream-from src="/sse?topic=room">`), the backend can mutate client DOM elements instantly without requiring custom client JavaScript code.
5. **Dynamic OpenAPI 3.0 Generation**: Introspects Echo routes and struct validation tags (`json` and `validate`) using Go reflection to automatically build OpenAPI 3.0.3 specs rendered interactively via **Scalar UI**, with recursive property merging for embedded anonymous structs.
6. **Zero-Trust Web Security Suite**: Built-in Web Application Firewall (WAF) payload inspection (URIs and request bodies up to 128KB with 413 oversized rejection), multiline XSS protection, targeted HTML event-handler blocking, resilient URL unescaping, SQL comment normalization, per-request nonce-based Content Security Policy (`fullstack.Nonce(c)`), HSTS, hardened CSRF (`HttpOnly` with automatic `/api/` and `/docs` skipping, and `fullstack.CSRFField(c)` helpers), Argon2id password hashing, and AES-256-GCM database field encryption (`crypto.EncryptedString` initialized via `app.SetCipherKey` or `ZTATIC_CIPHER_KEY`).

---

## 2. Ztatic Framework Directory & Module Structure

When you scaffold a Ztatic project, you work with a standard architectural layout designed for scalability, separation of concerns, and rapid maintenance.

### Standard Scaffolding Structure (`ztatic new`)

```text
mywebsite/
├── assets/                  # Frontend raw source files
│   ├── css/                 # Global stylesheets & CSS source files
│   └── js/                  # Alpine.js modules, controllers, TypeScript/JS source
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint & HTTP server bootstrapping
├── db/
│   └── migrations/          # Embedded SQL schema migrations (Goose format)
├── internal/
│   ├── controllers/         # HTTP Route Handlers & Controller logic
│   ├── models/              # Business Domain structs & database schemas
│   ├── repositories/        # Data Access layer (Squirrel & Generic BaseRepository)
│   └── views/               # Type-safe Templ view templates
│       ├── components/      # Modular, reusable UI components
│       └── layouts/         # Master layout wrappers (HTML head, shell, nav)
└── go.mod                   # Go module definition
```

### Core Framework Modules Breakdown

| Module Package | Path | Responsibilities |
| :--- | :--- | :--- |
| **`ztatic`** | [`ztatic.go`](file:///home/opeteer/ztatic-go-framework/ztatic.go) | Central engine constructor (`NewSecure()`), exporting convenience type aliases (`Context`, `HandlerFunc`, `Map`, `Group`), pre-wiring WAF, CSRF, CSP, and rate limiting onto Echo v5. |
| **`fullstack`** | [`fullstack/`](file:///home/opeteer/ztatic-go-framework/fullstack) | Layout rendering (`RenderLayout`), Turbo Stream responses (`RenderTurboStream`), asset pipeline (`MountAssets`, `esbuild`), and content hashing manifest. |
| **`realtime`** | [`realtime/`](file:///home/opeteer/ztatic-go-framework/realtime) | Memory and Redis Pub/Sub brokers (`MemoryBroker`, `RedisBroker`), SSE Handler (`SSEHandler`), and WebSocket Handler (`WebSocketHandler`). |
| **`security`** | [`security/`](file:///home/opeteer/ztatic-go-framework/security) | WAF inspection engine (deep URI & request body payload scanning, multiline XSS, HTML event attributes, SQL comment normalization), security response headers, hardened CSRF middleware, Argon2id hashing, and AES-256 field encryption. |
| **`data`** | [`data/`](file:///home/opeteer/ztatic-go-framework/data) | Database pool wrapper (`DBEngine`), Squirrel query builder (`squirrel.StatementBuilderType`), generic `BaseRepository[T]`, and Goose embedded SQL migrations (`RunMigrations`). |
| **`rapid`** | [`rapid/`](file:///home/opeteer/ztatic-go-framework/rapid) | Automatic RESTful controller mapping (`RegisterResource`), OpenAPI 3.0 generation (`OpenAPIGenerator`) with recursive anonymous struct property merging, reflection tag parsing, and Scalar UI docs rendering. |
| **`echo`** | [`echo/`](file:///home/opeteer/ztatic-go-framework/echo) | Core router with radix tree node compaction (`Remove`), and HTTPS/TLS reverse proxying (`proxyHTTP`, `proxyRaw`). |

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

---

## Step 2: Scaffolding a New Application

Create a new application named `mywebsite`:

```bash
ztatic new mywebsite
cd mywebsite
```

This command generates the complete Ztatic architectural directory structure along with pre-configured `cmd/server/main.go` and `go.mod`.

---

## Step 3: Database & Data Tier Setup (Squirrel & Goose)

Ztatic provides a high-performance data layer via `ztatic-go-framework/data`.

### 1. Define a Data Model (`internal/models/article.go`)

Create `internal/models/article.go` with validation tags:

```go
package models

import "time"

type Article struct {
	ID        int       `json:"id" db:"id" validate:"required"`
	Title     string    `json:"title" db:"title" validate:"required,min=3,max=100"`
	Content   string    `json:"content" db:"content" validate:"required"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

> [!TIP]
> **Field-Level Encryption**: To encrypt sensitive fields (e.g., API keys, tokens) at rest using AES-256-GCM, tag fields with `ztatic:"encrypt"` or use the `crypto.EncryptedString` type. Configure your 32-byte key in code via `app.SetCipherKey([]byte("..."))` or set the `ZTATIC_CIPHER_KEY` environment variable.

### 2. Embedded SQL Migrations (`db/migrations/00001_create_articles_table.sql`)

Create an embedded migration using Goose format:

```sql
-- +goose Up
CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE articles;
```

### 3. Implement Generic Repository with Squirrel AST (`internal/repositories/article_repository.go`)

Leverage Ztatic's generic `BaseRepository[T]` and Squirrel AST query builder:

```go
package repositories

import (
	"context"

	"github.com/Masterminds/squirrel"
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

// Custom query using Squirrel AST query building
func (r *ArticleRepository) FindByAuthor(ctx context.Context, author string) ([]models.Article, error) {
	query, args, err := r.DB.Builder.
		Select("id", "title", "content", "author", "created_at").
		From(r.TableName).
		Where(squirrel.Eq{"author": author}).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.DB.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Author, &a.CreatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}
```

> [!TIP]
> **Database Initialization & Generic Stubs**:
> - Initialize your database pool using `sql.Open("sqlite", "app.db")` and wrap it via `dbEngine := &data.DBEngine{SQL: sqlDB, Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)}`.
> - `BaseRepository[T]` implements default stubs for `rapid.Resource[T]` (`FindAll`, `FindByID`, `Create`, `Update`, `Delete`) returning `HTTP 501 Not Implemented`. Developers can selectively override these methods in `ArticleRepository` with SQL queries or Squirrel AST operations.

> [!TIP]
> After setting up your database models and repository, run `go mod tidy` in your project root to download the required data tier dependencies (`squirrel` and `goose`):
> ```bash
> go mod tidy
> ```

---

## Step 4: Crafting Views (Layouts & Templ Components)

Ztatic uses **Templ** for type-safe, compiled HTML templates in Go.

> [!TIP]
> Newly scaffolded Ztatic projects are pre-wired with the Templ runtime dependency via `tools.go`. If configuring an existing module manually, add the dependency with:
> ```bash
> go get github.com/a-h/templ
> ```

### 1. Main Layout (`internal/views/layouts/app_layout.templ`)

Define the outer shell with Hotwire Turbo & Alpine.js scripts included:

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
		
		<!-- Loaded from native esbuild pipeline -->
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

### 2. Article Component (`internal/views/components/article_card.templ`)

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

## Step 5: Implementing Rapid REST APIs & Interactive OpenAPI Docs

Ztatic's `rapid` package makes building REST APIs and OpenAPI documentation effortless.

### 1. Mounting Scaffolder Controllers (`rapid.RegisterResource`)

In `cmd/server/main.go`, map your repository directly to a REST endpoint:

```go
package main

import (
	"log"

	"ztatic-go-framework"
	"ztatic-go-framework/rapid"
	"mywebsite/internal/repositories"
)

func RegisterAPI(app *ztatic.Engine, articleRepo *repositories.ArticleRepository) {
	// Automatically registers GET /, GET /:id, POST /, PUT /:id, DELETE /:id
	// and extracts struct tags (json & validate) for OpenAPI 3.0 docs!
	rapid.RegisterResource(app.Group("/api"), "articles", articleRepo)

	// Serve interactive Scalar API documentation UI at /docs
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")
}
```

Navigate to `http://localhost:8080/docs` in your browser to view the interactive Scalar API documentation!

---

## Step 6: Adding Real-Time Updates (SSE, WebSockets & Redis Pub/Sub)

Ztatic supports real-time event broadcasting over Server-Sent Events (SSE) and WebSockets.

### 1. Choosing an Event Broker (`MemoryBroker` vs. `RedisBroker`)

- **`MemoryBroker`**: Ideal for single-server setups (`realtime.NewMemoryBroker()`).
- **`RedisBroker`**: Ideal for distributed multi-node clusters (`realtime.NewRedisBroker(redisClient)` using `go-redis/v9`).

> [!TIP]
> **Local Frontend Development (WebSocket CORS):** By default, `realtime.WebSocketHandler` enforces
> same-origin verification to prevent Cross-Site WebSocket Hijacking (CSWSH). If your frontend
> dev server runs on a different port (e.g., Vite on `localhost:5173`), allowlist it explicitly:
> ```go
> app.GET("/ws", realtime.WebSocketHandlerWithConfig(broker, realtime.WebSocketConfig{
>     AllowedOrigins: []string{"http://localhost:5173"},
> }))
> ```
> Remove or restrict `AllowedOrigins` before deploying to production.

### 2. Broadcasting Real-Time Turbo Streams

```go
package controllers

import (
	"time"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
	"mywebsite/internal/models"
	"mywebsite/internal/views/components"
)

type ArticleController struct {
	Broker realtime.EventBroker
}

func (ac *ArticleController) Create(c *ztatic.Context) error {
	newArticle := models.Article{
		ID:        time.Now().Nanosecond(),
		Title:     c.FormValue("title"),
		Content:   c.FormValue("content"),
		Author:    "Anonymous",
		CreatedAt: time.Now(),
	}

	// 1. Broadcast Turbo Stream event to all active subscribers on "articles" topic
	ac.Broker.Publish(c.Request().Context(), "articles", fullstack.TurboStreamItem{
		Action:    fullstack.StreamPrepend,
		Target:    "articles-container",
		Component: components.ArticleCard(newArticle),
	})

	// 2. Return direct Turbo Stream response to creator
	return fullstack.RenderTurboStream(
		c,
		fullstack.StreamPrepend,
		"articles-container",
		components.ArticleCard(newArticle),
	)
}
```

> [!TIP]
> After adding your real-time controller and importing the `realtime` package, run `go mod tidy` in your project root to resolve event broker dependencies:
> ```bash
> go mod tidy
> ```
> *(Note: In strict air-gapped or offline sandbox environments, use `GOSUMDB=off go mod tidy` if your network restricts outbound DNS to Go checksum servers).*

---

## Step 7: Asset Management & Client Micro-Interactions

Ztatic uses a native Go `esbuild` binding to bundle frontend scripts and styles on the fly in sub-10ms.

### 1. CSS & Asset Pipeline (`assets/css/app.css`)

```css
body {
    -webkit-font-smoothing: antialiased;
}

.turbo-progress-bar {
    height: 3px;
    background-color: #2563eb;
}
```

### 2. Micro-Interactions with Alpine.js

Embed Alpine.js directly inside Templ components:

```html
templ CreateArticleModal() {
	<div x-data="{ open: false }">
		<button @click="open = true" class="bg-blue-600 text-white px-4 py-2 rounded">
			+ New Article
		</button>

		<div x-show="open" x-cloak class="fixed inset-0 bg-black/50 flex items-center justify-center">
			<div @click.away="open = false" class="bg-white p-6 rounded-lg w-96 shadow-xl">
				<h3 class="text-lg font-bold mb-4">Create New Article</h3>
				<form action="/articles" method="POST" @submit="open = false">
					<!-- Note: In browser sessions, modern browsers send Sec-Fetch-Site: same-origin automatically. -->
					<!-- To explicitly bind CSRF tokens in forms, you can include fullstack.CSRFField(c) -->
					<input type="text" name="title" placeholder="Title" required class="w-full mb-3 p-2 border rounded"/>
					<textarea name="content" placeholder="Content" required class="w-full mb-3 p-2 border rounded"></textarea>
					<div class="flex justify-end space-x-2">
						<button type="button" @click="open = false" class="px-4 py-2 border rounded">Cancel</button>
						<button type="submit" class="px-4 py-2 bg-blue-600 text-white rounded">Post</button>
					</div>
				</form>
			</div>
		</div>
	</div>
}
```

### 3. Assembling the Application Server (`cmd/server/main.go`)

Now tie all components together in `cmd/server/main.go`—mounting static assets, OpenAPI documentation, generic database repositories, and real-time event brokers:

```go
package main

import (
	"embed"
	"io/fs"
	"log"
	"os"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/realtime"
	"mywebsite/internal/controllers"
	"mywebsite/internal/models"
	"mywebsite/internal/repositories"
	"mywebsite/internal/views/components"
	"mywebsite/internal/views/layouts"
)

// Embed compiled static assets directly into the binary for zero-dependency single-binary deployment
//go:embed all:dist
var distFS embed.FS

func main() {
	// 1. Initialize Zero-Trust security engine
	app := ztatic.NewSecure()

	// 2. Mount static asset pipeline (embed.FS for production single-binary, os.DirFS for dev)
	distSub, err := fs.Sub(distFS, "dist")
	if err == nil {
		fullstack.MountAssets(app.Echo, distSub, false)
	} else {
		fullstack.MountAssets(app.Echo, os.DirFS("dist"), false)
	}

	// 3. Initialize data tier & event broker
	broker := realtime.NewMemoryBroker()
	articleRepo := repositories.NewArticleRepository(nil)
	articleController := &controllers.ArticleController{Broker: broker}

	// 4. Register application routes
	app.GET("/", func(c *ztatic.Context) error {
		sampleArticles := []models.Article{
			{ID: 1, Title: "First Article", Content: "Hello from Ztatic!", Author: "Alice"},
		}
		return fullstack.Render(c, 200, layouts.AppLayout(components.ArticleList(sampleArticles)))
	})
	app.GET("/sse", realtime.SSEHandler(broker))
	app.POST("/articles", articleController.Create)

	// 5. Mount REST CRUD & OpenAPI Scalar docs
	rapid.RegisterResource(app.Group("/api"), "articles", articleRepo)
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Ztatic application starting on :%s...\n", port)
	log.Fatal(app.Start(":" + port))
}
```

> [!TIP]
> **WAF Body Limit:** `NewSecure()` enforces a default **128 KB** maximum request body for WAF
> inspection. If your application accepts larger payloads (document imports, image uploads),
> configure the limit right after `ztatic.NewSecure()` and before `app.Start()`:
> ```go
> app := ztatic.NewSecure()
> app.SetMaxBodySize(4 * 1024 * 1024) // Allow up to 4 MB
> ```

---

## Step 8: Development Workflow (Live Reload)

During active development, run:

```bash
ztatic dev
```

### What happens in Dev Mode?
1. Monitors `.go`, `.templ`, `.css`, and `.js` files using `fsnotify` with a 100ms debouncer.
2. Automatically compiles modified Templ components (`templ generate`).
3. Executes sub-10ms `esbuild` bundling for CSS and JavaScript.
4. Auto-rebuilds and restarts the Go application process seamlessly.

---

## Step 9: Production Build & Single-Binary Deployment

To build your application for production distribution:

```bash
ztatic build
```

### The 4-Step Production Build Pipeline
1. **Templ Compilation**: Compiles `.templ` files into Go code.
2. **Asset Minification**: `esbuild` minifies JS/CSS targeting ES2022.
3. **Content Hashing**: Generates SHA-256 asset content hashes and `manifest.json`.
4. **Single-Binary Artifact**: Executes `go build -ldflags="-s -w" -trimpath` embedding all assets into a single static binary in `bin/server`.

### Deploying the Binary

Deploying to production requires zero external runtime dependencies or static folder uploads because `dist/` is compiled directly into the binary via `//go:embed all:dist`:

```bash
# Copy binary to deployment host
scp bin/server user@your-server.com:/opt/mywebsite/

# Run the single binary on server (optionally override port with PORT env)
PORT=80 /opt/mywebsite/server
```

---

## Summary Checklist

- [x] Installed `ztatic` CLI tool.
- [x] Scaffolded project with `ztatic new mywebsite`.
- [x] Learned framework module layout (`ztatic`, `fullstack`, `realtime`, `security`, `data`, `rapid`, `echo`).
- [x] Built data access layer with `data.DBEngine`, `squirrel`, `goose`, and `BaseRepository`.
- [x] Created Templ views & master layout shell.
- [x] Auto-generated OpenAPI 3.0 documentation & Scalar UI at `/docs`.
- [x] Added real-time DOM updates via SSE, WebSockets, and Pub/Sub brokers.
- [x] Built single self-contained binary artifact with `ztatic build`.

Congratulations! You have successfully built a full-stack, secure, real-time web application using the Ztatic framework!
