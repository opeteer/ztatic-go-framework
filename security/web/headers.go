package web

import (
	"github.com/labstack/echo/v5"
)

// SecureHeaders returns a middleware with relaxed security headers for optimal speed
// and compatibility with external CDNs and frontend frameworks.
func SecureHeaders() echo.MiddlewareFunc {
	return SecureHeadersWithConfig(DefaultHeaderConfig())
}

// SecureHeadersWithConfig returns a SecureHeaders middleware with config.
func SecureHeadersWithConfig(cfg HeaderConfig) echo.MiddlewareFunc {
	if cfg.Skipper == nil {
		cfg.Skipper = DefaultHeaderConfig().Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			res := c.Response()

			if cfg.XSSProtection != "" {
				res.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}
			if cfg.ContentTypeOptions != "" {
				res.Header().Set("X-Content-Type-Options", cfg.ContentTypeOptions)
			}
			if cfg.FrameOptions != "" {
				res.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}
			if cfg.ContentSecurityPolicy != "" {
				res.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}
			if cfg.ReferrerPolicy != "" {
				res.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			return next(c)
		}
	}
}

