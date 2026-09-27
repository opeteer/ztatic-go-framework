package ztatic

import (
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/security/audit"
	"ztatic-go-framework/security/crypto"
	"ztatic-go-framework/security/web"
)

// Engine represents the Ztatic Framework engine, wrapping Echo v5
// with enterprise-grade Zero-Trust Security defaults.
type Engine struct {
	*echo.Echo
	wafCfg      *web.WAFConfig // pointer allows SetMaxBodySize to adjust WAF limit after construction
	auditLogger audit.Logger
}

// AuditLogger returns the active audit logger instance, or nil if disabled.
func (eng *Engine) AuditLogger() audit.Logger {
	return eng.auditLogger
}

// SetAuditLogger sets a custom audit logger on the engine.
func (eng *Engine) SetAuditLogger(l audit.Logger) {
	eng.auditLogger = l
}

// UseAudit mounts the audit logging middleware onto the engine.
func (eng *Engine) UseAudit(cfg audit.AuditConfig) {
	eng.auditLogger = cfg.Logger
	eng.Use(audit.AuditWithConfig(cfg))
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

// SetMaxBodySize sets the WAF's maximum allowed request body size in bytes.
// Call this before starting the server. The default is 128 KB (131072 bytes).
//
// Example:
//
//	app := ztatic.NewSecure()
//	app.SetMaxBodySize(4 * 1024 * 1024) // Allow up to 4 MB
//	log.Fatal(app.Start(":8080"))
func (eng *Engine) SetMaxBodySize(bytes int64) {
	if eng.wafCfg != nil {
		eng.wafCfg.MaxBodySize = bytes
	}
}

// DX Type Aliases to match README.md and simplify developer usage
type Context = echo.Context
type HandlerFunc = echo.HandlerFunc
type Map map[string]any
type Group = echo.Group

type AuditEntry = audit.Entry
type AuditLogger = audit.Logger
type AuditConfig = audit.AuditConfig

// AuditFromContext retrieves the active audit entry from the request context.
func AuditFromContext(c *Context) *audit.Entry {
	return audit.FromContext(c)
}

// AuditRecord records a domain audit event on the active request context.
func AuditRecord(c *Context, action string, targetType, targetID string) *audit.Entry {
	return audit.Record(c, action, targetType, targetID)
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
	
	// WAFConfig is stored by pointer so SetMaxBodySize can adjust it at runtime.
	wafCfgCopy := cfg.Security.WAF
	if cfg.Security.EnableWAF {
		e.Use(web.WAFWithConfigPtr(&wafCfgCopy))
	}
	
	if cfg.Security.EnableCSRF {
		e.Use(web.HardenedCSRFWithConfig(cfg.Security.CSRF))
	}
	
	if cfg.Security.EnableRateLimiter {
		e.Use(web.AdaptiveRateLimiterWithConfig(cfg.Security.RateLimiter))
	}

	var auditLogger audit.Logger
	if cfg.Security.EnableAudit {
		auditCfg := cfg.Security.Audit
		if auditCfg.Logger == nil {
			auditCfg = audit.DefaultAuditConfig()
		}
		auditLogger = auditCfg.Logger
		e.Use(audit.AuditWithConfig(auditCfg))
	}
	
	// Register the high-performance Struct Validator for Rapid DX
	e.Validator = rapid.NewStructValidator()
	
	eng := &Engine{Echo: e, wafCfg: &wafCfgCopy, auditLogger: auditLogger}

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
