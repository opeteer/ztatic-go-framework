# Ztatic Framework Tutorial: Building High-Performance Web Applications in Go

Welcome to the official step-by-step tutorial for building modern, secure, full-stack web applications using the **Ztatic Framework**.

Ztatic is an enterprise-grade, security-first full-stack Go framework built on top of **Echo v5**. It enables developers to build rich, dynamic, real-time web applications using the **HOTW (HTML Over The Wire)** stack—combining **Templ**, **Hotwire (Turbo 8)**, **Alpine.js**, native **esbuild** integration, and **Server-Sent Events (SSE)**—all without requiring Node.js or npm.

---

## Table of Contents

1. [Understanding the Ztatic Architecture](#1-understanding-the-ztatic-architecture)
2. [Ztatic Framework Directory & Module Structure](#2-ztatic-framework-directory--module-structure)
3. [Step 1: Environment Setup & Installing the `ztatic` CLI](#step-1-environment-setup--installing-the-ztatic-cli)
4. [Step 2: Scaffolding a New Application](#step-2-scaffolding-a-new-application)
5. [Step 3: Database & Data Tier Setup](#step-3-database--data-tier-setup)
6. [Step 4: Crafting Views (Layouts & Templ Components)](#step-4-crafting-views-layouts--templ-components)
7. [Step 5: Implementing Controllers & Routing](#step-5-implementing-controllers--routing)
8. [Step 6: Adding Real-Time Updates via SSE & Turbo Streams](#step-6-adding-real-time-updates-via-sse--turbo-streams)
9. [Step 7: Asset Management & Client Micro-Interactions](#step-7-asset-management--client-micro-interactions)
10. [Step 8: Development Workflow (Live Reload)](#step-8-development-workflow-live-reload)
11. [Step 9: Production Build & Deployment](#step-9-production-build--deployment)

---

## 1. Understanding the Ztatic Architecture

Before writing code, it is essential to understand how Ztatic operates under the hood.

```
                  ┌────────────────────────────────────────────────────────┐
                  │                    Client Browser                      │
                  └───────┬────────────────────────┬───────────────────────┘
                          │                        │
               HTML /     │                        │ SSE Event Stream
         Turbo Streams    │                        │ (<turbo-stream-from>)
                          ▼                        ▼
      ┌────────────────────────────────────────────────────────────────┐
      │                      Ztatic Engine                             │
      │  ┌──────────────────────────────────────────────────────────┐  │
      │  │ Security Pipeline (WAF, CSP, Hardened CSRF, HSTS)        │  │
      │  └────────────────────────────┬─────────────────────────────┘  │
      │                               ▼                                │
      │  ┌──────────────────────────────────────────────────────────┐  │
      │  │ Echo v5 Core Router & Request Handling                   │  │
      │  └──────┬─────────────────────┬─────────────────────┬───────┘  │
      │         │                     │                     │          │
      │         ▼                     ▼                     ▼          │
      │   ┌───────────┐         ┌───────────┐         ┌───────────┐    │
      │   │ Fullstack │         │ Realtime  │         │   Data    │    │
      │   │ (Templ +  │         │  (SSE +   │         │ (Engine + │    │
      │   │ esbuild)  │         │ Pub/Sub)  │         │  Generic) │    │
      │   └───────────┘         └───────────┘         └───────────┘    │
      └────────────────────────────────────────────────────────────────┘
```

### Key Technical Pillars

1. **Zero Node.js Overhead**: Native Go bindings run `esbuild` directly in memory to transpile TypeScript, bundle JavaScript, and compile CSS in under 10 milliseconds.
2. **HTML Over The Wire (HOTW)**: Instead of sending JSON and executing heavy JavaScript frameworks on the client, Ztatic renders type-safe Go components using **Templ** and streams DOM mutations over HTTP via **Hotwire Turbo 8**.
3. **Smart Layout Unwrapping**: During navigation inside `<turbo-frame>` containers, Ztatic automatically detects frame request headers (`Turbo-Frame`) and bypasses outer HTML layout rendering, reducing bandwidth consumption by up to 80%.
4. **Real-Time Push without JS Boilerplate**: By linking Server-Sent Events (SSE) directly to Hotwire Turbo Streams (`<turbo-stream-from src="/sse?topic=room:101">`), the backend can mutate client DOM elements instantly without requiring any custom JavaScript client code.
5. **Zero-Trust Web Security Suite**: Features built-in Web Application Firewall (WAF) payload inspection, nonces-based Content Security Policy (CSP), Argon2id password hashing, AES-256-GCM field encryption, and zero-allocation PII log masking out of the box.

---

## 2. Ztatic Framework Directory & Module Structure

When you scaffold a Ztatic project, you work with a standard architectural layout designed for scalability, separation of concerns, and rapid maintenance.

### Standard Scaffolding Structure (`ztatic new`)

```text
mywebsite/
├── assets/                  # Frontend raw source files
│   ├── css/                 # Global Stylesheets & Tailwind/CSS source files
│   └── js/                  # Alpine.js modules, controllers, TypeScript/JS source
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint & HTTP server bootstrapping
├── db/
│   └── migrations/          # Embedded SQL schema migrations (Goose format)
├── internal/
│   ├── controllers/         # HTTP Route Handlers & Controller logic
│   ├── models/              # Business Domain structs & database schemas
│   ├── repositories/        # Generic CRUD Data Access layer
│   └── views/               # Type-safe Templ view templates
│       ├── components/      # Modular, reusable UI components
│       └── layouts/         # Master layout wrappers (HTML head, shell, nav)
└── go.mod                   # Go module definition
```

### Core Framework Modules Breakdown

Ztatic is structured into several modular packages within `ztatic-go-framework`:

| Module Package | Path | Responsibilities |
| :--- | :--- | :--- |
| **`ztatic`** | [`ztatic.go`](file:///home/opeteer/ztatic-go-framework/ztatic.go) | Central engine constructor (`NewSecure()`), pre-wiring WAF, CSRF, CSP, and rate limiting onto Echo v5. |
| **`fullstack`** | [`fullstack/`](file:///home/opeteer/ztatic-go-framework/fullstack) | Layout rendering (`RenderLayout`), Turbo Stream responses (`RenderTurboStream`), asset pipeline (`MountAssets`, `esbuild`), and content hashing manifest. |
| **`realtime`** | [`realtime/`](file:///home/opeteer/ztatic-go-framework/realtime) | Memory event broker, SSE HTTP Handler (`SSEHandler`), topic pub/sub, and WebSocket utilities. |
| **`security`** | [`security/`](file:///home/opeteer/ztatic-go-framework/security) | WAF inspection engine, security response headers, hardened CSRF middleware, Argon2id hashing, and AES-256 field encryption. |
| **`data`** | [`data/`](file:///home/opeteer/ztatic-go-framework/data) | Database pool wrapper (`DBEngine`), panic-safe ACID transactions (`Transaction`), generic `BaseRepository[T]`, and embedded SQL migrations (`Goose`). |
| **`rapid`** | [`rapid/`](file:///home/opeteer/ztatic-go-framework/rapid) | Automatic RESTful controller mapping (`RegisterResource`) and struct tag validation (`BindAndValidate`). |

---

## Step 1: Environment Setup & Installing the `ztatic` CLI

### Prerequisites
- **Go 1.21+** installed on your operating system.
- Standard C/C++ compiler toolchain (optional, for native compilation steps).

### Installing the Ztatic CLI

Clone or locate the `ztatic-go-framework` repository and compile the unified `ztatic` CLI binary:

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

## Step 3: Database & Data Tier Setup

Ztatic provides an enterprise data layer via `ztatic-go-framework/data`.

### 1. Define a Data Model (`internal/models/article.go`)

Create `internal/models/article.go` with validation and encryption tags:

```go
package models

import "time"

type Article struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title" validate:"required,min=3,max=100"`
	Content   string    `json:"content" db:"content" validate:"required"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

### 2. Implement Generic Repository (`internal/repositories/article_repository.go`)

Leverage Ztatic's generic `BaseRepository[T]` to avoid writing repetitive CRUD code:

```go
package repositories

import (
	"context"
	"database/sql"
	
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

// Custom row scanner mapping SQL columns into struct
func ScanArticle(row data.Scanner, a *models.Article) error {
	return row.Scan(&a.ID, &a.Title, &a.Content, &a.Author, &a.CreatedAt)
}

// Custom repository method wrapped in panic-safe ACID transaction
func (r *ArticleRepository) CreateWithAudit(ctx context.Context, article *models.Article) error {
	return r.DB.Transaction(ctx, func(tx *sql.Tx) error {
		query := `INSERT INTO articles (title, content, author, created_at) VALUES (?, ?, ?, ?)`
		res, err := tx.ExecContext(ctx, query, article.Title, article.Content, article.Author, article.CreatedAt)
		if err != nil {
			return err
		}
		
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		article.ID = int(id)
		return nil
	})
}
```

---

## Step 4: Crafting Views (Layouts & Templ Components)

Ztatic uses **Templ** for type-safe, compiled HTML templates in Go.

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
		<link rel="stylesheet" href={ fullstack.AssetURL("/assets/css/app.css") }/>
		<script src="https://cdn.jsdelivr.net/npm/@hotwired/turbo@8.0.0-beta.2/dist/turbo.es2017-umd.js"></script>
		<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
	</head>
	<body class="bg-gray-100 text-gray-900 font-sans">
		<header class="bg-blue-600 text-white p-4 shadow-md">
			<div class="container mx-auto flex justify-between items-center">
				<h1 class="text-xl font-bold">Ztatic Powered Site</h1>
				<nav class="space-x-4">
					<a href="/" class="hover:underline">Home</a>
					<a href="/articles" class="hover:underline">Articles</a>
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

Create a reusable component with Turbo Frame acceleration:

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

## Step 5: Implementing Controllers & Routing

### Article Controller (`internal/controllers/article_controller.go`)

Connect repository models, layout rendering, and Turbo Stream mutation responses:

```go
package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
	"mywebsite/internal/models"
	"mywebsite/internal/repositories"
	"mywebsite/internal/views/components"
	"mywebsite/internal/views/layouts"
)

type ArticleController struct {
	Repo   *repositories.ArticleRepository
	Broker realtime.EventBroker
}

func NewArticleController(repo *repositories.ArticleRepository, broker realtime.EventBroker) *ArticleController {
	return &ArticleController{Repo: repo, Broker: broker}
}

// Index renders full page or unwrapped Turbo Frame automatically
func (ac *ArticleController) Index(c *echo.Context) error {
	// Sample articles (or fetch from repository)
	articles := []models.Article{
		{ID: 1, Title: "Getting Started with Ztatic", Content: "Ztatic combines Go and HOTW stack for ultra-fast rendering.", Author: "Alex", CreatedAt: time.Now()},
	}

	// RenderLayout unwraps outer shell automatically on Turbo-Frame requests
	return fullstack.RenderLayout(c, http.StatusOK, layouts.AppLayout, components.ArticleList(articles))
}

// Create handles submission & broadcasts real-time DOM mutation
func (ac *ArticleController) Create(c *echo.Context) error {
	title := c.FormValue("title")
	content := c.FormValue("content")

	newArticle := models.Article{
		ID:        time.Now().Nanosecond(),
		Title:     title,
		Content:   content,
		Author:    "Anonymous",
		CreatedAt: time.Now(),
	}

	// 1. Publish real-time Turbo Stream append event to all connected SSE clients
	ac.Broker.Publish(c.Request().Context(), "articles", fullstack.TurboStreamItem{
		Action:    fullstack.StreamPrepend,
		Target:    "articles-container",
		Component: components.ArticleCard(newArticle),
	})

	// 2. Return direct Turbo Stream response to the creator client
	return fullstack.RenderTurboStream(
		c,
		fullstack.StreamPrepend,
		"articles-container",
		components.ArticleCard(newArticle),
	)
}
```

---

## Step 6: Adding Real-Time Updates via SSE & Turbo Streams

Ztatic includes a zero-dependency real-time pub/sub broker out of the box in `ztatic-go-framework/realtime`.

### Wiring SSE in `cmd/server/main.go`

```go
package main

import (
	"log"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
	
	"mywebsite/internal/controllers"
)

func main() {
	// Initialize Ztatic Secure Engine (WAF, CSRF, CSP enabled)
	app := ztatic.NewSecure()

	// Initialize Realtime Memory Broker
	broker := realtime.NewMemoryBroker()

	// Mount SSE endpoint
	app.GET("/sse", realtime.SSEHandler(broker))

	// Mount Static Assets Pipeline
	fullstack.MountAssets(app.Engine, "assets", true)

	// Controllers
	articleCtrl := &controllers.ArticleController{Broker: broker}

	// Routes
	app.GET("/", articleCtrl.Index)
	app.POST("/articles", articleCtrl.Create)

	log.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(app.Start(":8080"))
}
```

Now, when any user submits a new article, `ac.Broker.Publish` triggers an SSE update. All browsing clients connected via `<turbo-stream-from src="/sse?topic=articles">` instantly receive the `<turbo-stream>` payload and prepend the new article to their DOM with zero custom JavaScript!

---

## Step 7: Asset Management & Client Micro-Interactions

Ztatic uses a native Go `esbuild` binding to bundle frontend scripts and styles on the fly.

### Asset Directory Setup (`assets/css/app.css`)

```css
/* Custom CSS styling processed by esbuild */
body {
    -webkit-font-smoothing: antialiased;
}

.turbo-progress-bar {
    height: 3px;
    background-color: #2563eb;
}
```

### Using Alpine.js for Client State

Alpine.js can be embedded inside Templ components for local micro-interactions (modals, dropdowns, form toggles):

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

---

## Step 8: Development Workflow (Live Reload)

During active development, use the `ztatic dev` command:

```bash
ztatic dev
```

### What happens in Dev Mode?
1. Monitors changes to `.go`, `.templ`, `.css`, and `.js` files.
2. Automatically compiles modified Templ components (`templ generate`).
3. Executes sub-10ms `esbuild` bundling for CSS and JavaScript.
4. Auto-rebuilds and restarts the Go application process seamlessly.

---

## Step 9: Production Build & Deployment

To build your site for production distribution:

```bash
ztatic build
```

### The Production Build Pipeline
1. **Templ Component Compilation**: Compiles all `.templ` files into Go bytecode.
2. **Minification & Asset Bundling**: `esbuild` minifies JS/CSS targeting ES2022.
3. **Content Hashing**: Generates SHA-256 asset content hashes and `manifest.json`.
4. **Single-Binary Artifact**: Executes `go build -ldflags="-s -w" -trimpath` embedding all assets into a single static binary artifact in `bin/server`.

### Deploying the Binary

Deploying your site to production requires no external dependencies, Node.js runtime, or static folder uploading:

```bash
# Copy binary to deployment host
scp bin/server user@your-server.com:/opt/mywebsite/

# Run the single binary on server
/opt/mywebsite/server
```

---

## Summary Checklist

- [x] Installed `ztatic` CLI tool.
- [x] Scaffolded project with `ztatic new mywebsite`.
- [x] Learned framework module layout (`ztatic`, `fullstack`, `realtime`, `security`, `data`, `rapid`).
- [x] Built generic repositories using `data.BaseRepository`.
- [x] Created type-safe Templ components & master layout shell.
- [x] Integrated smart layout unwrapping using `fullstack.RenderLayout`.
- [x] Added real-time DOM updates via `realtime.SSEHandler` and Hotwire Turbo Streams.
- [x] Built single self-contained binary artifact with `ztatic build`.

Congratulations! You have successfully built a full-stack, secure, real-time web application using the Ztatic framework!
