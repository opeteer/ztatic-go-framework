package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"ztatic-go-framework"
	"ztatic-go-framework/log"
)

func main() {
	// Initialize the Ztatic Framework with Zero-Trust Security-by-Default.
	// This automatically wires Structured Request Logging, WAF, Strict Headers, CSRF, and Rate Limiting.
	app := ztatic.NewSecure()

	// Dynamic log level adjustment: easily switch between DEBUG, INFO, WARN, ERROR
	app.SetLogLevel(log.LevelInfo)

	// Add a simple route demonstrating structured context logging
	app.GET("/", func(c *echo.Context) error {
		// c.Logger() or log.FromContext(c.Request().Context()) provides
		// a request-scoped logger enriched with request_id, method, and client IP.
		logger := log.FromContext(c.Request().Context())
		logger.Info("handling welcome request", "endpoint", "/")

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to the secure Ztatic framework!",
			"status":  "secure",
		})
	})

	// Start the server
	log.Info("Starting Ztatic Engine on :8080...")
	if err := app.Start(":8080"); err != nil {
		log.Error("server crashed", "error", err)
	}
}
