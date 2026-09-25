package ztatic

import "ztatic-go-framework/security/web"

// Config defines the top-level configuration for the Ztatic engine.
type Config struct {
	// Security contains the web security suite configuration.
	Security web.Config
}

// DefaultConfig returns the default enterprise-grade Zero-Trust configuration.
func DefaultConfig() Config {
	return Config{
		Security: web.DefaultConfig(),
	}
}
