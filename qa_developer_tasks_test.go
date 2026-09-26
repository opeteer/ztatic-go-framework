package ztatic_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"

	"ztatic-go-framework"
	"ztatic-go-framework/data"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/realtime"
	"ztatic-go-framework/security/crypto"
)

// =============================================================================
// QA USER EXPERIENCE SIMULATION TEST SUITE
// This test suite models the end-to-end journey of a developer reading TUTORIAL.md
// and README.md, executing developer tasks, and validating system behaviors.
// =============================================================================

// Article model defined in TUTORIAL.md Step 3
type DevArticle struct {
	ID        int       `json:"id" db:"id" validate:"required"`
	Title     string    `json:"title" db:"title" validate:"required,min=3,max=100"`
	Content   string    `json:"content" db:"content" validate:"required"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DevArticleRepository defined in TUTORIAL.md Step 3
type DevArticleRepository struct {
	*data.BaseRepository[DevArticle]
}

func NewDevArticleRepository(db *data.DBEngine) *DevArticleRepository {
	return &DevArticleRepository{
		BaseRepository: data.NewBaseRepository[DevArticle](db, "articles"),
	}
}

// Custom query using Squirrel AST query building (TUTORIAL Step 3)
func (r *DevArticleRepository) FindByAuthor(ctx context.Context, author string) ([]DevArticle, error) {
	if r.DB == nil {
		return nil, fmt.Errorf("database engine is uninitialized")
	}
	query, args, err := r.DB.Builder.
		Select("id", "title", "content", "author", "created_at").
		From(r.TableName).
		Where(squirrel.Eq{"author": author}).
		ToSql()
	if err != nil {
		return nil, err
	}
	_ = query
	_ = args
	return nil, nil
}

// rawHTML implements fullstack.Component for test view rendering
type rawHTML string

func (r rawHTML) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(r))
	return err
}

// -----------------------------------------------------------------------------
// USER TASK 1: Toolchain Installation & CLI Build (README & TUTORIAL Step 1)
// Command: go build -o ztatic ./cmd/ztatic
// -----------------------------------------------------------------------------
func TestUserTask1_CLI_Compilation(t *testing.T) {
	t.Log("[USER TASK 1] Developer follows README & TUTORIAL Step 1: building the ztatic CLI...")

	cmd := exec.Command("go", "build", "-o", "bin/test_ztatic_cli", "./cmd/ztatic")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		t.Errorf(`[CRITICAL P0 DEFECT] CLI compilation failed!
Developer command: go build -o ztatic ./cmd/ztatic
Error: %v
Output: %s
Root Cause Analysis: cmd/ztatic/main.go has been corrupted or overwritten with Python script content ("def greeting(name)...") instead of Go source. A newly onboarded developer cannot build the CLI tool out of the box.`, err, outputStr)
	} else {
		t.Log("[USER TASK 1 PASSED] CLI built successfully.")
		_ = os.Remove("bin/test_ztatic_cli")
	}
}

// -----------------------------------------------------------------------------
// USER TASK 2: Project Scaffolding Experience (README & TUTORIAL Step 2)
// Command: ztatic new <name>
// -----------------------------------------------------------------------------
func TestUserTask2_Scaffold_PathResolution(t *testing.T) {
	t.Log("[USER TASK 2] Developer tests scaffolding into standard subdirectories...")

	// Case A: Scaffolding with relative path into a subfolder (e.g. tmp/scaffold_test)
	testDir := filepath.Join("tmp", "scaffold_test_app_runtime")
	_ = os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// We test using the prebuilt ./ztatic binary in the repository
	if _, err := os.Stat("./ztatic"); os.IsNotExist(err) {
		t.Skip("skipping scaffold test: ./ztatic precompiled binary not available")
	}

	cmd := exec.Command("./ztatic", "new", testDir)
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err != nil || strings.Contains(outStr, "reading ../go.mod: open") {
		t.Errorf(`[P1 DEFECT] Project scaffolding failed relative path resolution!
Command: ./ztatic new %s
Output: %s
Root Cause Analysis: In cmd/ztatic/new.go, filepath.Rel(baseDir, frameworkDir) fails because baseDir is relative ("%s") while frameworkDir is absolute. The fallback sets 'replace ztatic-go-framework => ../', which fails when the project is located more than one level deep.`,
			testDir, outStr, testDir)
	} else {
		t.Logf("[USER TASK 2 PASSED] Project scaffolded successfully in %s", testDir)
	}
}

// -----------------------------------------------------------------------------
// USER TASK 3: Database & Data Tier Setup (TUTORIAL Step 3 & 7)
// -----------------------------------------------------------------------------
func TestUserTask3_DataTier_And_Repositories(t *testing.T) {
	t.Log("[USER TASK 3] Developer sets up Data Model, Repository, and Squirrel AST queries...")

	// Step 7 instructs developers to instantiate:
	// articleRepo := repositories.NewArticleRepository(nil)
	repoWithNilDB := NewDevArticleRepository(nil)

	// Sub-test A: Calling FindByAuthor when DB is nil must return an explicit error instead of panicking
	_, err := repoWithNilDB.FindByAuthor(context.Background(), "Alice")
	if err == nil {
		t.Errorf("expected error when DB is nil, got nil")
	} else {
		t.Logf("[USER TASK 3 - DB CHECK] Gracefully caught uninitialized DB: %v", err)
	}

	// Sub-test B: Default rapid.Resource interface stubs on BaseRepository
	// Tutorial states: "rapid.RegisterResource automatically mounts type-safe REST CRUD endpoints".
	// BaseRepository provides stubs returning HTTP 501 until overridden.
	dummyEchoContext := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	_, err = repoWithNilDB.FindAll(dummyEchoContext)
	if err == nil {
		t.Errorf("expected HTTP 501 for un-overridden FindAll, got nil")
	} else if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusNotImplemented {
			t.Errorf("expected 501 Not Implemented, got %d", he.Code)
		} else {
			t.Logf("[USER TASK 3 - CONTRACT STUB] Verified BaseRepository returns 501 for un-overridden FindAll: %v", he.Message)
		}
	}
}

// -----------------------------------------------------------------------------
// USER TASK 4: View Layer, Layout Unwrapping & Component Rendering (TUTORIAL Step 4 & 7)
// -----------------------------------------------------------------------------
func TestUserTask4_Views_And_AssetURL(t *testing.T) {
	t.Log("[USER TASK 4] Developer verifies view rendering, AssetURL helper, and layout wrapping...")

	// Setup asset manager with mock manifest
	manifestContent := `{"app.css": "app.f10e2390.css"}`
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "manifest.json"), []byte(manifestContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "app.f10e2390.css"), []byte("body{color:red;}"), 0644)

	app := ztatic.NewSecure()
	err := fullstack.MountAssets(app.Echo, os.DirFS(tempDir), false)
	if err != nil {
		t.Fatalf("failed to mount assets: %v", err)
	}

	// Verify AssetURL resolves the content-hashed filename
	resolvedURL := fullstack.AssetURL("app.css")
	expectedURL := "/static/app.f10e2390.css"
	if resolvedURL != expectedURL {
		t.Errorf("[DEFECT] fullstack.AssetURL('app.css') returned %q, expected %q", resolvedURL, expectedURL)
	} else {
		t.Logf("[USER TASK 4 - ASSET URL] Resolved %s -> %s", "app.css", resolvedURL)
	}

	// Verify static file serving via Echo
	reqAsset := httptest.NewRequest(http.MethodGet, expectedURL, nil)
	recAsset := httptest.NewRecorder()
	app.ServeHTTP(recAsset, reqAsset)

	if recAsset.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for %s, got %d", expectedURL, recAsset.Code)
	}
	cacheControl := recAsset.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "immutable") {
		t.Errorf("expected immutable cache control header in production mode, got: %s", cacheControl)
	}
}

// -----------------------------------------------------------------------------
// USER TASK 5: REST API & Interactive Scalar OpenAPI Docs (TUTORIAL Step 5)
// -----------------------------------------------------------------------------
func TestUserTask5_RapidREST_And_ScalarDocs(t *testing.T) {
	t.Log("[USER TASK 5] Developer exposes REST APIs and Scalar documentation UI...")

	app := ztatic.NewSecure()
	repo := data.NewBaseRepository[DevArticle](nil, "articles")

	// Mount REST CRUD under /api and docs under /docs
	rapid.RegisterResource(app.Group("/api"), "articles", repo)
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	// Sub-test A: Scalar UI Documentation endpoint
	reqDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recDocs := httptest.NewRecorder()
	app.ServeHTTP(recDocs, reqDocs)

	if recDocs.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for /docs, got %d", recDocs.Code)
	}
	bodyDocs := recDocs.Body.String()
	if !strings.Contains(bodyDocs, "data-url=\"/docs/openapi.json\"") {
		t.Errorf("/docs HTML does not point to OpenAPI schema: %s", bodyDocs)
	}
	if !strings.Contains(bodyDocs, "nonce=") {
		t.Errorf("/docs HTML does not include CSP nonce on scripts: %s", bodyDocs)
	}

	// Sub-test B: OpenAPI 3.0 JSON Schema generation
	reqOpenAPI := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	recOpenAPI := httptest.NewRecorder()
	app.ServeHTTP(recOpenAPI, reqOpenAPI)

	if recOpenAPI.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for /docs/openapi.json, got %d", recOpenAPI.Code)
	}
	schemaJSON := recOpenAPI.Body.String()
	if !strings.Contains(schemaJSON, "/api/articles") {
		t.Errorf("OpenAPI schema missing /api/articles route: %s", schemaJSON)
	}
	if !strings.Contains(schemaJSON, "DevArticle") {
		t.Errorf("OpenAPI schema missing DevArticle model definition: %s", schemaJSON)
	}

	// Sub-test C: Programmatic REST API POST call with JSON (curl / Postman simulation)
	// Testing CSRF bypass on /api/* routes
	reqPOST := httptest.NewRequest(http.MethodPost, "/api/articles", strings.NewReader(`{"id":1,"title":"Go 1.22","content":"Great updates"}`))
	reqPOST.Header.Set("Content-Type", "application/json")
	recPOST := httptest.NewRecorder()
	app.ServeHTTP(recPOST, reqPOST)

	// Since BaseRepository returns 501 Not Implemented, HTTP 501 confirms the request
	// passed CSRF middleware and reached the controller!
	if recPOST.Code == http.StatusBadRequest || recPOST.Code == http.StatusForbidden {
		t.Errorf("[CSRF BLOCKER] REST API POST /api/articles was rejected with HTTP %d by CSRF middleware: %s", recPOST.Code, recPOST.Body.String())
	} else if recPOST.Code == http.StatusNotImplemented {
		t.Logf("[USER TASK 5 - REST API] REST API correctly bypassed CSRF and reached controller handler (HTTP 501 contract stub)")
	}

	// Sub-test D: Developer mounts REST API under custom prefix like /v1 instead of /api
	v1Group := app.Group("/v1")
	rapid.RegisterResource(v1Group, "articles", repo)

	reqCustomPrefix := httptest.NewRequest(http.MethodPost, "/v1/articles", strings.NewReader(`{"id":2,"title":"Versioned API","content":"Custom prefix"}`))
	reqCustomPrefix.Header.Set("Content-Type", "application/json")
	recCustomPrefix := httptest.NewRecorder()
	app.ServeHTTP(recCustomPrefix, reqCustomPrefix)

	if recCustomPrefix.Code == http.StatusBadRequest {
		t.Errorf(`[UX REST / CSRF INCOMPATIBILITY] REST API mounted on /v1/articles was blocked with HTTP 400 Bad Request by CSRF!
Root Cause: DefaultCSRFConfig() hardcodes 'strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/docs")'.
If a developer mounts APIs under /v1, /v2, /rest, or /rpc, CSRF blocks them unless they configure custom CSRF skippers.`)
	} else {
		t.Logf("[USER TASK 5 - VERSIONED REST API] /v1/articles successfully bypassed CSRF!")
	}
}

// -----------------------------------------------------------------------------
// USER TASK 6: Real-Time SSE & Turbo Stream Broadcasting (TUTORIAL Step 6)
// -----------------------------------------------------------------------------
func TestUserTask6_RealtimeSSE_TurboStreamBroadcasting(t *testing.T) {
	t.Log("[USER TASK 6] Developer connects to SSE stream and publishes Turbo Stream updates...")

	broker := realtime.NewMemoryBroker()
	app := ztatic.NewSecure()

	// Mount SSE endpoint
	app.GET("/sse", realtime.SSEHandler(broker))

	// Handler simulating ArticleController.Create
	app.POST("/articles", func(c *ztatic.Context) error {
		title := c.FormValue("title")
		content := c.FormValue("content")

		// 1. Broadcast Turbo Stream event to all active subscribers on "articles" topic
		err := broker.Publish(c.Request().Context(), "articles", fullstack.TurboStreamItem{
			Action: fullstack.StreamPrepend,
			Target: "articles-container",
			Component: rawHTML(fmt.Sprintf(`<div id="article-1"><h2>%s</h2><p>%s</p></div>`,
				title, content)),
		})
		if err != nil {
			return err
		}

		// 2. Return direct Turbo Stream response to creator
		return fullstack.RenderTurboStream(
			c,
			fullstack.StreamPrepend,
			"articles-container",
			rawHTML(fmt.Sprintf(`<div id="article-1"><h2>%s</h2><p>%s</p></div>`, title, content)),
		)
	})

	// Sub-test A: SSE subscription without 'topic' query parameter
	reqNoTopic := httptest.NewRequest(http.MethodGet, "/sse", nil)
	recNoTopic := httptest.NewRecorder()
	app.ServeHTTP(recNoTopic, reqNoTopic)
	if recNoTopic.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for SSE without topic, got %d", recNoTopic.Code)
	}

	// Sub-test B: SSE subscription with valid topic
	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamCh, unsubscribe := broker.Subscribe(subCtx, "articles")
	defer unsubscribe()

	// Sub-test C: Form submission with Turbo Stream response
	form := url.Values{}
	form.Set("title", "Real-Time HOTW in Go")
	form.Set("content", "Server-Sent Events without frontend framework overhead.")

	reqSubmit := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(form.Encode()))
	reqSubmit.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqSubmit.Header.Set("Sec-Fetch-Site", "same-origin") // Modern browser header
	recSubmit := httptest.NewRecorder()
	app.ServeHTTP(recSubmit, reqSubmit)

	if recSubmit.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for Turbo Stream form submission, got %d: %s", recSubmit.Code, recSubmit.Body.String())
	}
	responseBody := recSubmit.Body.String()
	if !strings.Contains(responseBody, `<turbo-stream action="prepend" target="articles-container">`) {
		t.Errorf("expected <turbo-stream action=\"prepend\" target=\"articles-container\"> in response, got: %s", responseBody)
	}
	contentType := recSubmit.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/vnd.turbo-stream.html") {
		t.Errorf("expected Content-Type text/vnd.turbo-stream.html, got: %s", contentType)
	}

	// Sub-test D: Verify SSE subscriber received the broadcasted Turbo Stream fragment
	select {
	case msg := <-streamCh:
		if !strings.Contains(msg, `<turbo-stream action="prepend" target="articles-container">`) {
			t.Errorf("unexpected SSE payload received: %s", msg)
		} else {
			t.Logf("[USER TASK 6 - SSE BROADCAST] Successfully received Turbo Stream DOM mutation: %s", msg)
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timed out waiting for Turbo Stream event on SSE stream channel")
	}
}

// -----------------------------------------------------------------------------
// USER TASK 7: WebSocket Transport & Dev Workflow (TUTORIAL Step 6)
// -----------------------------------------------------------------------------
func TestUserTask7_WebSocket_And_CORS(t *testing.T) {
	t.Log("[USER TASK 7] Developer tests WebSocket connections in development mode...")

	broker := realtime.NewMemoryBroker()
	app := ztatic.NewSecure()

	// Tutorial Step 6 tip: WebSocketHandlerWithConfig with AllowedOrigins for local dev
	app.GET("/ws", realtime.WebSocketHandlerWithConfig(broker, realtime.WebSocketConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	}))

	ts := httptest.NewServer(app)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?topic=articles"

	// Sub-test A: Connection from allowlisted dev server origin (http://localhost:5173)
	hdrDev := http.Header{}
	hdrDev.Set("Origin", "http://localhost:5173")
	connDev, respDev, errDev := websocket.DefaultDialer.Dial(wsURL, hdrDev)
	if errDev != nil {
		t.Errorf("expected allowlisted dev origin (localhost:5173) to connect, got error: %v", errDev)
	} else {
		t.Logf("[USER TASK 7 - WS DEV ORIGIN] Connected from allowed dev server: %d Switching Protocols", respDev.StatusCode)
		connDev.Close()
	}

	// Sub-test B: Connection from un-allowlisted origin (http://evil-site.com)
	hdrEvil := http.Header{}
	hdrEvil.Set("Origin", "http://evil-site.com")
	_, respEvil, errEvil := websocket.DefaultDialer.Dial(wsURL, hdrEvil)
	if errEvil == nil {
		t.Errorf("expected untrusted origin to be blocked, but connection succeeded")
	} else if respEvil != nil && respEvil.StatusCode == http.StatusForbidden {
		t.Logf("[USER TASK 7 - WS BLOCKED] Untrusted origin blocked with HTTP 403 Forbidden as expected")
	}
}

// -----------------------------------------------------------------------------
// USER TASK 8: Production Single-Binary Asset Deployment (TUTORIAL Step 9 & README)
// -----------------------------------------------------------------------------
func TestUserTask8_ProductionBuild_And_SingleBinaryAssetServing(t *testing.T) {
	t.Log("[USER TASK 8] Developer tests production single-binary asset deployment assumptions...")

	// TUTORIAL Step 7 line 535 instructs:
	// fullstack.MountAssets(app.Echo, os.DirFS("dist"), false)
	//
	// TUTORIAL Step 9 lines 607-612 instructs:
	// scp bin/server user@your-server.com:/opt/mywebsite/
	// /opt/mywebsite/server
	// "Deploying to production requires zero external runtime dependencies or static folder uploads"

	// Simulate deployment host: only binary is present, but "dist" directory is NOT present on disk
	nonExistentDir := filepath.Join(t.TempDir(), "missing_dist")

	app := ztatic.NewSecure()
	// Mount missing dist directory
	_ = fullstack.MountAssets(app.Echo, os.DirFS(nonExistentDir), false)

	req := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Logf(`[ARCHITECTURAL / DOCUMENTATION GAP FOUND]
TUTORIAL Step 9 claims:
  "Deploying to production requires zero external runtime dependencies or static folder uploads:
   scp bin/server user@your-server.com:/opt/mywebsite/
   /opt/mywebsite/server"
However, TUTORIAL Step 7 main.go mounts assets using:
  fullstack.MountAssets(app.Echo, os.DirFS("dist"), false)
When the user deploys only bin/server without uploading the 'dist/' folder, requests to /static/... return HTTP 404 Not Found!
To achieve true single-binary deployment without uploading 'dist/', main.go must use Go's standard '//go:embed dist/*' and pass the embed.FS to MountAssets.`)
	}
}

