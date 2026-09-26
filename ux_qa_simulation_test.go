package ztatic_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"ztatic-go-framework"
	"ztatic-go-framework/data"
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/realtime"
	"ztatic-go-framework/security/crypto"
)

// Article model exactly as defined in TUTORIAL.md Step 3
type TutorialArticle struct {
	ID        int    `json:"id" db:"id" validate:"required"`
	Title     string `json:"title" db:"title" validate:"required,min=3,max=100"`
	Content   string `json:"content" db:"content" validate:"required"`
	Author    string `json:"author" db:"author"`
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 1: Quickstart & Documentation Experience (README / TUTORIAL Step 1 & 5)
// -----------------------------------------------------------------------------

// TestUX_Task1_ScalarDocs_CSP_Compatibility tests the developer experience
// when accessing Scalar UI at /docs on a NewSecure() application.
func TestUX_Task1_ScalarDocs_CSP_Compatibility(t *testing.T) {
	app := ztatic.NewSecure()
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for /docs, got %d", rec.Code)
	}

	body := rec.Body.String()
	csp := rec.Header().Get("Content-Security-Policy")

	t.Logf("[UX CHECK] /docs CSP header: %s", csp)

	// In Scalar UI, the page has: <script id="api-reference" data-url="/docs/openapi.json"></script>
	// Check if this script tag contains the required CSP nonce.
	hasNonceInCSP := strings.Contains(csp, "'nonce-")
	if hasNonceInCSP {
		idx := strings.Index(csp, "'nonce-")
		endIdx := strings.Index(csp[idx+7:], "'")
		nonce := csp[idx+7 : idx+7+endIdx]

		scriptTagWithNonce := fmt.Sprintf(`nonce="%s"`, nonce)
		if !strings.Contains(body, scriptTagWithNonce) {
			t.Errorf("[UX ISSUE FOUND] Scalar UI inline script lacks CSP nonce!\nCSP requires nonce %q, but /docs HTML does not include it on <script id=\"api-reference\"> or <script src=\"...\">.\nModern browsers will block execution of the Scalar API docs.\nHTML snippet:\n%s", nonce, body)
		}
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 2: REST API CRUD with curl/Postman (README & TUTORIAL Step 5)
// -----------------------------------------------------------------------------

// TestUX_Task2_RestAPI_CRUD_CSRF_Blocking tests what happens when a developer
// uses curl, Postman, or Scalar UI to POST to the auto-generated REST CRUD endpoints.
func TestUX_Task2_RestAPI_CRUD_CSRF_Blocking(t *testing.T) {
	app := ztatic.NewSecure()

	// Developer mounts REST CRUD endpoints exactly as shown in TUTORIAL Step 5
	repo := data.NewBaseRepository[TutorialArticle](nil, "articles")
	rapid.RegisterResource(app.Group("/api"), "articles", repo)

	// Developer tries to create a new article via standard REST POST /api/articles (curl / Postman / Scalar UI)
	articleJSON := `{"id": 1, "title": "My First Post", "content": "Welcome to Ztatic!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/articles", strings.NewReader(articleJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	t.Logf("[UX CHECK] POST /api/articles HTTP Status: %d, Response: %s", rec.Code, rec.Body.String())

	// Echo's CSRF middleware returns 400 Bad Request when the token extractor fails to find X-CSRF-Token or _csrf
	if rec.Code == http.StatusBadRequest || rec.Code == http.StatusForbidden {
		t.Errorf("[UX CRITICAL BLOCKER] REST API CRUD endpoint /api/articles rejected with HTTP %d CSRF error!\nDevelopers following the README/TUTORIAL cannot use curl, Postman, or Scalar UI to test APIs out-of-the-box because NewSecure() applies browser CSRF middleware to all API routes without a skipper or exemption for programmatic/JSON API calls.\nResponse: %s", rec.Code, rec.Body.String())
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 3: Web Form Submission from Tutorial Component (TUTORIAL Step 7)
// -----------------------------------------------------------------------------

// TestUX_Task3_WebForm_Submission tests form submission workflows from TUTORIAL Step 7,
// verifying both browser same-origin requests and tokenized form submissions.
func TestUX_Task3_WebForm_Submission(t *testing.T) {
	app := ztatic.NewSecure()

	// Developer registers the form action handler exactly as shown in TUTORIAL Step 6 & 7:
	// app.POST("/articles", articleController.Create)
	var articleCreated bool
	app.POST("/articles", func(c *ztatic.Context) error {
		title := c.FormValue("title")
		content := c.FormValue("content")
		if title != "" && content != "" {
			articleCreated = true
			return c.String(http.StatusOK, "Article Created: "+title)
		}
		return c.String(http.StatusBadRequest, "Missing fields")
	})

	// Scenario A: Standard browser submission with Sec-Fetch-Site: same-origin
	form := url.Values{}
	form.Set("title", "My Great Article")
	form.Set("content", "This is the content of my tutorial article")

	reqBrowser := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(form.Encode()))
	reqBrowser.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqBrowser.Header.Set("Sec-Fetch-Site", "same-origin")
	recBrowser := httptest.NewRecorder()

	app.ServeHTTP(recBrowser, reqBrowser)

	t.Logf("[UX CHECK] Browser POST /articles Form Submit HTTP Status: %d, Response: %s", recBrowser.Code, recBrowser.Body.String())
	if recBrowser.Code != http.StatusOK || !articleCreated {
		t.Errorf("expected browser same-origin form submit to succeed with 200 OK, got %d: %s", recBrowser.Code, recBrowser.Body.String())
	}

	// Scenario B: Form submission with fullstack.CSRFField / fullstack.CSRFToken
	reqProbe := httptest.NewRequest(http.MethodGet, "/articles", nil)
	recProbe := httptest.NewRecorder()
	app.ServeHTTP(recProbe, reqProbe)

	var cookie *http.Cookie
	for _, c := range recProbe.Result().Cookies() {
		if c.Name == "_csrf" {
			cookie = c
			break
		}
	}

	formWithToken := url.Values{}
	formWithToken.Set("title", "My Tokenized Article")
	formWithToken.Set("content", "Content with CSRF token")
	if cookie != nil {
		formWithToken.Set("_csrf", cookie.Value)
	}

	reqToken := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(formWithToken.Encode()))
	reqToken.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		reqToken.AddCookie(cookie)
	}
	recToken := httptest.NewRecorder()
	app.ServeHTTP(recToken, reqToken)

	t.Logf("[UX CHECK] Tokenized POST /articles Form Submit HTTP Status: %d, Response: %s", recToken.Code, recToken.Body.String())
	if recToken.Code != http.StatusOK {
		t.Errorf("expected tokenized form submit to succeed with 200 OK, got %d: %s", recToken.Code, recToken.Body.String())
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 4: WAF False Positives on Legitimate Developer Content
// -----------------------------------------------------------------------------

// TestUX_Task4_WAF_FalsePositives tests common developer blog posts, technical articles,
// and queries to see if the WAF falsely blocks legitimate user workflows.
func TestUX_Task4_WAF_FalsePositives(t *testing.T) {
	app := ztatic.NewSecure()

	app.POST("/articles/publish", func(c *ztatic.Context) error {
		return c.String(http.StatusOK, "Published successfully")
	})
	app.GET("/search", func(c *ztatic.Context) error {
		q := c.QueryParam("q")
		return c.String(http.StatusOK, "Search results for: "+q)
	})

	// Obtain valid CSRF token so we isolate WAF behavior
	reqProbe := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
	recProbe := httptest.NewRecorder()
	app.ServeHTTP(recProbe, reqProbe)

	var csrfCookie *http.Cookie
	var csrfToken string
	for _, cookie := range recProbe.Result().Cookies() {
		if cookie.Name == "_csrf" {
			csrfCookie = cookie
			csrfToken = cookie.Value
			break
		}
	}

	cases := []struct {
		name        string
		method      string
		target      string
		body        string
		contentType string
		description string
	}{
		{
			name:        "DevArticle_ExecuteCommand",
			method:      http.MethodPost,
			target:      "/articles/publish",
			body:        `{"title": "How to execute commands in Go", "content": "You can execute commands using the os/exec package safely."}`,
			contentType: "application/json",
			description: "Legitimate tech article mentioning 'execute commands' (matches \\b(exec|execute)\\s*[\\s(])",
		},
		{
			name:        "DevArticle_DropTableDiscussion",
			method:      http.MethodPost,
			target:      "/articles/publish",
			body:        `{"title": "Database Best Practices", "content": "Before you drop table in dev, ensure you have a backup."}`,
			contentType: "application/json",
			description: "Legitimate DBA article mentioning 'drop table' (matches \\bdrop\\s+table\\b)",
		},
		{
			name:        "DevArticle_OnlineStatusOrNumberOne",
			method:      http.MethodPost,
			target:      "/articles/publish",
			body:        `{"title": "System Status", "content": "Service is online = true. Priority one = highest."}`,
			contentType: "application/json",
			description: "Text with 'online =' or 'one =' (falsely matches HTML event handler regex \\bon\\w+\\s*=)",
		},
		{
			name:        "DevArticle_MarkdownDoubleDash",
			method:      http.MethodPost,
			target:      "/articles/publish",
			body:        `{"title": "Release Notes", "content": "Ztatic Framework -- modern, fast, and secure."}`,
			contentType: "application/json",
			description: "Markdown text with standard em-dash '-- ' (matches SQL comment regex (--\\s*$|--\\s+))",
		},
		{
			name:        "SearchQuery_LegitimateDevSearch",
			method:      http.MethodGet,
			target:      "/search?q=" + url.QueryEscape("how to execute tests in parallel"),
			body:        "",
			contentType: "",
			description: "Search query containing 'execute' (blocked on GET query params)",
		},
		{
			name:        "SearchQuery_OnlineStatusSearch",
			method:      http.MethodGet,
			target:      "/search?q=" + url.QueryEscape("online = false"),
			body:        "",
			contentType: "",
			description: "Search query containing 'online =' (blocked on GET query params)",
		},
		{
			name:        "SearchQuery_DropTableSearch",
			method:      http.MethodGet,
			target:      "/search?q=" + url.QueryEscape("sql syntax drop table if exists"),
			body:        "",
			contentType: "",
			description: "Search query containing 'drop table' (blocked on GET query params)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			}
			req := httptest.NewRequest(tc.method, tc.target, bodyReader)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			if csrfToken != "" {
				req.Header.Set("X-CSRF-Token", csrfToken)
			}
			if csrfCookie != nil {
				req.AddCookie(csrfCookie)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			t.Logf("[%s] Status: %d, Description: %s", tc.name, rec.Code, tc.description)
			if rec.Code == http.StatusForbidden {
				t.Errorf("[UX WAF FALSE POSITIVE] %s: Legitimate user input was BLOCKED with HTTP 403 Forbidden!\nScenario: %s\nPayload/URL: %s%s\nWAF blocked legitimate technical developer content.",
					tc.name, tc.description, tc.target, tc.body)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 5: Large Blog Post / Document Upload (>128KB)
// -----------------------------------------------------------------------------

// TestUX_Task5_OversizedPayload_Enforcement tests the developer experience
// when uploading a legitimate technical document, article with embedded markdown/SVG,
// or batch data that exceeds 128KB.
func TestUX_Task5_OversizedPayload_Enforcement(t *testing.T) {
	app := ztatic.NewSecure()
	app.POST("/articles/import", func(c *ztatic.Context) error {
		return c.String(http.StatusOK, "Imported")
	})

	// 150KB markdown article (completely benign, no attack payload)
	benignContent := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 2500)
	payload := fmt.Sprintf(`{"title": "Comprehensive Architecture Guide", "body": "%s"}`, benignContent)

	req := httptest.NewRequest(http.MethodPost, "/articles/import", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	// Provide CSRF to isolate WAF size limit
	req.Header.Set("X-CSRF-Token", "dummy")

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	t.Logf("[UX CHECK] 150KB Upload HTTP Status: %d", rec.Code)
	if rec.Code == http.StatusRequestEntityTooLarge {
		t.Logf("[UX FINDING] WAF strictly rejects any body > 128KB with HTTP 413 Payload Too Large.\nWhile secure against padding attacks, developers uploading legitimate long articles, rich HTML, or images will find their requests rejected without clear developer-facing documentation or convenient configuration hooks in NewSecure().")
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 6: CSP vs Alpine.js Micro-Interactions (TUTORIAL Step 4 & 7)
// -----------------------------------------------------------------------------

// TestUX_Task6_CSP_AlpineJS_Compatibility tests whether the CSP headers
// allow Alpine.js evaluation as prescribed in TUTORIAL Step 4 & 7.
func TestUX_Task6_CSP_AlpineJS_Compatibility(t *testing.T) {
	app := ztatic.NewSecure()
	app.GET("/", func(c *ztatic.Context) error {
		return c.String(http.StatusOK, "homepage")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	t.Logf("[UX CHECK] Homepage CSP: %s", csp)

	// Alpine.js standard build evaluates x-data and @click using JavaScript eval/new Function().
	// Standard Alpine.js REQUIRES 'unsafe-eval' in script-src to function.
	hasUnsafeEval := strings.Contains(csp, "'unsafe-eval'")
	if !hasUnsafeEval {
		t.Errorf("[UX CRITICAL INCOMPATIBILITY] Content-Security-Policy lacks 'unsafe-eval' in script-src!\nTUTORIAL Step 4 & 7 instructs developers to use Alpine.js for micro-interactions (x-data, @click).\nHowever, Alpine.js requires 'unsafe-eval' to parse and execute directives.\nIn production and modern browsers, Alpine.js will crash with EvalError and the modal will fail to open!\nCSP Header: %s", csp)
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 7: Field Encryption Defaults (README & TUTORIAL Architecture)
// -----------------------------------------------------------------------------

// TestUX_Task7_FieldEncryption verifies that field encryption enforces fail-closed
// safety (returns an error if key is unconfigured) and properly encrypts when key is set.
func TestUX_Task7_FieldEncryption(t *testing.T) {
	// Scenario A: When cipher suite is uninitialized, Value() must fail-closed with error
	crypto.SetDefaultCipherSuite(nil)
	app := ztatic.NewSecure()

	secretValue := "my-super-secret-api-key-12345"
	encryptedField := crypto.EncryptedString(secretValue)

	_, err := encryptedField.Value()
	if err == nil {
		t.Fatalf("expected fail-closed error when cipher suite is uninitialized, got nil")
	}
	t.Logf("[UX CHECK] Uninitialized field encryption properly failed closed: %v", err)

	// Scenario B: When key is configured via app.SetCipherKey, encryption succeeds
	key := []byte("12345678901234567890123456789012") // 32-byte key
	_, err = app.SetCipherKey(key)
	if err != nil {
		t.Fatalf("failed to set cipher key: %v", err)
	}

	val, err := encryptedField.Value()
	if err != nil {
		t.Fatalf("unexpected error encrypting with key: %v", err)
	}
	valStr, ok := val.(string)
	if !ok || valStr == secretValue || valStr == "" {
		t.Fatalf("expected encrypted ciphertext, got: %v", val)
	}
	t.Logf("[UX CHECK] Field encryption active: plaintext %q -> ciphertext %q", secretValue, valStr)
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 8: WebSocket Same-Origin UX in Development (TUTORIAL Step 6)
// -----------------------------------------------------------------------------

// TestUX_Task8_WebSocket_DevWorkflow tests the developer experience when
// connecting to WebSockets during local frontend development.
func TestUX_Task8_WebSocket_DevWorkflow(t *testing.T) {
	broker := realtime.NewMemoryBroker()
	app := ztatic.NewSecure()
	app.GET("/ws", realtime.WebSocketHandler(broker))

	ts := httptest.NewServer(app)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?topic=chat"

	// Scenario A: Standard same-origin browser connection
	headerSame := http.Header{}
	headerSame.Set("Origin", ts.URL)
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headerSame)
	if err != nil {
		t.Fatalf("expected same-origin WS connection to succeed: %v", err)
	}
	conn.Close()
	t.Logf("[UX CHECK] Same-origin WS connection: %d Switching Protocols", resp.StatusCode)

	// Scenario B: Frontend dev server (e.g. Vite on localhost:5173, mobile dev, or 127.0.0.1)
	headerDevServer := http.Header{}
	headerDevServer.Set("Origin", "http://localhost:5173")
	_, respDev, errDev := websocket.DefaultDialer.Dial(wsURL, headerDevServer)
	if errDev == nil {
		t.Errorf("expected cross-origin to be blocked by default")
	} else if respDev != nil {
		t.Logf("[UX CHECK] Cross-origin frontend dev server WS connection blocked: %d Forbidden (Expected by CSWSH, but requires documentation for allowed origins in dev)", respDev.StatusCode)
	}
}

// -----------------------------------------------------------------------------
// USER JOURNEY TEST 9: Comprehensive End-to-End Developer Workflow
// -----------------------------------------------------------------------------

// TestUX_EndToEnd_DeveloperJourney simulates a complete user journey reading TUTORIAL.md
// and assembling the entire application: layout, forms, REST API, Scalar docs, SSE, and encryption.
func TestUX_EndToEnd_DeveloperJourney(t *testing.T) {
	// 1. Initialize Zero-Trust security engine
	app := ztatic.NewSecure()

	// 2. Configure 32-byte field encryption key
	key := []byte("12345678901234567890123456789012")
	_, err := app.SetCipherKey(key)
	if err != nil {
		t.Fatalf("failed to set cipher key: %v", err)
	}

	// 3. Initialize real-time broker & generic repo
	broker := realtime.NewMemoryBroker()
	repo := data.NewBaseRepository[TutorialArticle](nil, "articles")

	// 4. Mount OpenAPI docs & REST CRUD
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")
	rapid.RegisterResource(app.Group("/api"), "articles", repo)

	// 5. Register application routes
	app.GET("/", func(c *ztatic.Context) error {
		return c.String(http.StatusOK, "Home Page")
	})
	app.GET("/sse", realtime.SSEHandler(broker))

	var createdArticles []TutorialArticle
	app.POST("/articles", func(c *ztatic.Context) error {
		article := TutorialArticle{
			ID:      100,
			Title:   c.FormValue("title"),
			Content: c.FormValue("content"),
			Author:  "QA Developer",
		}
		createdArticles = append(createdArticles, article)
		return c.String(http.StatusOK, "Created: "+article.Title)
	})

	// --- PHASE A: Homepage Access & Security Headers Inspection ---
	reqHome := httptest.NewRequest(http.MethodGet, "/", nil)
	recHome := httptest.NewRecorder()
	app.ServeHTTP(recHome, reqHome)

	if recHome.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for homepage, got %d", recHome.Code)
	}
	csp := recHome.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "'unsafe-eval'") {
		t.Errorf("expected CSP to allow 'unsafe-eval' for Alpine.js, got %s", csp)
	}

	// --- PHASE B: Scalar Documentation UI & Nonce Inspection ---
	reqDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recDocs := httptest.NewRecorder()
	app.ServeHTTP(recDocs, reqDocs)

	if recDocs.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for /docs, got %d", recDocs.Code)
	}
	docsBody := recDocs.Body.String()
	if !strings.Contains(docsBody, "nonce=") {
		t.Errorf("expected Scalar docs HTML to contain nonce on scripts, got: %s", docsBody)
	}

	// --- PHASE C: REST API JSON CRUD Interaction (curl / Postman) ---
	reqAPI := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	recAPI := httptest.NewRecorder()
	app.ServeHTTP(recAPI, reqAPI)
	if recAPI.Code == http.StatusBadRequest || recAPI.Code == http.StatusForbidden {
		t.Fatalf("REST API GET /api/articles blocked by CSRF: %d", recAPI.Code)
	}

	// --- PHASE D: Web Form Submission with Technical Content ---
	form := url.Values{}
	form.Set("title", "How to execute commands in Go")
	form.Set("content", "Learn os/exec -- fast and reliable. Priority one = highest.")

	reqForm := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(form.Encode()))
	reqForm.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqForm.Header.Set("Sec-Fetch-Site", "same-origin")
	recForm := httptest.NewRecorder()
	app.ServeHTTP(recForm, reqForm)

	if recForm.Code != http.StatusOK {
		t.Fatalf("expected form submit with technical content to succeed, got %d: %s", recForm.Code, recForm.Body.String())
	}
	if len(createdArticles) != 1 {
		t.Fatalf("expected 1 created article, got %d", len(createdArticles))
	}

	// --- PHASE E: Field Encryption Verification ---
	encField := crypto.EncryptedString("sensitive-token-12345")
	encryptedVal, err := encField.Value()
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	if encryptedVal == "sensitive-token-12345" || encryptedVal == "" {
		t.Fatalf("expected encrypted ciphertext, got plaintext: %v", encryptedVal)
	}

	t.Logf("[E2E VERIFIED] Complete developer journey succeeded with zero friction across all modules!")
}

