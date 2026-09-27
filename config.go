package ztatic

import (
	"ztatic-go-framework/errors"
	"ztatic-go-framework/log"
	"ztatic-go-framework/security/session"
	"ztatic-go-framework/security/token"
	"ztatic-go-framework/security/web"
)

// Config defines the top-level configuration for the Ztatic engine.
type Config struct {
	// Security contains the web security suite configuration.
	Security web.Config

	// Log contains the structured logging engine configuration.
	Log log.Config

	// Error contains the standardized error handling configuration.
	Error errors.Config

	// Session contains optional default session configuration.
	Session *session.SessionConfig

	// Token contains optional default token manager configuration.
	Token *token.Config
}

// DefaultConfig returns the default enterprise-grade Zero-Trust configuration.
func DefaultConfig() Config {
	return Config{
		Security: web.DefaultConfig(),
		Log:      log.DefaultConfig(),
		Error:    errors.DefaultConfig(),
	}
}