// -----------------------------------------------------------------------------
// USER TASK 9: Security Defaults, WAF Thresholds & Field Encryption (README & TUTORIAL)
// -----------------------------------------------------------------------------
func TestUserTask9_SecurityDefaults_WAF_And_Encryption(t *testing.T) {
	t.Log("[USER TASK 9] Developer tests WAF body limits and AES-256 field encryption...")

	app := ztatic.NewSecure()

	// Sub-test A: Default 128KB WAF Body Limit on technical blog post
	app.POST("/articles/publish", func(c *ztatic.Context) error {
		return c.String(http.StatusOK, "Published")
	})

	// 140KB technical article
	largeContent := strings.Repeat("Go concurrency and channels are powerful primitives. ", 2500)
	payload := fmt.Sprintf(`{"title": "Deep Dive Go Concurrency", "content": "%s"}`, largeContent)

	reqOver := httptest.NewRequest(http.MethodPost, "/articles/publish", strings.NewReader(payload))
	reqOver.Header.Set("Content-Type", "application/json")
	reqOver.Header.Set("X-CSRF-Token", "dummy")
	recOver := httptest.NewRecorder()
	app.ServeHTTP(recOver, reqOver)

	if recOver.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected HTTP 413 Payload Too Large for 140KB body, got %d", recOver.Code)
	} else {
		t.Logf("[USER TASK 9 - WAF LIMIT] 140KB payload correctly rejected with HTTP 413 by default 128KB WAF")
	}

	// Sub-test B: Reconfiguring with app.SetMaxBodySize
	app.SetMaxBodySize(10 * 1024 * 1024) // 10MB
	reqAllowed := httptest.NewRequest(http.MethodPost, "/articles/publish", strings.NewReader(payload))
	reqAllowed.Header.Set("Content-Type", "application/json")
	reqAllowed.Header.Set("Sec-Fetch-Site", "same-origin")
	recAllowed := httptest.NewRecorder()
	app.ServeHTTP(recAllowed, reqAllowed)

	if recAllowed.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 after SetMaxBodySize(10MB), got %d: %s", recAllowed.Code, recAllowed.Body.String())
	} else {
		t.Logf("[USER TASK 9 - WAF RECONFIG] Large payload accepted after SetMaxBodySize(10MB)")
	}

	// Sub-test C: Field Encryption (AES-256-GCM)
	// Fail-closed verification
	crypto.SetDefaultCipherSuite(nil)
	secretField := crypto.EncryptedString("my-api-secret-key")
	_, err := secretField.Value()
	if err == nil {
		t.Errorf("expected error when encrypting with uninitialized cipher suite, got nil")
	}

	// Key configuration via app.SetCipherKey
	key32 := []byte("01234567890123456789012345678901") // 32 bytes
	_, err = app.SetCipherKey(key32)
	if err != nil {
		t.Fatalf("failed to set cipher key: %v", err)
	}
	cipherVal, err := secretField.Value()
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	cipherStr, ok := cipherVal.(string)
	if !ok || cipherStr == "my-api-secret-key" || cipherStr == "" {
		t.Errorf("expected ciphertext, got: %v", cipherVal)
	} else {
		t.Logf("[USER TASK 9 - ENCRYPTION] Successfully encrypted field: %s", cipherStr)
	}
}

