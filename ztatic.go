package ztatic

import (
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/security/crypto"
	"ztatic-go-framework/security/web"
)

// Engine represents the Ztatic Framework engine, wrapping Echo v5
// with enterprise-grade Zero-Trust Security defaults.
type Engine struct {
	*echo.Echo
}

// SetCipherSuite configures the AES-256-GCM cipher suite on the engine and sets it as default for field encryption.
func (eng *Engine) SetCipherSuite(cs *crypto.CipherSuite) {
	crypto.SetDefaultCipherSuite(cs)
}

// SetCipherKey initializes an AES-256-GCM cipher suite from a 32-byte key and enables field encryption.
func (eng *Engine) SetCipherKey(key []byte) (*crypto.CipherSuite, error) {
	cs, err := crypto.NewCipherSuite(key)
	if err != nil {
		return nil, err
	}
	crypto.SetDefaultCipherSuite(cs)
	return cs, nil
}

// DX Type Aliases to match README.md and simplify developer usage
type Context = echo.Context
type HandlerFunc = echo.HandlerFunc
type Map map[string]any
type Group = echo.Group

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
	
	eng := &Engine{Echo: e}

	// Auto-configure AES-256 field encryption cipher suite if environment variable is present
	if keyStr := os.Getenv("ZTATIC_CIPHER_KEY"); keyStr != "" {
		keyBytes := []byte(keyStr)
		if len(keyBytes) == 32 {
			_, _ = eng.SetCipherKey(keyBytes)
		}
	}

	return eng
}

// New creates a raw Ztatic Engine without the full security pipeline,
// primarily for internal services or APIs behind another gateway.
func New() *Engine {
	e := echo.New()
	e.Use(middleware.Recover())
	return &Engine{Echo: e}
}
