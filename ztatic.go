package ztatic

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/security/web"
)

// Engine represents the Ztatic Framework engine, wrapping Echo v5
// with enterprise-grade Zero-Trust Security defaults.
type Engine struct {
	*echo.Echo
}

// NewSecure initializes a new Ztatic Engine pre-wired with the complete
// Zero-Trust Web Security Suite. It is safe by default.
func NewSecure() *Engine {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig initializes a new Ztatic Engine with the provided configuration.
func NewWithConfig(cfg Config) *Engine {
	e := echo.New()
	
	// Core robust middleware
	e.Use(middleware.Recover())
	
	// Phase 2: Web Security Hardening Pipeline
	if cfg.Security.EnableHeaders {
		e.Use(web.SecureHeadersWithConfig(cfg.Security.Headers))
	}
	
	if cfg.Security.EnableWAF {
		e.Use(web.WAFWithConfig(cfg.Security.WAF))
	}
	
	if cfg.Security.EnableCSRF {
		e.Use(web.HardenedCSRFWithConfig(cfg.Security.CSRF))
	}
	
	if cfg.Security.EnableRateLimiter {
		e.Use(web.AdaptiveRateLimiterWithConfig(cfg.Security.RateLimiter))
	}
	
	// Register the high-performance Struct Validator for Rapid DX
	e.Validator = rapid.NewStructValidator()
	
	return &Engine{Echo: e}
}

// New creates a raw Ztatic Engine without the full security pipeline,
// primarily for internal services or APIs behind another gateway.
func New() *Engine {
	e := echo.New()
	e.Use(middleware.Recover())
	return &Engine{Echo: e}
}
