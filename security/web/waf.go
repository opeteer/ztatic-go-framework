package web

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/labstack/echo/v5"
)

var (
	// sqliPatterns detects severe SQL Injection payloads
	sqliPatterns = []*regexp.Regexp{
		// Hardened: using \s* around words allows catching 'union  select' or 'union select' after comments are stripped
		regexp.MustCompile(`(?i)\b(union\s+select|drop\s+table|exec\s+)\b`),
	}

	// xssPatterns detects severe Cross-Site Scripting (XSS) payloads
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?is)<script>.*?</script>`), // Hardened: (?is) makes dot match newlines
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)\bon\w+\s*=`), // Hardened: catches HTML event handlers like onerror=, onload=
	}

	// sqlCommentPattern matches SQL inline comments /* ... */
	sqlCommentPattern = regexp.MustCompile(`(?s)/\*.*?\*/`)
)

const maxWAFBodySize = 128 * 1024 // 128KB limit for memory-safe body inspection

// WAF returns a lightweight Web Application Firewall middleware that inspects
// request URI and request body (up to 128KB) for critical signatures, ensuring maximum request throughput and speed.
func WAF() echo.MiddlewareFunc {
	return WAFWithConfig(DefaultWAFConfig())
}

// WAFWithConfig returns a WAF middleware with config.
func WAFWithConfig(cfg WAFConfig) echo.MiddlewareFunc {
	if cfg.Skipper == nil {
		cfg.Skipper = DefaultWAFConfig().Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			
			// 1. Check Request URI
			uri := req.URL.RequestURI()
			if isMalicious(uri) {
				c.Logger().Warn("WAF blocked malicious URI", "uri", uri, "ip", c.RealIP())
				return echo.NewHTTPError(http.StatusForbidden, "Forbidden: Security policy violation")
			}

			// 2. Check Request Body (for POST/PUT/PATCH with payload)
			if req.Body != nil && req.ContentLength > 0 && req.ContentLength <= maxWAFBodySize {
				bodyBytes, err := io.ReadAll(io.LimitReader(req.Body, maxWAFBodySize))
				if err == nil && len(bodyBytes) > 0 {
					// Restore the body stream immediately for downstream handlers
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					
					if isMalicious(string(bodyBytes)) {
						c.Logger().Warn("WAF blocked malicious request body", "ip", c.RealIP())
						return echo.NewHTTPError(http.StatusForbidden, "Forbidden: Security policy violation")
					}
				}
			}

			return next(c)
		}
	}
}

// recursiveQueryUnescape mitigates double-encoding WAF bypasses.
func recursiveQueryUnescape(input string) string {
	maxDepth := 5
	prev := input
	for i := 0; i < maxDepth; i++ {
		decoded, err := url.QueryUnescape(prev)
		if err != nil || decoded == prev {
			return prev
		}
		prev = decoded
	}
	return prev
}

// isMalicious checks input string against key attack signatures
func isMalicious(input string) bool {
	if input == "" {
		return false
	}

	// Hardened: mitigate double URL encoding bypass
	input = recursiveQueryUnescape(input)

	// XSS Check
	for _, p := range xssPatterns {
		if p.MatchString(input) {
			return true
		}
	}

	// Hardened: strip SQL comments before checking SQLi signatures to prevent obfuscation bypass
	inputWithoutComments := sqlCommentPattern.ReplaceAllString(input, " ")

	// SQLi Check
	for _, p := range sqliPatterns {
		if p.MatchString(inputWithoutComments) {
			return true
		}
	}

	return false
}
