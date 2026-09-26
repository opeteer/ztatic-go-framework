package controllers

import (
	"time"

	"ztatic-go-framework"
	"ztatic-go-framework/fullstack"
	"ztatic-go-framework/realtime"
	"myapp/internal/models"
	"myapp/internal/views/components"
)

type ArticleController struct {
	Broker realtime.EventBroker
}

func (ac *ArticleController) Create(c *ztatic.Context) error {
	newArticle := models.Article{
		ID:        time.Now().Nanosecond(),
		Title:     c.FormValue("title"),
		Content:   c.FormValue("content"),
		Author:    "Anonymous",
		CreatedAt: time.Now(),
	}

	// 1. Broadcast Turbo Stream event to all active subscribers on "articles" topic
	ac.Broker.Publish(c.Request().Context(), "articles", fullstack.TurboStreamItem{
		Action:    fullstack.StreamPrepend,
		Target:    "articles-container",
		Component: components.ArticleCard(newArticle),
	})

	// 2. Return direct Turbo Stream response to creator
	return fullstack.RenderTurboStream(
		c,
		fullstack.StreamPrepend,
		"articles-container",
		components.ArticleCard(newArticle),
	)
}