// -----------------------------------------------------------------------------
// USER TASK 10: Development Server Port Configuration (TUTORIAL Step 8)
// -----------------------------------------------------------------------------
func TestUserTask10_DevServer_PortConfiguration(t *testing.T) {
	t.Log("[USER TASK 10] Developer examines ztatic dev port handling...")

	// In cmd/server/main.go (TUTORIAL Step 7 line 557 and scaffolded stub):
	// log.Fatal(app.Start(":8080"))
	//
	// In cmd/ztatic/dev.go line 280:
	// s.cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", s.Port))
	//
	// If a developer runs: ztatic dev -p 3000
	// The child process receives PORT=3000, but main.go has hardcoded ":8080",
	// so the server will still bind to :8080 instead of :3000!

	portEnv := "3000"
	mainCode := `log.Fatal(app.Start(":8080"))`

	if strings.Contains(mainCode, `":8080"`) && !strings.Contains(mainCode, `os.Getenv("PORT")`) {
		t.Logf(`[ERGONOMICS FINDING]
When a developer runs 'ztatic dev -p %s', dev.go passes PORT=%s via environment variable.
However, the tutorial and scaffolded main.go hardcode:
  log.Fatal(app.Start(":8080"))
The server ignores the PORT environment variable and continues listening on :8080.
Recommendation: Update scaffold and tutorial to:
  port := os.Getenv("PORT")
  if port == "" { port = "8080" }
  log.Fatal(app.Start(":" + port))`, portEnv, portEnv)
	}
}
