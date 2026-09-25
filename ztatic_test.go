package ztatic

import (
	"net/http"
	"net/http/httptest"
	"testing"

)

func TestNewSecure(t *testing.T) {
	app := NewSecure()
	if app == nil || app.Echo == nil {
		t.Fatalf("expected non-nil app and echo instance")
	}

	if app.Validator == nil {
		t.Errorf("expected StructValidator to be set on secure engine")
	}

	app.GET("/ping", func(c *Context) error {
		return c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", rec.Code)
	}

	// Verify default Security Headers applied by NewSecure
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options header to be nosniff")
	}
}

func TestNewWithConfig_CustomSecurity(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Security.EnableHeaders = false
	cfg.Security.EnableWAF = false
	cfg.Security.EnableCSRF = false
	cfg.Security.EnableRateLimiter = false

	app := NewWithConfig(cfg)
	if app == nil {
		t.Fatalf("expected non-nil app")
	}

	app.GET("/test", func(c *Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?id=1%20UNION%20SELECT%201", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	// WAF is disabled, so SQLi URI should not be blocked by WAF
	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 when WAF is disabled, got %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "" {
		t.Errorf("expected security headers to be omitted when disabled")
	}
}

func TestNew_LeanEngine(t *testing.T) {
	app := New()
	if app == nil || app.Echo == nil {
		t.Fatalf("expected non-nil app")
	}

	app.GET("/raw", func(c *Context) error {
		return c.String(http.StatusOK, "raw response")
	})

	req := httptest.NewRequest(http.MethodGet, "/raw", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", rec.Code)
	}
	if rec.Body.String() != "raw response" {
		t.Errorf("expected body 'raw response', got %q", rec.Body.String())
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Security.EnableHeaders || !cfg.Security.EnableWAF || !cfg.Security.EnableCSRF || !cfg.Security.EnableRateLimiter {
		t.Errorf("expected default config to enable all web security modules")
	}
}
