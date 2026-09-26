package ztatic_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/realtime"
)

// =============================================================================
// QA USER EXPERIENCE EVALUATION SUITE V2
// Tests the exact user experience of developers following TUTORIAL.md & README.md
// and verifies fixes for forward-reference compilation issues and go:embed constraints.
// =============================================================================

// -----------------------------------------------------------------------------
// USER TASK A: go:embed Package Scope Constraint vs Root dist.go
// Verifies that placing `//go:embed all:dist` inside cmd/server/main.go fails due to
// Go's package scoping rule, and confirms that root-level dist.go compiles cleanly.
// -----------------------------------------------------------------------------
func TestUX_TaskA_EmbedFS_PackageScopeConstraint(t *testing.T) {
	t.Log("[QA TEST A] Verifying go:embed package scope constraint and root dist.go resolution...")

	tempProjectDir := filepath.Join(t.TempDir(), "embed_test_project")
	_ = os.MkdirAll(filepath.Join(tempProjectDir, "cmd", "server"), 0755)
	_ = os.MkdirAll(filepath.Join(tempProjectDir, "dist"), 0755)
	_ = os.WriteFile(filepath.Join(tempProjectDir, "dist", "app.css"), []byte("body{}"), 0644)

	// 1. Prove that placing //go:embed all:dist in cmd/server/main.go fails in Go
	invalidMain := `package main

import (
	"embed"
	"fmt"
)

//go:embed all:dist
var distFS embed.FS

func main() {
	fmt.Println(distFS)
}
`
	mainPath := filepath.Join(tempProjectDir, "cmd", "server", "main.go")
	_ = os.WriteFile(mainPath, []byte(invalidMain), 0644)
	_ = os.WriteFile(filepath.Join(tempProjectDir, "go.mod"), []byte("module embed_test_project\n\ngo 1.21\n"), 0644)

	cmd := exec.Command("go", "build", "./cmd/server")
	cmd.Dir = tempProjectDir
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err == nil {
		t.Fatalf("expected go:embed inside cmd/server to fail when dist/ is outside package, but build succeeded")
	}
	if !strings.Contains(outputStr, "no matching files found") {
		t.Fatalf("unexpected error message: %s", outputStr)
	}
	t.Logf("Confirmed: placing //go:embed all:dist in cmd/server fails as expected: %s", strings.TrimSpace(outputStr))

	// 2. Now verify that root dist.go pattern succeeds
	distGoContent := `package embed_test_project

import "embed"

//go:embed all:dist
var DistFS embed.FS
`
	_ = os.WriteFile(filepath.Join(tempProjectDir, "dist.go"), []byte(distGoContent), 0644)

	validMain := `package main

import (
	"embed_test_project"
	"fmt"
	"io/fs"
)

func main() {
	sub, err := fs.Sub(embed_test_project.DistFS, "dist")
	if err != nil {
		panic(err)
	}
	fmt.Println(sub)
}
`
	_ = os.WriteFile(mainPath, []byte(validMain), 0644)

	cmdValid := exec.Command("go", "build", "./cmd/server")
	cmdValid.Dir = tempProjectDir
	validOut, validErr := cmdValid.CombinedOutput()
	if validErr != nil {
		t.Fatalf("root dist.go architecture failed to build: %v, output: %s", validErr, string(validOut))
	}
	t.Log("Successfully verified: root dist.go compiles cleanly with cmd/server importing root module!")
}

