package ztatic

import (
	"ztatic-go-framework/log"
	"ztatic-go-framework/security/web"
)

// Config defines the top-level configuration for the Ztatic engine.
type Config struct {
	// Security contains the web security suite configuration.
	Security web.Config

	// Log contains the structured logging engine configuration.
	Log log.Config
}

// DefaultConfig returns the default enterprise-grade Zero-Trust configuration.
func DefaultConfig() Config {
	return Config{
		Security: web.DefaultConfig(),
		Log:      log.DefaultConfig(),
	}
}
