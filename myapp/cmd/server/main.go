package main

import (
	"log"
	"os"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/rapid"
	"ztatic-go-framework/realtime"
	"myapp/internal/controllers"
	"myapp/internal/models"
	"myapp/internal/repositories"
	"myapp/internal/views/components"
	"myapp/internal/views/layouts"
)

func main() {
	// 1. Initialize Zero-Trust security engine
	app := ztatic.NewSecure()

	// 2. Mount static asset pipeline (serves dist/ with cache-busting)
	fullstack.MountAssets(app.Echo, os.DirFS("dist"), false)

	// 3. Initialize data tier & event broker
	broker := realtime.NewMemoryBroker()
	articleRepo := repositories.NewArticleRepository(nil)
	articleController := &controllers.ArticleController{Broker: broker}

	// 4. Register application routes
	app.GET("/", func(c *ztatic.Context) error {
		sampleArticles := []models.Article{
			{ID: 1, Title: "First Article", Content: "Hello from Ztatic!", Author: "Alice"},
		}
		return fullstack.Render(c, 200, layouts.AppLayout(components.ArticleList(sampleArticles)))
	})
	app.GET("/sse", realtime.SSEHandler(broker))
	app.POST("/articles", articleController.Create)

	// 5. Mount REST CRUD & OpenAPI Scalar docs
	rapid.RegisterResource(app.Group("/api"), "articles", articleRepo)
	rapid.DefaultOpenAPIGenerator.ServeDocs(app.Echo, "/docs")

	log.Println("Ztatic application starting on :8080...")
	log.Fatal(app.Start(":8080"))
}