// -----------------------------------------------------------------------------
// USER TASK B: Templ Component Step 4 Self-Containment
// Tutorial Step 4 defines ArticleCard, ArticleList, and CreateArticleModal in
// internal/views/components/article_card.templ so linear compilation succeeds.
// -----------------------------------------------------------------------------
func TestUX_TaskB_TemplForwardReference_Step4_vs_Step7(t *testing.T) {
	t.Log("[QA TEST B] Testing whether Step 4 views compile cleanly when CreateArticleModal is included...")

	tempDir := filepath.Join(t.TempDir(), "templ_test_project")
	viewsDir := filepath.Join(tempDir, "components")
	_ = os.MkdirAll(viewsDir, 0755)

	// Updated article_card.templ with CreateArticleModal defined
	templContent := `package components

templ ArticleCard(title string) {
	<div>{ title }</div>
}

templ CreateArticleModal() {
	<div x-data="{ open: false }">
		<button @click="open = true">+ New Article</button>
		<div x-show="open">Modal Body</div>
	</div>
}

templ ArticleList(titles []string) {
	<div>
		<h1>Latest Articles</h1>
		@CreateArticleModal()
		for _, t := range titles {
			@ArticleCard(t)
		}
	</div>
}
`
	_ = os.WriteFile(filepath.Join(viewsDir, "article_card.templ"), []byte(templContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module templ_test_project\n\ngo 1.21\n\nrequire github.com/a-h/templ v0.3.1020\n"), 0644)
	if sumData, err := os.ReadFile("go.sum"); err == nil {
		_ = os.WriteFile(filepath.Join(tempDir, "go.sum"), sumData, 0644)
	}

	// Run templ generate
	cmdGen := exec.Command("templ", "generate")
	cmdGen.Dir = tempDir
	if err := cmdGen.Run(); err != nil {
		t.Skip("skipping templ test: templ CLI not available in test environment")
	}

	// Verify compilation at Step 4 succeeds
	cmdBuild := exec.Command("go", "build", "-mod=mod", "./components")
	cmdBuild.Dir = tempDir
	cmdBuild.Env = append(os.Environ(), "GOSUMDB=off", "GOPROXY=off")
	out, err := cmdBuild.CombinedOutput()
	if err != nil {
		t.Fatalf("expected Step 4 components to compile cleanly, but failed: %v\nOutput: %s", err, string(out))
	}
	t.Log("Successfully verified: Step 4 components compile without forward-reference errors!")
}

// -----------------------------------------------------------------------------
// USER TASK C: End-to-End Single-Binary Asset Delivery Verification
// Verifies that root-level embedded DistFS serves assets properly via MountAssets.
// -----------------------------------------------------------------------------
func TestUX_TaskC_RootEmbedFS_Serving(t *testing.T) {
	t.Log("[QA TEST C] Verifying root-level embedded assets serving via fullstack.MountAssets...")

	tempDir := t.TempDir()
	distDir := filepath.Join(tempDir, "dist")
	_ = os.MkdirAll(distDir, 0755)
	_ = os.WriteFile(filepath.Join(distDir, "manifest.json"), []byte(`{"app.css": "app.abc12345.css"}`), 0644)
	_ = os.WriteFile(filepath.Join(distDir, "app.abc12345.css"), []byte(`body{font-family:sans-serif;}`), 0644)

	app := ztatic.NewSecure()
	err := fullstack.MountAssets(app.Echo, os.DirFS(distDir), false)
	if err != nil {
		t.Fatalf("failed to mount assets: %v", err)
	}

	// Test URL resolution
	cssURL := fullstack.AssetURL("app.css")
	if cssURL != "/static/app.abc12345.css" {
		t.Errorf("expected /static/app.abc12345.css, got %s", cssURL)
	}

	// Test HTTP GET request
	req := httptest.NewRequest(http.MethodGet, cssURL, nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "sans-serif") {
		t.Errorf("expected asset body content, got: %s", rec.Body.String())
	}
}

// -----------------------------------------------------------------------------
// USER TASK D: Scalar UI & OpenAPI Docs Availability
// -----------------------------------------------------------------------------
func TestUX_TaskD_ScalarDocs_RouteVerification(t *testing.T) {
	t.Log("[QA TEST D] Verifying Scalar UI at /docs and OpenAPI schema at /docs/openapi.json...")

	app := ztatic.NewSecure()
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	reqDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recDocs := httptest.NewRecorder()
	app.ServeHTTP(recDocs, reqDocs)

	if recDocs.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for /docs, got %d", recDocs.Code)
	}
	if !strings.Contains(recDocs.Body.String(), "@scalar/api-reference") {
		t.Errorf("expected Scalar script in /docs, got: %s", recDocs.Body.String())
	}

	reqJSON := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	recJSON := httptest.NewRecorder()
	app.ServeHTTP(recJSON, reqJSON)

	if recJSON.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for /docs/openapi.json, got %d", recJSON.Code)
	}
}

// -----------------------------------------------------------------------------
// USER TASK E: Realtime Turbo Stream Event Delivery
// -----------------------------------------------------------------------------
func TestUX_TaskE_RealtimePubSub_EventDelivery(t *testing.T) {
	t.Log("[QA TEST E] Verifying Realtime Pub/Sub SSE message delivery...")

	broker := realtime.NewMemoryBroker()
	stream, unsubscribe := broker.Subscribe(context.Background(), "live-feed")
	defer unsubscribe()

	item := fullstack.TurboStreamItem{
		Action:    fullstack.StreamAppend,
		Target:    "feed-container",
		Component: mockHTML(`<div>New Live Item</div>`),
	}

	err := broker.Publish(context.Background(), "live-feed", item)
	if err != nil {
		t.Fatalf("broker.Publish failed: %v", err)
	}

	select {
	case msg := <-stream:
		if !strings.Contains(msg, `<turbo-stream action="append" target="feed-container">`) {
			t.Errorf("unexpected stream message: %s", msg)
		}
	default:
		t.Errorf("no message received on subscriber channel")
	}
}

type mockHTML string

func (m mockHTML) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(m))
	return err
}
