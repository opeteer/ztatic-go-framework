package main

import (
	"log"
	"os"

	"ztatic-go-framework"
)

func main() {
	app := ztatic.NewSecure()
	
	app.GET("/", func(c *ztatic.Context) error {
		return c.String(200, "Welcome to Ztatic!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Start(":" + port))
}
